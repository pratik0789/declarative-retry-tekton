package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/pratik0789/declarative-retry-tekton/internal/planner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: retryctl <plan|evaluate> [options]")
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
		_ = fs.Parse(os.Args[2:])
		manifest, err := planner.CreateRecoveryRun(*pipeline, *run, *policy, *taskRuns)
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

func writeJSON(value any, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "retryctl:", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(value, "", "  ")
	fmt.Println(string(out))
}
