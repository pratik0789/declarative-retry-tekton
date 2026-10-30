package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
)

func PlanFiles(pipelinePath, runPath, policyPath string) (Plan, error) {
	return PlanFilesWithNewer(pipelinePath, runPath, policyPath, "")
}

func PlanFilesWithNewer(pipelinePath, runPath, policyPath, newerRunsPath string) (Plan, error) {
	return PlanFilesWithInputs(pipelinePath, runPath, policyPath, "", newerRunsPath, nil)
}

func PlanFilesWithInputs(pipelinePath, runPath, policyPath, taskRunsPath, newerRunsPath string, selected []string) (Plan, error) {
	var pipeline Pipeline
	var run PipelineRun
	var policy Policy
	for _, item := range []struct {
		path string
		out  any
	}{{pipelinePath, &pipeline}, {runPath, &run}, {policyPath, &policy}} {
		data, err := os.ReadFile(item.path)
		if err != nil {
			return Plan{}, err
		}
		if err := yaml.Unmarshal(data, item.out); err != nil {
			return Plan{}, fmt.Errorf("parse %s: %w", item.path, err)
		}
	}
	var newerRuns []PipelineRun
	if newerRunsPath != "" {
		data, err := os.ReadFile(newerRunsPath)
		if err != nil {
			return Plan{}, err
		}
		if err := yaml.Unmarshal(data, &newerRuns); err != nil {
			return Plan{}, fmt.Errorf("parse %s: %w", newerRunsPath, err)
		}
	}
	if taskRunsPath != "" {
		data, err := os.ReadFile(taskRunsPath)
		if err != nil {
			return Plan{}, err
		}
		var list TaskRunList
		if err := yaml.Unmarshal(data, &list); err != nil {
			return Plan{}, fmt.Errorf("parse %s: %w", taskRunsPath, err)
		}
		applyTaskRuns(&run, list.Items)
	}
	return BuildWithSelection(pipeline, run, policy, newerRuns, selected, time.Now())
}

func applyTaskRuns(run *PipelineRun, taskRuns []TaskRun) {
	taskByRunName := map[string]string{}
	for _, child := range run.Status.ChildReferences {
		taskByRunName[child.Name] = child.PipelineTask
	}
	for _, taskRun := range taskRuns {
		if source := taskRun.Metadata.Labels["tekton.dev/pipelineRun"]; source != "" && source != run.Metadata.Name {
			continue
		}
		task := taskRun.Metadata.Labels["tekton.dev/pipelineTask"]
		if task == "" {
			task = taskByRunName[taskRun.Metadata.Name]
		}
		if task == "" {
			continue
		}
		status := "Unknown"
		for _, condition := range taskRun.Status.Conditions {
			if condition.Type != "Succeeded" {
				continue
			}
			switch strings.ToLower(condition.Status) {
			case "true":
				status = "Succeeded"
			case "false":
				status = "Failed"
			}
		}
		run.Status.ChildReferences = append(run.Status.ChildReferences, ChildReference{Name: taskRun.Metadata.Name, PipelineTask: task, Status: status, Results: taskRun.Status.Results})
	}
}

func Build(p Pipeline, r PipelineRun, policy Policy, now time.Time) (Plan, error) {
	return BuildWithNewerRuns(p, r, policy, nil, now)
}

func BuildWithNewerRuns(p Pipeline, r PipelineRun, policy Policy, newerRuns []PipelineRun, now time.Time) (Plan, error) {
	return BuildWithSelection(p, r, policy, newerRuns, nil, now)
}

