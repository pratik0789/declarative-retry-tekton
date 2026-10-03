package planner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

type EvaluationManifest struct {
	Cases []EvaluationCase `json:"cases"`
}

type EvaluationCase struct {
	Name                   string            `json:"name"`
	Pipeline               string            `json:"pipeline"`
	Run                    string            `json:"run"`
	Policy                 string            `json:"policy"`
	TaskRuns               string            `json:"taskRuns"`
	NewerRuns              string            `json:"newerRuns"`
	ExpectedDecision       string            `json:"expectedDecision"`
	ExpectedActions        map[string]string `json:"expectedActions"`
	ExpectedReasonContains string            `json:"expectedReasonContains"`
}

type EvaluationReport struct {
	TotalCases           int                    `json:"totalCases"`
	PassedCases          int                    `json:"passedCases"`
	CorrectnessPercent   float64                `json:"correctnessPercent"`
	RecoverableCases     int                    `json:"recoverableCases"`
	TotalTasksAvoided    int                    `json:"totalTasksAvoided"`
	MeanAvoidancePercent float64                `json:"meanAvoidancePercent"`
	Cases                []EvaluationCaseResult `json:"cases"`
}

type EvaluationCaseResult struct {
	Name          string  `json:"name"`
	Passed        bool    `json:"passed"`
	Decision      string  `json:"decision,omitempty"`
	Reason        string  `json:"reason,omitempty"`
	TasksAvoided  int     `json:"tasksAvoided,omitempty"`
	AvoidanceRate float64 `json:"avoidancePercent,omitempty"`
	Error         string  `json:"error,omitempty"`
}

func Evaluate(manifestPath string) (EvaluationReport, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return EvaluationReport{}, err
	}
	var manifest EvaluationManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return EvaluationReport{}, fmt.Errorf("parse %s: %w", manifestPath, err)
	}
	base := filepath.Dir(manifestPath)
	report := EvaluationReport{TotalCases: len(manifest.Cases)}
	for _, item := range manifest.Cases {
		resolve := func(path string) string {
			if path == "" || filepath.IsAbs(path) {
				return path
			}
			return filepath.Join(base, path)
		}
		plan, planErr := PlanFilesWithInputs(resolve(item.Pipeline), resolve(item.Run), resolve(item.Policy), resolve(item.TaskRuns), resolve(item.NewerRuns))
		result := EvaluationCaseResult{Name: item.Name}
		if planErr != nil {
			result.Error = planErr.Error()
		} else {
			result.Decision, result.Reason = plan.Decision, plan.RefusalReason
			result.TasksAvoided, result.AvoidanceRate = plan.Metrics.TasksAvoided, plan.Metrics.AvoidancePercent
			result.Passed = plan.Decision == item.ExpectedDecision
			if item.ExpectedReasonContains != "" {
				result.Passed = result.Passed && strings.Contains(plan.RefusalReason, item.ExpectedReasonContains)
			}
			actions := map[string]string{}
			for _, task := range plan.Tasks {
				actions[task.Name] = task.Action
			}
			for name, expected := range item.ExpectedActions {
				result.Passed = result.Passed && actions[name] == expected
			}
			if plan.Decision == "recover" {
				report.RecoverableCases++
				report.TotalTasksAvoided += plan.Metrics.TasksAvoided
				report.MeanAvoidancePercent += plan.Metrics.AvoidancePercent
			}
		}
		if result.Passed {
			report.PassedCases++
		}
		report.Cases = append(report.Cases, result)
	}
	if report.TotalCases > 0 {
		report.CorrectnessPercent = float64(report.PassedCases) * 100 / float64(report.TotalCases)
	}
	if report.RecoverableCases > 0 {
		report.MeanAvoidancePercent /= float64(report.RecoverableCases)
	}
	return report, nil
}
