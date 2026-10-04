package planner

import (
	"strings"
	"testing"
	"time"
)

func TestBuildClassifiesTasks(t *testing.T) {
	var p Pipeline
	p.Metadata.Name = "delivery"
	p.Spec.Tasks = []PipelineTask{{Name: "clone"}, {Name: "build"}, {Name: "scan"}, {Name: "deploy"}}
	var r PipelineRun
	r.Metadata.Name, r.Metadata.UID = "delivery-1", "run-uid"
	r.Spec.PipelineRef.Name = "delivery"
	r.Status.ChildReferences = []ChildReference{{Name: "clone-x", PipelineTask: "clone", Status: "Succeeded"}, {Name: "build-x", PipelineTask: "build", Status: "Succeeded"}, {Name: "scan-x", PipelineTask: "scan", Status: "Failed"}}
	var policy Policy
	policy.Metadata.Name = "safe-retry"
	policy.Spec.PipelineRef = "delivery"
	policy.Spec.Tasks = map[string]TaskRule{"scan": {RetryWith: []string{"build"}}}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"build": "rerun", "clone": "inherit", "deploy": "continue", "scan": "rerun"}
	for _, task := range plan.Tasks {
		if task.Action != want[task.Name] {
			t.Errorf("%s: got %s want %s", task.Name, task.Action, want[task.Name])
		}
	}
	if plan.Metrics.TasksAvoided != 1 || plan.Metrics.AvoidancePercent != 25 {
		t.Fatalf("unexpected metrics: %+v", plan.Metrics)
	}
}

