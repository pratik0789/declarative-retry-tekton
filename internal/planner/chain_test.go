package planner

import (
	"testing"
	"time"
)

func TestFailedRecoveryRunningBlocksResumeTaskRefuses(t *testing.T) {
	p, r := staleBuildFixture()
	r.Metadata.UID = "src-uid"
	r.Metadata.CreationTimestamp = time.Now().Add(-time.Hour)
	var policy Policy
	policy.Metadata.Name = "pol"
	policy.Spec.PipelineRef = "fi"
	policy.Spec.Tasks = map[string]TaskRule{"gate": {RetryWith: []string{"build"}}, "build": {BlocksResume: true}}
	var recovery PipelineRun
	recovery.Metadata.Name = "fi-retry-1"
	recovery.Metadata.CreationTimestamp = time.Now()
	recovery.Metadata.Annotations = map[string]string{annotationSourceRun: "fi-source", annotationSourceRunUID: "src-uid", annotationSourcePipeline: "fi"}
	recovery.Status.Conditions = []Condition{{Type: "Succeeded", Status: "False"}}
	recovery.inlineDefinition = map[string]any{"tasks": []any{map[string]any{"name": "build"}, map[string]any{"name": "gate"}}}
	plan, err := BuildWithNewerRuns(p, r, policy, []PipelineRun{recovery}, time.Now())
	if err != nil || plan.Decision != "refuse" {
		t.Fatalf("expected refusal after a failed recovery ran a blocksResume task: %v %+v", err, plan.Decision)
	}
	recovery.inlineDefinition = map[string]any{"tasks": []any{map[string]any{"name": "gate"}}}
	if plan, _ := BuildWithNewerRuns(p, r, policy, []PipelineRun{recovery}, time.Now()); plan.Decision != "recover" {
		t.Fatalf("a failed recovery without blocksResume tasks must not block: %s", plan.RefusalReason)
	}
}
