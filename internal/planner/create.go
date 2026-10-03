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
		PipelineSpec struct {
			Tasks []PipelineTask `json:"tasks"`
		} `json:"pipelineSpec"`
	} `json:"spec"`
}

func CreateRecoveryRun(pipelinePath, runPath, policyPath, taskRunsPath string) ([]byte, error) {
	if pipelinePath == "" || runPath == "" || policyPath == "" || taskRunsPath == "" {
		return nil, fmt.Errorf("--pipeline, --run, --policy, and --taskruns are required")
	}
	plan, err := PlanFilesWithInputs(pipelinePath, runPath, policyPath, taskRunsPath, "")
	if err != nil {
		return nil, err
	}
	if plan.Decision != "recover" {
		return nil, fmt.Errorf("recovery refused: %s", plan.RefusalReason)
	}
	var pipeline Pipeline
	data, err := os.ReadFile(pipelinePath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
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
			if strings.Contains(encoded, "$(tasks.") {
				return nil, fmt.Errorf("task %q contains a result reference requiring inherited-result substitution", task.Name)
			}
		}
		output.Spec.PipelineSpec.Tasks = append(output.Spec.PipelineSpec.Tasks, task)
	}
	return yaml.Marshal(output)
}