func BuildWithSelection(p Pipeline, r PipelineRun, policy Policy, newerRuns []PipelineRun, selected []string, now time.Time) (Plan, error) {
	plan := Plan{SourceRun: r.Metadata.Name, SourceRunUID: r.Metadata.UID, Pipeline: p.Metadata.Name, Policy: policy.Metadata.Name, PolicyGeneration: policy.Metadata.Generation, Decision: "recover"}
	refuse := func(reason string) (Plan, error) {
		plan.Decision = "refuse"
		plan.RefusalReason = reason
		return plan, nil
	}
	if p.Metadata.Name == "" || r.Metadata.Name == "" || policy.Metadata.Name == "" {
		return Plan{}, fmt.Errorf("resources must have metadata.name")
	}
	if r.Spec.PipelineRef.Name != p.Metadata.Name {
		return refuse("PipelineRun references a different Pipeline")
	}
	if policy.Spec.PipelineRef != p.Metadata.Name {
		return refuse("policy references a different Pipeline")
	}
	// Without a completion time, the creation time gives an earlier, more conservative window start.
	windowStart := r.Metadata.CreationTimestamp
	if r.Status.CompletionTime != nil {
		windowStart = *r.Status.CompletionTime
	}
	if policy.Spec.ResumeWithin != "" && !windowStart.IsZero() {
		d, err := time.ParseDuration(policy.Spec.ResumeWithin)
		if err != nil {
			return Plan{}, fmt.Errorf("invalid resumeWithin: %w", err)
		}
		if now.After(windowStart.Add(d)) {
			return refuse("recovery window has expired")
		}
	}
	if policy.Spec.RetryOnlyLatest {
		if newerRuns == nil {
			return refuse("retryOnlyLatest requires the PipelineRuns to compare against (--newer-runs)")
		}
		sourceValues := effectiveParameters(r)
		for _, name := range policy.Spec.TargetParameters {
			if value, ok := sourceValues[name]; !ok || value == nil {
				return refuse(fmt.Sprintf("cannot resolve target parameter %q for PipelineRun %q", name, r.Metadata.Name))
			}
		}
		for _, candidate := range newerRuns {
			if failedRecoveryOf(candidate, r) {
				continue
			}
			samePipeline := candidate.Spec.PipelineRef.Name == p.Metadata.Name || candidate.Metadata.Annotations[annotationSourcePipeline] == p.Metadata.Name
			if !samePipeline || !candidate.Metadata.CreationTimestamp.After(r.Metadata.CreationTimestamp) {
				continue
			}
			matches, err := sameTarget(policy.Spec.TargetParameters, sourceValues, candidate)
			if err != nil {
				return refuse(err.Error())
			}
			if matches {
				return refuse(fmt.Sprintf("newer PipelineRun %q supersedes the source run", candidate.Metadata.Name))
			}
		}
	}
	current, recorded := p.definition, r.recordedDefinition
	if current == nil {
		current = toAny(p.Spec)
	}
	if recorded == nil && r.Status.PipelineSpec != nil {
		recorded = toAny(r.Status.PipelineSpec)
	}
	if recorded == nil {
		return refuse("source PipelineRun does not record the Pipeline definition it ran")
	}
	if !definitionMatches(withoutNulls(current), withoutNulls(recorded)) {
		return refuse("Pipeline definition changed since the source run")
	}

	tasks := map[string]PipelineTask{}
	for _, task := range p.allTasks() {
		if _, exists := tasks[task.Name]; exists {
			return Plan{}, fmt.Errorf("duplicate pipeline task %q", task.Name)
		}
		tasks[task.Name] = task
	}
	ruleNames := make([]string, 0, len(policy.Spec.Tasks))
	for name := range policy.Spec.Tasks {
		ruleNames = append(ruleNames, name)
	}
	sort.Strings(ruleNames)
	for _, name := range ruleNames {
		for _, referenced := range append([]string{name}, policy.Spec.Tasks[name].RetryWith...) {
			if _, ok := tasks[referenced]; !ok {
				return refuse(fmt.Sprintf("policy references unknown task %q", referenced))
			}
		}
	}
	upstream, err := executionDependencies(current)
	if err != nil {
		return refuse(err.Error())
	}
	if err := validateRetryDirections(upstream, policy.Spec.Tasks); err != nil {
		return refuse(err.Error())
	}
	states := map[string]ChildReference{}
	failed := map[string]bool{}
	for _, child := range r.Status.ChildReferences {
		states[child.PipelineTask] = child
		if strings.EqualFold(child.Status, "Failed") {
			failed[child.PipelineTask] = true
		}
	}
	if len(failed) == 0 {
		return refuse("source PipelineRun has no failed tasks")
	}
	failedNames := make([]string, 0, len(failed))
	for name := range failed {
		failedNames = append(failedNames, name)
	}
	sort.Strings(failedNames)
	for _, name := range failedNames {
		if policy.Spec.Tasks[name].BlocksResume {
			return refuse(fmt.Sprintf("task %q failed and blocks resume", name))
		}
	}
	for _, name := range selected {
		if _, ok := tasks[name]; !ok {
			return refuse(fmt.Sprintf("selected unknown task %q", name))
		}
	}
	rerun := map[string]bool{}
	var visit func(string) error
	visit = func(name string) error {
		if _, ok := tasks[name]; !ok {
			return fmt.Errorf("policy references unknown task %q", name)
		}
		if rerun[name] {
			return nil
		}
		rerun[name] = true
		for _, dependency := range policy.Spec.Tasks[name].RetryWith {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		return nil
	}
	for _, name := range append(failedNames, selected...) {
		if err := visit(name); err != nil {
			return refuse(err.Error())
		}
	}
	names := make([]string, 0, len(tasks))
	for name := range tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		child, ran := states[name]
		taskPlan := TaskPlan{Name: name}
		switch {
		case rerun[name]:
			taskPlan.Action, taskPlan.Reason = "rerun", "failed or selected task, or retryWith closure"
		case !ran:
			taskPlan.Action, taskPlan.Reason = "continue", "task was not reached in source run"
		case strings.EqualFold(child.Status, "Succeeded"):
			taskPlan.Action, taskPlan.Reason, taskPlan.SourceTaskRun = "inherit", "successful result reused from source run", child.Name
			for _, result := range child.Results {
				if taskPlan.Results == nil {
					taskPlan.Results = map[string]any{}
				}
				taskPlan.Results[result.Name] = result.Value
			}
		default:
			taskPlan.Action, taskPlan.Reason = "rerun", "source task did not succeed"
		}
		plan.Tasks = append(plan.Tasks, taskPlan)
	}
	plan.Warnings = append(plan.Warnings, sharedResultWarnings(tasks, rerun, states)...)
	plan.Warnings = append(plan.Warnings, sharedWorkspaceWarnings(p, r, rerun, states)...)
	plan.Warnings = append(plan.Warnings, ephemeralWorkspaceWarnings(p, r, rerun, states)...)
	for name := range rerun {
		if !failed[name] && policy.Spec.Tasks[name].BlocksResume {
			plan.Warnings = append(plan.Warnings, Warning{
				Code:    "blocks-resume-task-rerun",
				Message: fmt.Sprintf("task %q declares blocksResume and will run again because it was selected or is in the retry closure", name),
				Tasks:   []string{name},
			})
		}
	}
	plan.Warnings = append(plan.Warnings, laterClaimWarnings(r, newerRuns)...)
	sort.Slice(plan.Warnings, func(i, j int) bool {
		left := plan.Warnings[i].Code + strings.Join(plan.Warnings[i].Tasks, "\x00") + plan.Warnings[i].StateName + plan.Warnings[i].Message
		right := plan.Warnings[j].Code + strings.Join(plan.Warnings[j].Tasks, "\x00") + plan.Warnings[j].StateName + plan.Warnings[j].Message
		return left < right
	})
	plan.Metrics.TotalTasks = len(plan.Tasks)
	for _, task := range plan.Tasks {
		switch task.Action {
		case "rerun":
			plan.Metrics.RerunTasks++
		case "inherit":
			plan.Metrics.InheritedTasks++
		case "continue":
			plan.Metrics.ContinuingTasks++
		}
	}
	plan.Metrics.TasksAvoided = plan.Metrics.InheritedTasks
	if plan.Metrics.TotalTasks > 0 {
		plan.Metrics.AvoidancePercent = float64(plan.Metrics.TasksAvoided) * 100 / float64(plan.Metrics.TotalTasks)
	}
	return plan, nil
}