func TestBuildWarnsForInheritedConsumerOfRerunResult(t *testing.T) {
	p, r, policy := safetyFixture()
	p.Spec.Tasks[1].Params = []Param{{Name: "image", Value: "$(tasks.build.results.image)"}}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Decision != "recover" || !warningContains(plan.Warnings, "shared-result-across-closure") {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestBuildWarnsForMutableWorkspaceAcrossBoundary(t *testing.T) {
	p, r, policy := safetyFixture()
	p.Spec.Workspaces = []Workspace{{Name: "source"}}
	p.Spec.Tasks[0].Workspaces = []WorkspaceBinding{{Name: "src", Workspace: "source"}}
	p.Spec.Tasks[1].Workspaces = []WorkspaceBinding{{Name: "src", Workspace: "source"}}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Decision != "recover" || !warningContains(plan.Warnings, "shared-workspace-across-closure") {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestBuildRefusesChangedPipeline(t *testing.T) {
	p, r, policy := safetyFixture()
	r.Status.PipelineSpec = &struct {
		Tasks      []PipelineTask `json:"tasks"`
		Workspaces []Workspace    `json:"workspaces"`
	}{Tasks: []PipelineTask{{Name: "old-build"}}}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Decision != "refuse" || !contains(plan.RefusalReason, "changed") {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestBuildRefusesSupersededRun(t *testing.T) {
	p, r, policy := safetyFixture()
	policy.Spec.RetryOnlyLatest = true
	r.Metadata.CreationTimestamp = time.Now().Add(-time.Hour)
	var newer PipelineRun
	newer.Metadata.Name = "newer"
	newer.Metadata.CreationTimestamp = time.Now()
	newer.Spec.PipelineRef.Name = "p"
	plan, err := BuildWithNewerRuns(p, recorded(r, p), policy, []PipelineRun{newer}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Decision != "refuse" || !contains(plan.RefusalReason, "supersedes") {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestBuildRefusesRetryOnlyLatestWithoutComparisonRuns(t *testing.T) {
	p, r, policy := safetyFixture()
	policy.Spec.RetryOnlyLatest = true
	plan, err := BuildWithNewerRuns(p, recorded(r, p), policy, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Decision != "refuse" || !contains(plan.RefusalReason, "--newer-runs") {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestBuildRetryOnlyLatestHonorsTargetParameters(t *testing.T) {
	p, r, policy := safetyFixture()
	policy.Spec.RetryOnlyLatest = true
	policy.Spec.TargetParameters = []string{"branch"}
	r.Metadata.CreationTimestamp = time.Now().Add(-time.Hour)
	r.Spec.Params = []Param{{Name: "branch", Value: "main"}}
	newer := func(branch string) PipelineRun {
		var run PipelineRun
		run.Metadata.Name = "newer-" + branch
		run.Metadata.CreationTimestamp = time.Now()
		run.Spec.PipelineRef.Name = "p"
		run.Spec.Params = []Param{{Name: "branch", Value: branch}}
		return run
	}
	plan, err := BuildWithNewerRuns(p, recorded(r, p), policy, []PipelineRun{newer("feature")}, time.Now())
	if err != nil || plan.Decision != "recover" {
		t.Fatalf("other branch must not supersede: %+v %v", plan, err)
	}
	plan, err = BuildWithNewerRuns(p, recorded(r, p), policy, []PipelineRun{newer("main")}, time.Now())
	if err != nil || plan.Decision != "refuse" {
		t.Fatalf("same branch must supersede: %+v %v", plan, err)
	}
}

func TestBuildRetryOnlyLatestCountsRecoveryRuns(t *testing.T) {
	p, r, policy := safetyFixture()
	policy.Spec.RetryOnlyLatest = true
	r.Metadata.CreationTimestamp = time.Now().Add(-time.Hour)
	var recovery PipelineRun
	recovery.Metadata.Name = "p-retry-x"
	recovery.Metadata.CreationTimestamp = time.Now()
	recovery.Metadata.Annotations = map[string]string{annotationSourcePipeline: "p"}
	plan, err := BuildWithNewerRuns(p, recorded(r, p), policy, []PipelineRun{recovery}, time.Now())
	if err != nil || plan.Decision != "refuse" {
		t.Fatalf("recovery run with embedded spec must supersede: %+v %v", plan, err)
	}
}

func TestBuildResumeWithinStartsAtCompletion(t *testing.T) {
	p, r, policy := safetyFixture()
	policy.Spec.ResumeWithin = "4h"
	now := time.Now()
	r.Metadata.CreationTimestamp = now.Add(-5 * time.Hour)
	completed := now.Add(-time.Hour)
	r.Status.CompletionTime = &completed
	plan, err := Build(p, recorded(r, p), policy, now)
	if err != nil || plan.Decision != "recover" {
		t.Fatalf("window must start at completion: %+v %v", plan, err)
	}
	r.Status.CompletionTime = nil
	plan, err = Build(p, recorded(r, p), policy, now)
	if err != nil || plan.Decision != "refuse" {
		t.Fatalf("without completion time, creation time bounds the window: %+v %v", plan, err)
	}
}

func safetyFixture() (Pipeline, PipelineRun, Policy) {
	var p Pipeline
	p.Metadata.Name = "p"
	p.Spec.Tasks = []PipelineTask{{Name: "build"}, {Name: "consume"}, {Name: "fail"}}
	var r PipelineRun
	r.Metadata.Name = "r"
	r.Spec.PipelineRef.Name = "p"
	r.Status.ChildReferences = []ChildReference{{Name: "build-run", PipelineTask: "build", Status: "Succeeded"}, {Name: "consume-run", PipelineTask: "consume", Status: "Succeeded"}, {Name: "fail-run", PipelineTask: "fail", Status: "Failed"}}
	var policy Policy
	policy.Metadata.Name = "policy"
	policy.Spec.PipelineRef = "p"
	policy.Spec.Tasks = map[string]TaskRule{"fail": {RetryWith: []string{"build"}}}
	return p, r, policy
}

func contains(value, fragment string) bool {
	return strings.Contains(value, fragment)
}

func warningContains(warnings []Warning, fragment string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning.Code, fragment) || strings.Contains(warning.Message, fragment) {
			return true
		}
	}
	return false
}

func TestBuildRefusesBlockedTask(t *testing.T) {
	var p Pipeline
	p.Metadata.Name = "p"
	p.Spec.Tasks = []PipelineTask{{Name: "migrate"}}
	var r PipelineRun
	r.Metadata.Name = "r"
	r.Spec.PipelineRef.Name = "p"
	r.Status.ChildReferences = []ChildReference{{PipelineTask: "migrate", Status: "Failed"}}
	var policy Policy
	policy.Metadata.Name = "policy"
	policy.Spec.PipelineRef = "p"
	policy.Spec.Tasks = map[string]TaskRule{"migrate": {BlocksResume: true}}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Decision != "refuse" {
		t.Fatalf("got %q", plan.Decision)
	}
}

func TestBuildAllowsRetryClosureCycle(t *testing.T) {
	var p Pipeline
	p.Metadata.Name = "p"
	p.Spec.Tasks = []PipelineTask{{Name: "build"}, {Name: "scan"}}
	var r PipelineRun
	r.Metadata.Name = "r"
	r.Spec.PipelineRef.Name = "p"
	r.Status.ChildReferences = []ChildReference{{PipelineTask: "build", Status: "Succeeded"}, {PipelineTask: "scan", Status: "Failed"}}
	var policy Policy
	policy.Metadata.Name = "policy"
	policy.Spec.PipelineRef = "p"
	policy.Spec.Tasks = map[string]TaskRule{
		"build": {RetryWith: []string{"scan"}},
		"scan":  {RetryWith: []string{"build"}},
	}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Decision != "recover" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	for _, task := range plan.Tasks {
		if task.Action != "rerun" {
			t.Fatalf("task %q: got %q, want rerun", task.Name, task.Action)
		}
	}
}

func recorded(r PipelineRun, p Pipeline) PipelineRun {
	if r.Status.PipelineSpec == nil {
		r.Status.PipelineSpec = &struct {
			Tasks      []PipelineTask `json:"tasks"`
			Workspaces []Workspace    `json:"workspaces"`
		}{Tasks: p.Spec.Tasks, Workspaces: p.Spec.Workspaces}
	}
	return r
}

func TestBuildBlocksResumeOnlyWhenThatTaskFailed(t *testing.T) {
	p, r, policy := safetyFixture()
	policy.Spec.Tasks["build"] = TaskRule{BlocksResume: true}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil || plan.Decision != "recover" {
		t.Fatalf("blocksResume on a succeeded task pulled into the closure must not refuse: %+v %v", plan, err)
	}
	if !warningContains(plan.Warnings, "blocks-resume-task-rerun") {
		t.Fatalf("rerunning a blocksResume task must warn: %+v", plan.Warnings)
	}
	policy.Spec.Tasks["fail"] = TaskRule{BlocksResume: true}
	plan, err = Build(p, recorded(r, p), policy, time.Now())
	if err != nil || plan.Decision != "refuse" || !contains(plan.RefusalReason, `"fail" failed and blocks resume`) {
		t.Fatalf("blocksResume on the failed task must refuse: %+v %v", plan, err)
	}
}

func TestBuildSelectedTasksSeedTheClosure(t *testing.T) {
	p, r, policy := safetyFixture()
	policy.Spec.Tasks["consume"] = TaskRule{RetryWith: []string{"build"}}
	plan, err := BuildWithSelection(p, recorded(r, p), policy, nil, []string{"consume"}, time.Now())
	if err != nil || plan.Decision != "recover" {
		t.Fatalf("unexpected plan: %+v %v", plan, err)
	}
	for _, task := range plan.Tasks {
		if task.Action != "rerun" {
			t.Errorf("%s: got %s, want rerun", task.Name, task.Action)
		}
	}
	plan, _ = BuildWithSelection(p, recorded(r, p), policy, nil, []string{"missing"}, time.Now())
	if plan.Decision != "refuse" || !contains(plan.RefusalReason, "selected unknown task") {
		t.Fatalf("unknown selection must refuse: %+v", plan)
	}
}

func TestBuildRefusesWithoutRecordedDefinition(t *testing.T) {
	p, r, policy := safetyFixture()
	plan, err := Build(p, r, policy, time.Now())
	if err != nil || plan.Decision != "refuse" || !contains(plan.RefusalReason, "does not record") {
		t.Fatalf("unexpected plan: %+v %v", plan, err)
	}
}

func TestDefinitionMatchesTreatsVariablesAsWildcards(t *testing.T) {
	current := map[string]any{"script": `case "$(context.pipelineRun.name)" in x) exit 1 ;; esac`}
	if !definitionMatches(current, map[string]any{"script": `case "retry-integration-source" in x) exit 1 ;; esac`}) {
		t.Fatal("substituted variable must match")
	}
	if definitionMatches(current, map[string]any{"script": `case "retry-integration-source" in x) exit 2 ;; esac`}) {
		t.Fatal("a literal change must not match")
	}
	if definitionMatches(map[string]any{"retries": float64(2)}, map[string]any{"retries": float64(3)}) {
		t.Fatal("changed number must not match")
	}
}

func TestBuildRefusesPolicyKeyForUnknownTask(t *testing.T) {
	p, r, policy := safetyFixture()
	policy.Spec.Tasks["ghost-task"] = TaskRule{BlocksResume: true}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil || plan.Decision != "refuse" || !contains(plan.RefusalReason, `unknown task "ghost-task"`) {
		t.Fatalf("a policy entry for a missing task must refuse: %+v %v", plan, err)
	}
}
