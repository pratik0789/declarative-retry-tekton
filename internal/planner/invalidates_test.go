package planner

import (
	"testing"
	"time"
)

func staleBuildFixture() (Pipeline, PipelineRun) {
	var p Pipeline
	p.Metadata.Name = "fi"
	def := map[string]any{"tasks": []any{
		map[string]any{"name": "observe"},
		map[string]any{"name": "prepare", "runAfter": []any{"observe"}},
		map[string]any{"name": "build", "runAfter": []any{"prepare"}},
		map[string]any{"name": "gate", "runAfter": []any{"build"}},
		map[string]any{"name": "publish", "runAfter": []any{"gate"}},
	}}
	p.definition = def
	for _, n := range []string{"observe", "prepare", "build", "gate", "publish"} {
		p.Spec.Tasks = append(p.Spec.Tasks, PipelineTask{Name: n})
	}
	var r PipelineRun
	r.Metadata.Name = "fi-source"
	r.Spec.PipelineRef.Name = "fi"
	r.Status.Conditions = []Condition{{Type: "Succeeded", Status: "False"}}
	for _, n := range []string{"observe", "prepare", "build"} {
		r.Status.ChildReferences = append(r.Status.ChildReferences, ChildReference{Name: "fi-" + n, PipelineTask: n, Status: "Succeeded"})
	}
	r.Status.ChildReferences = append(r.Status.ChildReferences, ChildReference{Name: "fi-gate", PipelineTask: "gate", Status: "Failed"})
	r.recordedDefinition = def
	return p, r
}

func TestInvalidatesAddsCompletedLaterTasks(t *testing.T) {
	p, r := staleBuildFixture()
	var policy Policy
	policy.Metadata.Name = "pol"
	policy.Spec.PipelineRef = "fi"
	policy.Spec.Tasks = map[string]TaskRule{"prepare": {Invalidates: []string{"build"}}}
	plan, err := BuildWithSelection(p, r, policy, nil, []string{"prepare"}, time.Now())
	if err != nil || plan.Decision != "recover" {
		t.Fatalf("expected recovery: %v %s", err, plan.RefusalReason)
	}
	actions := map[string]string{}
	for _, task := range plan.Tasks {
		actions[task.Name] = task.Action
	}
	if actions["prepare"] != "rerun" || actions["build"] != "rerun" || actions["gate"] != "rerun" || actions["observe"] != "inherit" || actions["publish"] != "continue" {
		t.Fatalf("unexpected actions %v", actions)
	}
	without, _ := BuildWithSelection(p, r, Policy{Metadata: policy.Metadata, Spec: struct {
		PipelineRef      string              `json:"pipelineRef"`
		ResumeWithin     string              `json:"resumeWithin"`
		RetryOnlyLatest  bool                `json:"retryOnlyLatest"`
		TargetParameters []string            `json:"targetParameters,omitempty"`
		Tasks            map[string]TaskRule `json:"tasks"`
	}{PipelineRef: "fi"}}, nil, []string{"prepare"}, time.Now())
	for _, task := range without.Tasks {
		if task.Name == "build" && task.Action != "inherit" {
			t.Fatal("without invalidates the completed build is reused")
		}
	}
}

func TestInvalidatesRejectsEarlierTasks(t *testing.T) {
	p, r := staleBuildFixture()
	var policy Policy
	policy.Metadata.Name = "pol"
	policy.Spec.PipelineRef = "fi"
	policy.Spec.Tasks = map[string]TaskRule{"build": {Invalidates: []string{"prepare"}}}
	plan, _ := Build(p, r, policy, time.Now())
	if plan.Decision != "refuse" {
		t.Fatal("invalidating an earlier task must refuse")
	}
	policy.Spec.Tasks = map[string]TaskRule{"prepare": {Invalidates: []string{"nope"}}}
	if plan, _ := Build(p, r, policy, time.Now()); plan.Decision != "refuse" {
		t.Fatal("unknown invalidates target must refuse")
	}
}