// Compare the complete saved specification without resolving or wildcarding
// variables. A resolved value cannot establish which expression originally
// produced it, so ambiguous historical definitions are conservatively refused.
func definitionMatches(current, recorded any) bool {
	switch typed := current.(type) {
	case map[string]any:
		other, ok := recorded.(map[string]any)
		if !ok || len(typed) != len(other) {
			return false
		}
		for key, value := range typed {
			if otherValue, exists := other[key]; !exists || !definitionMatches(value, otherValue) {
				return false
			}
		}
		return true
	case []any:
		other, ok := recorded.([]any)
		if !ok || len(typed) != len(other) {
			return false
		}
		for index := range typed {
			if !definitionMatches(typed[index], other[index]) {
				return false
			}
		}
		return true
	case string:
		other, ok := recorded.(string)
		if !ok {
			return false
		}
		return typed == other
	default:
		return reflect.DeepEqual(current, recorded)
	}
}

func withoutNulls(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for key, item := range typed {
			if item != nil {
				out[key] = withoutNulls(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = withoutNulls(item)
		}
		return out
	}
	return value
}

func toAny(value any) any {
	encoded, _ := json.Marshal(value)
	var out any
	_ = json.Unmarshal(encoded, &out)
	return out
}

// A recovery run that itself failed must not stop the original run from being recovered again.
func failedRecoveryOf(candidate, source PipelineRun) bool {
	annotations := candidate.Metadata.Annotations
	ofSource := annotations[annotationSourceRun] == source.Metadata.Name
	if uid := annotations[annotationSourceRunUID]; uid != "" && source.Metadata.UID != "" {
		ofSource = uid == source.Metadata.UID
	}
	if !ofSource {
		return false
	}
	for _, condition := range candidate.Status.Conditions {
		if condition.Type == "Succeeded" && strings.EqualFold(condition.Status, "False") {
			return true
		}
	}
	return false
}

// Use the definition attached to each run, never today's Pipeline defaults:
// the referenced Pipeline may have changed since a candidate ran.
func effectiveParameters(run PipelineRun) map[string]any {
	definition := run.recordedDefinition
	if definition == nil && run.Status.PipelineSpec != nil {
		definition = toAny(run.Status.PipelineSpec)
	}
	if definition == nil {
		definition = run.inlineDefinition
	}
	values := map[string]any{}
	if spec, ok := definition.(map[string]any); ok {
		params, _ := spec["params"].([]any)
		for _, raw := range params {
			param, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			name, _ := param["name"].(string)
			if value, ok := param["default"]; ok && value != nil {
				values[name] = value
			}
		}
	}
	for _, param := range run.Spec.Params {
		// Object parameters retain default keys not supplied by the run.
		defaults, hasDefaults := values[param.Name].(map[string]any)
		overrides, isObject := param.Value.(map[string]any)
		if hasDefaults && isObject {
			merged := map[string]any{}
			for key, value := range defaults {
				merged[key] = value
			}
			for key, value := range overrides {
				merged[key] = value
			}
			values[param.Name] = merged
		} else {
			values[param.Name] = param.Value
		}
	}
	return values
}

func sameTarget(parameters []string, sourceValues map[string]any, candidate PipelineRun) (bool, error) {
	candidateValues := effectiveParameters(candidate)
	var unresolved error
	for _, name := range parameters {
		value, ok := candidateValues[name]
		if !ok || value == nil {
			if unresolved == nil {
				unresolved = fmt.Errorf("cannot resolve target parameter %q for PipelineRun %q", name, candidate.Metadata.Name)
			}
			continue
		}
		if !reflect.DeepEqual(sourceValues[name], value) {
			return false, nil
		}
	}
	if unresolved != nil {
		return false, unresolved
	}
	return true, nil
}

// Binding evidence is not proof of a write. Warn even for failed recoveries and
// other targets, whose effects are outside retryOnlyLatest's refusal rule.
func laterClaimWarnings(source PipelineRun, candidates []PipelineRun) []Warning {
	claims := map[string]string{}
	for _, binding := range source.Spec.Workspaces {
		pvc, ok := binding["persistentVolumeClaim"].(map[string]any)
		if !ok {
			continue
		}
		claim, _ := pvc["claimName"].(string)
		workspace, _ := binding["name"].(string)
		if claim != "" {
			claims[claim] = workspace
		}
	}
	var warnings []Warning
	seen := map[string]bool{}
	for _, candidate := range candidates {
		if !candidate.Metadata.CreationTimestamp.After(source.Metadata.CreationTimestamp) {
			continue
		}
		if source.Metadata.Namespace != "" && candidate.Metadata.Namespace != "" && source.Metadata.Namespace != candidate.Metadata.Namespace {
			continue
		}
		for _, binding := range candidate.Spec.Workspaces {
			pvc, ok := binding["persistentVolumeClaim"].(map[string]any)
			if !ok {
				continue
			}
			claim, _ := pvc["claimName"].(string)
			workspace, exists := claims[claim]
			key := candidate.Metadata.Name + "\x00" + claim
			if !exists || seen[key] {
				continue
			}
			seen[key] = true
			warnings = append(warnings, Warning{
				Code:      "later-run-shares-persistent-claim",
				Message:   fmt.Sprintf("later PipelineRun %q binds persistent claim %q used by source workspace %q; the claim may have changed, including after a failed recovery", candidate.Metadata.Name, claim, workspace),
				StateKind: "workspace",
				StateName: workspace,
			})
		}
	}
	return warnings
}

func ephemeralWorkspaceWarnings(p Pipeline, r PipelineRun, rerun map[string]bool, states map[string]ChildReference) []Warning {
	ephemeral := map[string]string{}
	for _, binding := range r.Spec.Workspaces {
		name, _ := binding["name"].(string)
		for _, kind := range []string{"volumeClaimTemplate", "emptyDir"} {
			if _, ok := binding[kind]; ok {
				ephemeral[name] = kind
			}
		}
	}
	inheritedUsers := map[string][]string{}
	remainingUsers := map[string][]string{}
	for _, task := range p.allTasks() {
		child, ran := states[task.Name]
		inherited := ran && strings.EqualFold(child.Status, "Succeeded") && !rerun[task.Name]
		for _, binding := range task.Workspaces {
			workspace := binding.Workspace
			if workspace == "" {
				workspace = binding.Name
			}
			if inherited {
				inheritedUsers[workspace] = append(inheritedUsers[workspace], task.Name)
			} else {
				remainingUsers[workspace] = append(remainingUsers[workspace], task.Name)
			}
		}
	}
	var warnings []Warning
	for workspace, kind := range ephemeral {
		inherited, remaining := inheritedUsers[workspace], remainingUsers[workspace]
		if len(inherited) == 0 || len(remaining) == 0 {
			continue
		}
		taskList := append(append([]string(nil), inherited...), remaining...)
		sort.Strings(taskList)
		warnings = append(warnings, Warning{
			Code:      "ephemeral-workspace-across-closure",
			Message:   fmt.Sprintf("workspace %q is bound by %s, so the recovery run starts with a new, empty volume; contents written by inherited tasks are not carried over", workspace, kind),
			Tasks:     taskList,
			StateKind: "workspace",
			StateName: workspace,
		})
	}
	return warnings
}

func sharedResultWarnings(tasks map[string]PipelineTask, rerun map[string]bool, states map[string]ChildReference) []Warning {
	var warnings []Warning
	for consumer, task := range tasks {
		child, ran := states[consumer]
		if !ran || !strings.EqualFold(child.Status, "Succeeded") || rerun[consumer] {
			continue
		}
		encoded, _ := json.Marshal(task.Params)
		for producer := range rerun {
			if strings.Contains(string(encoded), "$(tasks."+producer+".results.") {
				warnings = append(warnings, Warning{
					Code:      "shared-result-across-closure",
					Message:   fmt.Sprintf("task %q would be reused after consuming a result from rerun task %q", consumer, producer),
					Tasks:     []string{producer, consumer},
					StateKind: "result",
				})
			}
		}
	}
	return warnings
}

func sharedWorkspaceWarnings(p Pipeline, r PipelineRun, rerun map[string]bool, states map[string]ChildReference) []Warning {
	var warnings []Warning
	readOnly := map[string]bool{}
	for _, workspace := range p.Spec.Workspaces {
		readOnly[workspace.Name] = workspace.ReadOnly
	}
	// Kubernetes always mounts these volume sources read-only.
	for _, binding := range r.Spec.Workspaces {
		name, _ := binding["name"].(string)
		for _, kind := range []string{"secret", "configMap", "projected"} {
			if _, ok := binding[kind]; ok {
				readOnly[name] = true
			}
		}
	}
	users := map[string]map[string]bool{}
	for _, task := range p.allTasks() {
		for _, binding := range task.Workspaces {
			workspace := binding.Workspace
			if workspace == "" {
				workspace = binding.Name
			}
			if users[workspace] == nil {
				users[workspace] = map[string]bool{}
			}
			users[workspace][task.Name] = true
		}
	}
	for workspace, taskNames := range users {
		if readOnly[workspace] {
			continue
		}
		hasRerun, inherited := false, ""
		for name := range taskNames {
			if rerun[name] {
				hasRerun = true
			}
			if child, ok := states[name]; ok && strings.EqualFold(child.Status, "Succeeded") && !rerun[name] {
				inherited = name
			}
		}
		if hasRerun && inherited != "" {
			var taskList []string
			for name := range taskNames {
				if rerun[name] || (states[name].PipelineTask != "" && strings.EqualFold(states[name].Status, "Succeeded")) {
					taskList = append(taskList, name)
				}
			}
			sort.Strings(taskList)
			warnings = append(warnings, Warning{
				Code:      "shared-workspace-across-closure",
				Message:   fmt.Sprintf("writable workspace %q is shared by rerun and reused tasks", workspace),
				Tasks:     taskList,
				StateKind: "workspace",
				StateName: workspace,
			})
		}
	}
	return warnings
}
