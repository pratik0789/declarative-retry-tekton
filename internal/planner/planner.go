package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
)

func PlanFiles(pipelinePath, runPath, policyPath string) (Plan, error) {
	return PlanFilesWithNewer(pipelinePath, runPath, policyPath, "")
}

func PlanFilesWithNewer(pipelinePath, runPath, policyPath, newerRunsPath string) (Plan, error) {
	return PlanFilesWithInputs(pipelinePath, runPath, policyPath, "", newerRunsPath)
}

func PlanFilesWithInputs(pipelinePath, runPath, policyPath, taskRunsPath, newerRunsPath string) (Plan, error) {
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
	return BuildWithNewerRuns(pipeline, run, policy, newerRuns, time.Now())
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
		run.Status.ChildReferences = append(run.Status.ChildReferences, ChildReference{Name: taskRun.Metadata.Name, PipelineTask: task, Status: status})
	}
}

func Build(p Pipeline, r PipelineRun, policy Policy, now time.Time) (Plan, error) {
	return BuildWithNewerRuns(p, r, policy, nil, now)
}

func BuildWithNewerRuns(p Pipeline, r PipelineRun, policy Policy, newerRuns []PipelineRun, now time.Time) (Plan, error) {
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
	if policy.Spec.ResumeWithin != "" && !r.Metadata.CreationTimestamp.IsZero() {
		d, err := time.ParseDuration(policy.Spec.ResumeWithin)
		if err != nil {
			return Plan{}, fmt.Errorf("invalid resumeWithin: %w", err)
		}
		if now.After(r.Metadata.CreationTimestamp.Add(d)) {
			return refuse("recovery window has expired")
		}
	}
	if policy.Spec.RetryOnlyLatest {
		for _, candidate := range newerRuns {
			if candidate.Spec.PipelineRef.Name == p.Metadata.Name && candidate.Metadata.CreationTimestamp.After(r.Metadata.CreationTimestamp) {
				return refuse(fmt.Sprintf("newer PipelineRun %q supersedes the source run", candidate.Metadata.Name))
			}
		}
	}
	if r.Status.PipelineSpec != nil {
		current, _ := json.Marshal(struct {
			Tasks      []PipelineTask `json:"tasks"`
			Workspaces []Workspace    `json:"workspaces"`
		}{p.Spec.Tasks, p.Spec.Workspaces})
		recorded, _ := json.Marshal(r.Status.PipelineSpec)
		if string(current) != string(recorded) {
			return refuse("Pipeline specification changed since the source run")
		}
	}

	tasks := map[string]PipelineTask{}
	for _, task := range p.Spec.Tasks {
		if _, exists := tasks[task.Name]; exists {
			return Plan{}, fmt.Errorf("duplicate pipeline task %q", task.Name)
		}
		tasks[task.Name] = task
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
	rerun := map[string]bool{}
	var visit func(string) error
	visiting := map[string]bool{}
	visit = func(name string) error {
		if _, ok := tasks[name]; !ok {
			return fmt.Errorf("policy references unknown task %q", name)
		}
		if visiting[name] {
			return fmt.Errorf("cycle in retryWith at task %q", name)
		}
		if rerun[name] {
			return nil
		}
		visiting[name] = true
		rule := policy.Spec.Tasks[name]
		if rule.BlocksResume {
			return fmt.Errorf("task %q blocks resume", name)
		}
		for _, dependency := range rule.RetryWith {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[name] = false
		rerun[name] = true
		return nil
	}
	for name := range failed {
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
			taskPlan.Action, taskPlan.Reason = "rerun", "failed task or retryWith closure"
		case !ran:
			taskPlan.Action, taskPlan.Reason = "continue", "task was not reached in source run"
		case strings.EqualFold(child.Status, "Succeeded"):
			taskPlan.Action, taskPlan.Reason, taskPlan.SourceTaskRun = "inherit", "successful result reused from source run", child.Name
		default:
			taskPlan.Action, taskPlan.Reason = "rerun", "source task did not succeed"
		}
		plan.Tasks = append(plan.Tasks, taskPlan)
	}
	if reason := unsafeResultReuse(tasks, rerun, states); reason != "" {
		return refuse(reason)
	}
	if reason := unsafeSharedWorkspace(p, rerun, states); reason != "" {
		return refuse(reason)
	}
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

func unsafeResultReuse(tasks map[string]PipelineTask, rerun map[string]bool, states map[string]ChildReference) string {
	for consumer, task := range tasks {
		child, ran := states[consumer]
		if !ran || !strings.EqualFold(child.Status, "Succeeded") || rerun[consumer] {
			continue
		}
		encoded, _ := json.Marshal(task.Params)
		for producer := range rerun {
			if strings.Contains(string(encoded), "$(tasks."+producer+".results.") {
				return fmt.Sprintf("task %q would inherit results produced by rerun task %q", consumer, producer)
			}
		}
	}
	return ""
}

func unsafeSharedWorkspace(p Pipeline, rerun map[string]bool, states map[string]ChildReference) string {
	readOnly := map[string]bool{}
	for _, workspace := range p.Spec.Workspaces {
		readOnly[workspace.Name] = workspace.ReadOnly
	}
	users := map[string]map[string]bool{}
	for _, task := range p.Spec.Tasks {
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
			return fmt.Sprintf("mutable workspace %q is shared by rerun and inherited tasks (including %q)", workspace, inherited)
		}
	}
	return ""
}
