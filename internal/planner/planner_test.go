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
	plan, err := Build(p, r, policy, time.Now())
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
	plan, err := Build(p, r, policy, time.Now())
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
	plan, err := Build(p, r, policy, time.Now())
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
	plan, err := Build(p, r, policy, time.Now())
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
	plan, err := BuildWithNewerRuns(p, r, policy, []PipelineRun{newer}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Decision != "refuse" || !contains(plan.RefusalReason, "supersedes") {
		t.Fatalf("unexpected plan: %+v", plan)
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
	plan, err := Build(p, r, policy, time.Now())
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
	plan, err := Build(p, r, policy, time.Now())
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
