package planner

import (
	"testing"
	"time"
)

func identityFixture() (Pipeline, PipelineRun, Policy) {
	var p Pipeline
	p.Metadata.Name = "release"
	p.Spec.Tasks = []PipelineTask{{Name: "a"}, {Name: "b", RunAfter: []string{"a"}}}
	def := map[string]any{"tasks": []any{map[string]any{"name": "a"}, map[string]any{"name": "b", "runAfter": []any{"a"}}}}
	p.definition = def
	var r PipelineRun
	r.Metadata.Name = "release-run"
	r.Metadata.CreationTimestamp = time.Now().Add(-time.Minute)
	r.Status.Conditions = []Condition{{Type: "Succeeded", Status: "False"}}
	r.Status.ChildReferences = []ChildReference{{Name: "ta", PipelineTask: "a", Status: "Succeeded"}, {Name: "tb", PipelineTask: "b", Status: "Failed"}}
	r.recordedDefinition = def
	var policy Policy
	policy.Metadata.Name = "pol"
	policy.Spec.PipelineRef = "release"
	return p, r, policy
}

func TestResolverAndEmbeddedRunIdentity(t *testing.T) {
	p, r, policy := identityFixture()
	r.Metadata.Labels = map[string]string{"tekton.dev/pipeline": "release"}
	if plan, _ := Build(p, r, policy, time.Now()); plan.Decision != "recover" {
		t.Fatalf("resolver-labelled run should recover: %s", plan.RefusalReason)
	}
	r.Metadata.Labels = map[string]string{"tekton.dev/pipeline": "other"}
	if plan, _ := Build(p, r, policy, time.Now()); plan.Decision != "refuse" {
		t.Fatal("resolver run labelled with another Pipeline must refuse")
	}
	r.Metadata.Labels = nil
	if plan, _ := Build(p, r, policy, time.Now()); plan.Decision != "refuse" {
		t.Fatal("unnamed, unlabelled, non-embedded run must refuse")
	}
	r.inlineDefinition = r.recordedDefinition
	if plan, _ := Build(p, r, policy, time.Now()); plan.Decision != "recover" {
		t.Fatalf("embedded run with matching definition should recover: %s", plan.RefusalReason)
	}
	policy.Spec.RetryOnlyLatest = true
	if plan, _ := BuildWithNewerRuns(p, r, policy, []PipelineRun{}, time.Now()); plan.Decision != "refuse" {
		t.Fatal("retryOnlyLatest must refuse for an embedded source")
	}
	policy.Spec.RetryOnlyLatest = false
	r.Metadata.Annotations = map[string]string{annotationSourceRun: "earlier"}
	if plan, _ := Build(p, r, policy, time.Now()); plan.Decision != "refuse" {
		t.Fatal("embedded recovery runs created by retryctl stay refused")
	}
}
