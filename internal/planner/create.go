package planner

import (
	"encoding/base64"
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
	annotationPolicySpecSHA256 = "retry.tekton.dev/policy-spec-sha256"
	annotationPolicySpec       = "retry.tekton.dev/policy-spec"
	annotationInheritedRuns    = "retry.tekton.dev/inherited-taskruns"
	annotationWarningsAccepted = "retry.tekton.dev/shared-state-warnings-accepted"
	annotationWarningCodes     = "retry.tekton.dev/shared-state-warning-codes"
	annotationFinallyAnchor    = "retry.tekton.dev/finally-anchor"
	annotationResultPublishers = "retry.tekton.dev/result-publishers"
)

var (
	resultReference = regexp.MustCompile(`\$\((?:tasks|finally)\.([^.()]+)\.results\.([^.()\[\]]+)\)`)
	taskReference   = regexp.MustCompile(`\$\((?:tasks|finally)\.([^.()]+)\.`)
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
	upstream, err := executionDependencies(pipelineSpec)
	if err != nil {
		return nil, err
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
	if len(tasks) == 0 {
		delete(pipelineSpec, "tasks")
	} else {
		pipelineSpec["tasks"] = tasks
	}
	var finalTasks []any
	for _, raw := range asList(pipelineSpec["finally"]) {
		name, _ := asMap(raw)["name"].(string)
		if !remaining[name] {
			continue
		}
		prepared, err := prepare(raw)
		if err != nil {
			return nil, err
		}
		finalTasks = append(finalTasks, prepared)
	}
	if len(finalTasks) == 0 {
		delete(pipelineSpec, "finally")
	} else {
		pipelineSpec["finally"] = finalTasks
	}
	var publishers []any
	if results, ok := pipelineSpec["results"]; ok {
		holder := map[string]any{"results": results}
		publishers, err = publishInheritedPipelineResults(holder, inherited, upstream)
		if err != nil {
			return nil, fmt.Errorf("pipeline results: %w", err)
		}
		pipelineSpec["results"] = holder["results"]
		if len(publishers) > 0 {
			tasks = append(tasks, publishers...)
			pipelineSpec["tasks"] = tasks
		}
	}
	anchor := ""
	if len(tasks) == 0 && len(finalTasks) > 0 {
		// Tekton requires at least one ordinary task when finally is present.
		// This task is always skipped, so it creates no TaskRun or user work.
		anchor = "retryctl-finally-anchor"
		for suffix := 1; ; suffix++ {
			if _, exists := upstream[anchor]; !exists {
				break
			}
			anchor = fmt.Sprintf("retryctl-finally-anchor-%d", suffix)
		}
		pipelineSpec["tasks"] = []any{map[string]any{
			"name":     anchor,
			"when":     []any{map[string]any{"input": "false", "operator": "in", "values": []any{"true"}}},
			"taskSpec": map[string]any{"steps": []any{map[string]any{"name": "skipped", "image": "alpine:3.20", "script": "exit 0"}}},
		}}
	}
	spec := asMap(sourceRun["spec"])
	delete(spec, "pipelineRef")
	delete(spec, "status")
	spec["pipelineSpec"] = pipelineSpec
	if specs, ok := spec["taskRunSpecs"]; ok {
		var kept []any
		for _, raw := range asList(specs) {
			name, _ := asMap(raw)["pipelineTaskName"].(string)
			if remaining[name] {
				kept = append(kept, raw)
			}
		}
		spec["taskRunSpecs"] = kept
	}

	sourceMeta := asMap(sourceRun["metadata"])
	annotations := map[string]any{
		annotationSourceRun:        plan.SourceRun,
		annotationSourceRunUID:     plan.SourceRunUID,
		annotationSourcePipeline:   plan.Pipeline,
		annotationPolicy:           plan.Policy,
		annotationPolicySpecSHA256: plan.PolicySpecSHA256,
		annotationPolicySpec:       string(plan.PolicySpec),
	}
	if anchor != "" {
		annotations[annotationFinallyAnchor] = anchor
	}
	if len(publishers) > 0 {
		var names []string
		for _, publisher := range publishers {
			names = append(names, asMap(publisher)["name"].(string))
		}
		annotations[annotationResultPublishers] = strings.Join(names, ",")
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
	var substitutionErr error
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
					// Saved results are data. Pasting Tekton substitution syntax
					// into a task definition can create dependencies or rebind
					// that data to the recovery run's params/context. Refuse
					// until task-level literal transport is implemented.
					if strings.Contains(result, "$(") {
						if substitutionErr == nil {
							substitutionErr = fmt.Errorf("saved result %q of inherited task %q contains substitution syntax and cannot be inserted literally into a recovery task", parts[2], parts[1])
						}
						return match
					}
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
	if substitutionErr != nil {
		return substitutionErr
	}
	return rejectUnresolvedInheritedReferences(node, inherited)
}

// Tekton requires Pipeline outputs to reference task results, not literals.
// Publish saved strings without rerunning the original producing task. One
// publisher per producer avoids combining multiple source TaskRun result limits.
func publishInheritedPipelineResults(node map[string]any, inherited map[string]TaskPlan, originalTasks map[string][]string) ([]any, error) {
	var publishers []any
	byProducer := map[string]map[string]any{}
	references := map[string]map[string]string{}
	used := map[string]bool{}
	for name := range originalTasks {
		used[name] = true
	}
	var walk func(any) any
	walk = func(value any) any {
		switch typed := value.(type) {
		case string:
			return resultReference.ReplaceAllStringFunc(typed, func(match string) string {
				parts := resultReference.FindStringSubmatch(match)
				producer, resultName := parts[1], parts[2]
				task, ok := inherited[producer]
				if !ok {
					return match
				}
				value, ok := task.Results[resultName].(string)
				if !ok {
					return match
				}
				if reference := references[producer][resultName]; reference != "" {
					return reference
				}
				publisher := byProducer[producer]
				if publisher == nil {
					name := "retryctl-reused-results"
					for suffix := 1; used[name]; suffix++ {
						name = fmt.Sprintf("retryctl-reused-results-%d", suffix)
					}
					used[name] = true
					publisher = map[string]any{"name": name, "taskSpec": map[string]any{
						"results": []any{}, "steps": []any{map[string]any{"name": "publish", "image": "alpine:3.20", "script": "set -eu\n", "env": []any{}}},
					}}
					publishers = append(publishers, publisher)
					byProducer[producer] = publisher
					references[producer] = map[string]string{}
				}
				spec := asMap(publisher["taskSpec"])
				index := len(asList(spec["results"]))
				output := fmt.Sprintf("value-%d", index)
				env := fmt.Sprintf("RETRYCTL_VALUE_%d", index)
				spec["results"] = append(asList(spec["results"]), map[string]any{"name": output, "type": "string"})
				step := asMap(asList(spec["steps"])[0])
				step["env"] = append(asList(step["env"]), map[string]any{"name": env, "value": base64.StdEncoding.EncodeToString([]byte(value))})
				step["script"] = step["script"].(string) + fmt.Sprintf("printf '%%s' \"$%s\" | base64 -d > \"$(results.%s.path)\"\n", env, output)
				reference := fmt.Sprintf("$(tasks.%s.results.%s)", publisher["name"], output)
				references[producer][resultName] = reference
				return reference
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
	// Names and descriptions are metadata, not output expressions.
	for _, raw := range asList(node["results"]) {
		result := asMap(raw)
		result["value"] = walk(result["value"])
	}
	return publishers, rejectUnresolvedInheritedReferences(node, inherited)
}

func rejectUnresolvedInheritedReferences(node map[string]any, inherited map[string]TaskPlan) error {
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
