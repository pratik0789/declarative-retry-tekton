package planner

import (
	"testing"
	"time"
)

func TestNoRepeatRefusesRerunningCompletedTask(t *testing.T) {
	p, r := staleBuildFixture()
	var policy Policy
	policy.Metadata.Name = "pol"
	policy.Spec.PipelineRef = "fi"
	policy.Spec.Tasks = map[string]TaskRule{"gate": {RetryWith: []string{"build"}}, "build": {NoRepeat: true}}
	if plan, _ := Build(p, r, policy, time.Now()); plan.Decision != "refuse" {
		t.Fatal("a closure that reruns a completed noRepeat task must refuse")
	}
	policy.Spec.Tasks = map[string]TaskRule{"build": {NoRepeat: true}}
	if plan, _ := BuildWithSelection(p, r, policy, nil, []string{"build"}, time.Now()); plan.Decision != "refuse" {
		t.Fatal("selecting a completed noRepeat task must refuse")
	}
	if plan, _ := Build(p, r, policy, time.Now()); plan.Decision != "recover" {
		t.Fatalf("reusing a noRepeat task is fine: %s", plan.RefusalReason)
	}
}
