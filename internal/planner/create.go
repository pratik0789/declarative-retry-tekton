package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"
)

const (
	annotationSourceRun        = "retry.tekton.dev/source-pipeline-run"
	annotationSourceRunUID     = "retry.tekton.dev/source-pipeline-run-uid"
	annotationSourcePipeline   = "retry.tekton.dev/source-pipeline"
	annotationPolicy           = "retry.tekton.dev/policy"
	annotationPolicyGeneration = "retry.tekton.dev/policy-generation"
	annotationInheritedRuns    = "retry.tekton.dev/inherited-taskruns"
	annotationWarningsAccepted = "retry.tekton.dev/shared-state-warnings-accepted"
	annotationWarningCodes     = "retry.tekton.dev/shared-state-warning-codes"
)

var (
	resultReference = regexp.MustCompile(`\$\(tasks\.([^.()]+)\.results\.([^.()\[\]]+)\)`)
	taskReference   = regexp.MustCompile(`\$\(tasks\.([^.()]+)\.`)
)

// Operates on raw manifests so Tekton fields the planner does not model survive unchanged.
func CreateRecoveryRun(pipelinePath, runPath, policyPath, taskRunsPath, newerRunsPath string, selected []string, acceptWarnings bool) ([]byte, error) {
	if pipelinePath == "" || runPath == "" || policyPath == "" {
		return nil, fmt.Errorf("--pipeline, --run, and --policy are required")
	}
	plan, err := PlanFilesWithInputs(pipelinePath, runPath, policyPath, taskRunsPath, newerRunsPath, selected)
	if err != nil {
		return nil, err
	}
	if plan.Decision != "recover" {
		return nil, fmt.Errorf("recovery refused: %s", plan.RefusalReason)
	}
	if len(plan.Warnings) > 0 && !acceptWarnings {
		return nil, fmt.Errorf("recovery has %d warning(s); explicit confirmation is required", len(plan.Warnings))
	}
	pipeline, err := readRaw(pipelinePath)
	if err != nil {
		return nil, err
	}
	sourceRun, err := readRaw(runPath)
	if err != nil {
		return nil, err
	}

	remaining := map[string]bool{}
	inherited := map[string]TaskPlan{}
	for _, task := range plan.Tasks {
		if task.Action == "inherit" {
			inherited[task.Name] = task
		} else {
			remaining[task.Name] = true
		}
	}
	prepare := func(raw any) (map[string]any, error) {
		task, _ := raw.(map[string]any)
		name, _ := task["name"].(string)
		if err := substituteResults(task, inherited); err != nil {
			return nil, fmt.Errorf("task %q: %w", name, err)
		}
		return task, nil
	}

	pipelineSpec := asMap(pipeline["spec"])
	upstream := map[string][]string{}
	for _, raw := range asList(pipelineSpec["tasks"]) {
		task := asMap(raw)
		name, _ := task["name"].(string)
		for _, dependency := range asList(task["runAfter"]) {
			upstream[name] = append(upstream[name], fmt.Sprint(dependency))
		}
		encoded, _ := json.Marshal(task)
		for _, match := range taskReference.FindAllStringSubmatch(string(encoded), -1) {
			upstream[name] = append(upstream[name], match[1])
		}
	}
	// An edge to an inherited task still orders against whatever reruns above it.
	var nearestRemaining func(string, map[string]bool) []string
	nearestRemaining = func(task string, seen map[string]bool) []string {
		var found []string
		for _, parent := range upstream[task] {
			if seen[parent] {
				continue
			}
			seen[parent] = true
			if remaining[parent] {
				found = append(found, parent)
			} else {
				found = append(found, nearestRemaining(parent, seen)...)
			}
		}
		return found
	}
	var tasks []any
	for _, raw := range asList(pipelineSpec["tasks"]) {
		task := asMap(raw)
		name, _ := task["name"].(string)
		if !remaining[name] {
			continue
		}
		var runAfter []any
		ordered := map[string]bool{}
		for _, dependency := range asList(task["runAfter"]) {
			if remaining[fmt.Sprint(dependency)] {
				runAfter = append(runAfter, dependency)
				ordered[fmt.Sprint(dependency)] = true
			}
		}
		for _, dependency := range upstream[name] {
			if remaining[dependency] {
				continue
			}
			for _, ancestor := range nearestRemaining(dependency, map[string]bool{dependency: true}) {
				if !ordered[ancestor] && ancestor != name {
					runAfter = append(runAfter, ancestor)
					ordered[ancestor] = true
				}
			}
		}
		if len(runAfter) > 0 {
			task["runAfter"] = runAfter
		} else {
			delete(task, "runAfter")
		}
		prepared, err := prepare(task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, prepared)
	}
	pipelineSpec["tasks"] = tasks
	finallyNames := map[string]bool{}
	for _, raw := range asList(pipelineSpec["finally"]) {
		prepared, err := prepare(raw)
		if err != nil {
			return nil, err
		}
		name, _ := prepared["name"].(string)
		finallyNames[name] = true
	}
	if results, ok := pipelineSpec["results"]; ok {
		holder := map[string]any{"results": results}
		if err := substituteResults(holder, inherited); err != nil {
			return nil, fmt.Errorf("pipeline results: %w", err)
		}
		pipelineSpec["results"] = holder["results"]
	}

	spec := asMap(sourceRun["spec"])
	delete(spec, "pipelineRef")
	delete(spec, "status")
	spec["pipelineSpec"] = pipelineSpec
	if specs, ok := spec["taskRunSpecs"]; ok {
		var kept []any
		for _, raw := range asList(specs) {
			name, _ := asMap(raw)["pipelineTaskName"].(string)
			if remaining[name] || finallyNames[name] {
				kept = append(kept, raw)
			}
		}
		spec["taskRunSpecs"] = kept
	}

	sourceMeta := asMap(sourceRun["metadata"])
	annotations := map[string]any{
		annotationSourceRun:      plan.SourceRun,
		annotationSourceRunUID:   plan.SourceRunUID,
		annotationSourcePipeline: plan.Pipeline,
		annotationPolicy:         plan.Policy,
	}
	if plan.PolicyGeneration != 0 {
		annotations[annotationPolicyGeneration] = strconv.FormatInt(plan.PolicyGeneration, 10)
	}
	var inheritedRuns []string
	for name, task := range inherited {
		inheritedRuns = append(inheritedRuns, name+"="+task.SourceTaskRun)
	}
	sort.Strings(inheritedRuns)
	if len(inheritedRuns) > 0 {
		annotations[annotationInheritedRuns] = strings.Join(inheritedRuns, ",")
	}
	if len(plan.Warnings) > 0 {
		codes := make([]string, 0, len(plan.Warnings))
		for _, warning := range plan.Warnings {
			codes = append(codes, warning.Code)
		}
		annotations[annotationWarningsAccepted] = "true"
		annotations[annotationWarningCodes] = strings.Join(codes, ",")
	}
	metadata := map[string]any{
		"generateName": plan.Pipeline + "-retry-",
		"annotations":  annotations,
	}
	if namespace, ok := sourceMeta["namespace"]; ok {
		metadata["namespace"] = namespace
	}
	return yaml.Marshal(map[string]any{
		"apiVersion": "tekton.dev/v1",
		"kind":       "PipelineRun",
		"metadata":   metadata,
		"spec":       spec,
	})
}

func substituteResults(node map[string]any, inherited map[string]TaskPlan) error {
	var walk func(any) any
	walk = func(value any) any {
		switch typed := value.(type) {
		case string:
			return resultReference.ReplaceAllStringFunc(typed, func(match string) string {
				parts := resultReference.FindStringSubmatch(match)
				task, ok := inherited[parts[1]]
				if !ok {
					return match
				}
				if result, ok := task.Results[parts[2]].(string); ok {
					return result
				}
				return match
			})
		case map[string]any:
			for key, item := range typed {
				typed[key] = walk(item)
			}
		case []any:
			for index, item := range typed {
				typed[index] = walk(item)
			}
		}
		return value
	}
	walk(node)
	encoded, err := json.Marshal(node)
	if err != nil {
		return err
	}
	for _, match := range taskReference.FindAllStringSubmatch(string(encoded), -1) {
		if _, ok := inherited[match[1]]; ok {
			return fmt.Errorf("reference to inherited task %q cannot be resolved from the source TaskRun (missing or non-string result, or status reference)", match[1])
		}
	}
	return nil
}

func readRaw(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}

func asMap(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func asList(value any) []any {
	typed, _ := value.([]any)
	return typed
}
