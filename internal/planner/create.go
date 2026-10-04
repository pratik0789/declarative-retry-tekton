package planner

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

type recoveryRun struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		GenerateName string            `json:"generateName"`
		Annotations  map[string]string `json:"annotations"`
	} `json:"metadata"`
	Spec struct {
		Workspaces   []map[string]any `json:"workspaces,omitempty"`
		PipelineSpec struct {
			Tasks      []PipelineTask `json:"tasks"`
			Workspaces []Workspace    `json:"workspaces,omitempty"`
		} `json:"pipelineSpec"`
	} `json:"spec"`
}

func CreateRecoveryRun(pipelinePath, runPath, policyPath, taskRunsPath string, acceptWarnings bool) ([]byte, error) {
	if pipelinePath == "" || runPath == "" || policyPath == "" {
		return nil, fmt.Errorf("--pipeline, --run, and --policy are required")
	}
	plan, err := PlanFilesWithInputs(pipelinePath, runPath, policyPath, taskRunsPath, "")
	if err != nil {
		return nil, err
	}
	if plan.Decision != "recover" {
		return nil, fmt.Errorf("recovery refused: %s", plan.RefusalReason)
	}
	if len(plan.Warnings) > 0 && !acceptWarnings {
		return nil, fmt.Errorf("recovery has %d warning(s); explicit confirmation is required", len(plan.Warnings))
	}
	warningsAccepted := len(plan.Warnings) > 0 && acceptWarnings
	var pipeline Pipeline
	data, err := os.ReadFile(pipelinePath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return nil, err
	}
	var sourceRun PipelineRun
	runData, err := os.ReadFile(runPath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(runData, &sourceRun); err != nil {
		return nil, err
	}
	actions := map[string]string{}
	for _, task := range plan.Tasks {
		actions[task.Name] = task.Action
	}
	var output recoveryRun
	output.APIVersion, output.Kind = "tekton.dev/v1", "PipelineRun"
	output.Metadata.GenerateName = pipeline.Metadata.Name + "-retry-"
	output.Metadata.Annotations = map[string]string{
		"retry.tekton.dev/source-pipeline-run":     plan.SourceRun,
		"retry.tekton.dev/source-pipeline-run-uid": plan.SourceRunUID,
		"retry.tekton.dev/policy":                  plan.Policy,
	}
	if warningsAccepted {
		output.Metadata.Annotations["retry.tekton.dev/shared-state-warnings-accepted"] = "true"
		codes := make([]string, 0, len(plan.Warnings))
		for _, warning := range plan.Warnings {
			codes = append(codes, warning.Code)
		}
		output.Metadata.Annotations["retry.tekton.dev/shared-state-warning-codes"] = strings.Join(codes, ",")
	}
	output.Spec.Workspaces = sourceRun.Spec.Workspaces
	output.Spec.PipelineSpec.Workspaces = pipeline.Spec.Workspaces
	for _, original := range pipeline.Spec.Tasks {
		if actions[original.Name] != "rerun" && actions[original.Name] != "continue" {
			continue
		}
		task := original
		task.RunAfter = task.RunAfter[:0]
		for _, dependency := range original.RunAfter {
			if actions[dependency] == "rerun" || actions[dependency] == "continue" {
				task.RunAfter = append(task.RunAfter, dependency)
			}
		}
		for _, param := range task.Params {
			encoded := fmt.Sprint(param.Value)
			for producer, action := range actions {
				if strings.Contains(encoded, "$(tasks."+producer+".results.") && action == "inherit" {
					return nil, fmt.Errorf("task %q contains a result reference requiring inherited-result substitution", task.Name)
				}
			}
		}
		output.Spec.PipelineSpec.Tasks = append(output.Spec.PipelineSpec.Tasks, task)
	}
	return yaml.Marshal(output)
}
