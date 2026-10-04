package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pratik0789/declarative-retry-tekton/internal/planner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: retryctl <plan|create|evaluate> [options]")
		os.Exit(2)
	}
	if os.Args[1] == "evaluate" {
		fs := flag.NewFlagSet("evaluate", flag.ExitOnError)
		cases := fs.String("cases", "evaluation/cases.yaml", "evaluation manifest")
		_ = fs.Parse(os.Args[2:])
		report, err := planner.Evaluate(*cases)
		writeJSON(report, err)
		return
	}
	if os.Args[1] == "create" {
		fs := flag.NewFlagSet("create", flag.ExitOnError)
		pipeline := fs.String("pipeline", "", "Tekton Pipeline YAML")
		run := fs.String("run", "", "failed PipelineRun YAML")
		policy := fs.String("policy", "", "PipelineRetryPolicy YAML")
		taskRuns := fs.String("taskruns", "", "Kubernetes List of source TaskRuns")
		confirmWarnings := fs.Bool("confirm-warnings", false, "continue when shared results or Workspaces cross the retry closure")
		_ = fs.Parse(os.Args[2:])
		plan, err := planner.PlanFilesWithInputs(*pipeline, *run, *policy, *taskRuns, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "retryctl:", err)
			os.Exit(1)
		}
		accepted := *confirmWarnings
		if len(plan.Warnings) > 0 {
			printWarnings(plan.Warnings)
			if !accepted {
				accepted, err = confirmInteractively()
				if err != nil {
					fmt.Fprintln(os.Stderr, "retryctl:", err)
					os.Exit(1)
				}
			}
		}
		manifest, err := planner.CreateRecoveryRun(*pipeline, *run, *policy, *taskRuns, accepted)
		if err != nil {
			fmt.Fprintln(os.Stderr, "retryctl:", err)
			os.Exit(1)
		}
		fmt.Print(string(manifest))
		return
	}
	if os.Args[1] != "plan" {
		fmt.Fprintln(os.Stderr, "usage: retryctl <plan|create|evaluate> [options]")
		os.Exit(2)
	}
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	pipeline := fs.String("pipeline", "", "Tekton Pipeline YAML")
	run := fs.String("run", "", "failed PipelineRun YAML")
	policy := fs.String("policy", "", "PipelineRetryPolicy YAML")
	taskRuns := fs.String("taskruns", "", "optional Kubernetes List of source TaskRuns")
	newerRuns := fs.String("newer-runs", "", "optional YAML list of newer PipelineRuns")
	_ = fs.Parse(os.Args[2:])
	if *pipeline == "" || *run == "" || *policy == "" {
		fs.Usage()
		os.Exit(2)
	}
	plan, err := planner.PlanFilesWithInputs(*pipeline, *run, *policy, *taskRuns, *newerRuns)
	if err != nil {
		fmt.Fprintln(os.Stderr, "retryctl:", err)
		os.Exit(1)
	}
	writeJSON(plan, err)
}

func printWarnings(warnings []planner.Warning) {
	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "warning [%s]: %s\n", warning.Code, warning.Message)
	}
}

func confirmInteractively() (bool, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return false, fmt.Errorf("shared-state warnings require interactive confirmation or --confirm-warnings")
	}
	fmt.Fprint(os.Stderr, "Continue with the author-declared retry closure? [y/N] ")
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("confirmation not received; rerun interactively or use --confirm-warnings")
	}
	if strings.EqualFold(strings.TrimSpace(answer), "y") || strings.EqualFold(strings.TrimSpace(answer), "yes") {
		return true, nil
	}
	return false, fmt.Errorf("recovery cancelled by operator")
}

func writeJSON(value any, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "retryctl:", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(value, "", "  ")
	fmt.Println(string(out))
}
