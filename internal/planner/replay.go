package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"sigs.k8s.io/yaml"
)

type ReplayRow struct {
	FailedTask string   `json:"failedTask"`
	Decision   string   `json:"decision"`
	Reason     string   `json:"reason,omitempty"`
	Completed  int      `json:"completedBeforeFailure"`
	Rerun      []string `json:"rerun"`
	Inherited  int      `json:"inherited"`
	Continuing int      `json:"continuing"`
	Warnings   []string `json:"warnings,omitempty"`
}

type ReplayReport struct {
	Pipeline         string      `json:"pipeline"`
	Policy           string      `json:"policy"`
	WorkspaceBinding string      `json:"workspaceBinding"`
	Completed        string      `json:"completedModel"`
	Tasks            int         `json:"tasks"`
	Rows             []ReplayRow `json:"rows"`
}

// Replay injects a failure at each task in turn. With the "upstream" model only tasks upstream of
// the failure (through runAfter or result references) completed; with "independent", every task
// that does not depend on the failed task completed, including parallel siblings.
func Replay(pipelinePath, policyPath, binding, completedModel string, secretWorkspaces []string) (ReplayReport, error) {
	var pipeline Pipeline
	var policy Policy
	for path, out := range map[string]any{pipelinePath: &pipeline, policyPath: &policy} {
		data, err := os.ReadFile(path)
		if err != nil {
			return ReplayReport{}, err
		}
		if err := yaml.Unmarshal(data, out); err != nil {
			return ReplayReport{}, fmt.Errorf("parse %s: %w", path, err)
		}
	}
	if completedModel != "upstream" && completedModel != "independent" {
		return ReplayReport{}, fmt.Errorf("completed model must be upstream or independent")
	}
	if binding != "persistentVolumeClaim" && binding != "volumeClaimTemplate" {
		return ReplayReport{}, fmt.Errorf("binding must be persistentVolumeClaim or volumeClaimTemplate")
	}
	secret := map[string]bool{}
	for _, name := range secretWorkspaces {
		secret[name] = true
	}
	var workspaces []map[string]any
	for _, workspace := range pipeline.Spec.Workspaces {
		bound := map[string]any{"name": workspace.Name}
		switch {
		case secret[workspace.Name]:
			bound["secret"] = map[string]any{"secretName": workspace.Name}
		case binding == "persistentVolumeClaim":
			bound["persistentVolumeClaim"] = map[string]any{"claimName": workspace.Name}
		default:
			bound["volumeClaimTemplate"] = map[string]any{"spec": map[string]any{}}
		}
		workspaces = append(workspaces, bound)
	}

	upstream := map[string][]string{}
	for _, raw := range asList(asMap(pipeline.definition)["tasks"]) {
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
	ancestors := func(name string) map[string]bool {
		seen := map[string]bool{}
		var walk func(string)
		walk = func(current string) {
			for _, parent := range upstream[current] {
				if !seen[parent] {
					seen[parent] = true
					walk(parent)
				}
			}
		}
		walk(name)
		return seen
	}

	descendants := func(name string) map[string]bool {
		seen := map[string]bool{}
		changed := true
		for changed {
			changed = false
			for task, parents := range upstream {
				if seen[task] {
					continue
				}
				for _, parent := range parents {
					if parent == name || seen[parent] {
						seen[task] = true
						changed = true
						break
					}
				}
			}
		}
		return seen
	}

	report := ReplayReport{Pipeline: pipeline.Metadata.Name, Policy: policy.Metadata.Name, WorkspaceBinding: binding, Completed: completedModel, Tasks: len(pipeline.Spec.Tasks)}
	for _, failedTask := range pipeline.Spec.Tasks {
		var run PipelineRun
		run.Metadata.Name = "replay-" + failedTask.Name
		run.Spec.PipelineRef.Name = pipeline.Metadata.Name
		run.Spec.Workspaces = workspaces
		run.recordedDefinition = pipeline.definition
		completed := ancestors(failedTask.Name)
		if completedModel == "independent" {
			downstream := descendants(failedTask.Name)
			completed = map[string]bool{}
			for _, task := range pipeline.Spec.Tasks {
				if task.Name != failedTask.Name && !downstream[task.Name] {
					completed[task.Name] = true
				}
			}
		}
		for _, task := range pipeline.Spec.Tasks {
			if completed[task.Name] {
				run.Status.ChildReferences = append(run.Status.ChildReferences, ChildReference{Name: run.Metadata.Name + "-" + task.Name, PipelineTask: task.Name, Status: "Succeeded"})
			}
		}
		run.Status.ChildReferences = append(run.Status.ChildReferences, ChildReference{Name: run.Metadata.Name + "-" + failedTask.Name, PipelineTask: failedTask.Name, Status: "Failed"})
		plan, err := Build(pipeline, run, policy, time.Now())
		if err != nil {
			return ReplayReport{}, err
		}
		row := ReplayRow{FailedTask: failedTask.Name, Decision: plan.Decision, Reason: plan.RefusalReason, Completed: len(completed), Inherited: plan.Metrics.InheritedTasks, Continuing: plan.Metrics.ContinuingTasks}
		for _, task := range plan.Tasks {
			if task.Action == "rerun" {
				row.Rerun = append(row.Rerun, task.Name)
			}
		}
		codes := map[string]bool{}
		for _, warning := range plan.Warnings {
			codes[warning.Code] = true
		}
		for code := range codes {
			row.Warnings = append(row.Warnings, code)
		}
		sort.Strings(row.Warnings)
		report.Rows = append(report.Rows, row)
	}
	return report, nil
}
