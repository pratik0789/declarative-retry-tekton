package planner

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuildRetryOnlyLatestEffectiveTargetParameters(t *testing.T) {
	tests := []struct {
		name              string
		sourceDefault     any
		sourceValue       any
		candidateDefault  any
		candidateValue    any
		candidateRecorded bool
		candidateEmbedded bool
		older             bool
		otherPipeline     bool
		wantDecision      string
		wantReason        string
	}{
		{name: "omitted source matches explicit default", sourceDefault: "main", candidateValue: "main", wantDecision: "refuse", wantReason: "supersedes"},
		{name: "explicit source matches recorded candidate default", sourceDefault: "main", sourceValue: "main", candidateDefault: "main", candidateRecorded: true, wantDecision: "refuse", wantReason: "supersedes"},
		{name: "both runs use recorded defaults", sourceDefault: "main", candidateDefault: "main", candidateRecorded: true, wantDecision: "refuse", wantReason: "supersedes"},
		{name: "both runs explicitly match", sourceDefault: "main", sourceValue: "main", candidateValue: "main", wantDecision: "refuse", wantReason: "supersedes"},
		{name: "explicit source overrides its default", sourceDefault: "main", sourceValue: "feature", candidateValue: "main", wantDecision: "recover"},
		{name: "explicit candidate overrides its default", sourceDefault: "main", candidateDefault: "main", candidateRecorded: true, candidateValue: "feature", wantDecision: "recover"},
		{name: "candidate uses its own recorded default", sourceDefault: "main", candidateDefault: "feature", candidateRecorded: true, wantDecision: "recover"},
		{name: "candidate uses embedded recovery default", sourceDefault: "main", sourceValue: "main", candidateDefault: "main", candidateEmbedded: true, wantDecision: "refuse", wantReason: "supersedes"},
		{name: "unknown candidate default refuses", sourceDefault: "main", wantDecision: "refuse", wantReason: "cannot resolve target parameter"},
		{name: "recorded candidate has no default", sourceDefault: "main", candidateRecorded: true, wantDecision: "refuse", wantReason: "cannot resolve target parameter"},
		{name: "unknown source value refuses", candidateValue: "main", wantDecision: "refuse", wantReason: "cannot resolve target parameter"},
		{name: "older unresolved run is ignored", sourceDefault: "main", older: true, wantDecision: "recover"},
		{name: "unrelated unresolved run is ignored", sourceDefault: "main", otherPipeline: true, wantDecision: "recover"},
		{name: "array default matches explicit value", sourceDefault: []any{"main", "release"}, candidateValue: []any{"main", "release"}, wantDecision: "refuse", wantReason: "supersedes"},
		{name: "object default fills omitted keys", sourceDefault: map[string]any{"branch": "main", "repository": "example"}, sourceValue: map[string]any{"branch": "release"}, candidateValue: map[string]any{"branch": "release", "repository": "example"}, wantDecision: "refuse", wantReason: "supersedes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, source, policy := safetyFixture()
			p.definition = targetDefinition(p, tt.sourceDefault)
			source.recordedDefinition = p.definition
			if tt.sourceValue != nil {
				source.Spec.Params = []Param{{Name: "branch", Value: tt.sourceValue}}
			}
			now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
			source.Metadata.CreationTimestamp = now.Add(-time.Hour)
			policy.Spec.RetryOnlyLatest = true
			policy.Spec.TargetParameters = []string{"branch"}

			var candidate PipelineRun
			candidate.Metadata.Name = "newer"
			candidate.Metadata.CreationTimestamp = now.Add(-time.Minute)
			candidate.Spec.PipelineRef.Name = p.Metadata.Name
			if tt.candidateValue != nil {
				candidate.Spec.Params = []Param{{Name: "branch", Value: tt.candidateValue}}
			}
			if tt.candidateRecorded {
				candidate.recordedDefinition = targetDefinition(p, tt.candidateDefault)
			}
			if tt.candidateEmbedded {
				data, err := json.Marshal(map[string]any{
					"metadata": map[string]any{
						"name":              "newer",
						"creationTimestamp": candidate.Metadata.CreationTimestamp,
						"annotations":       map[string]string{annotationSourcePipeline: p.Metadata.Name},
					},
					"spec": map[string]any{"pipelineSpec": targetDefinition(p, tt.candidateDefault)},
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(data, &candidate); err != nil {
					t.Fatal(err)
				}
			}
			if tt.older {
				candidate.Metadata.CreationTimestamp = now.Add(-2 * time.Hour)
			}
			if tt.otherPipeline {
				candidate.Spec.PipelineRef.Name = "other"
			}

			plan, err := BuildWithNewerRuns(p, source, policy, []PipelineRun{candidate}, now)
			if err != nil {
				t.Fatal(err)
			}
			if plan.Decision != tt.wantDecision || !contains(plan.RefusalReason, tt.wantReason) {
				t.Fatalf("got decision %q, reason %q; want %q containing %q", plan.Decision, plan.RefusalReason, tt.wantDecision, tt.wantReason)
			}
		})
	}
}

func targetDefinition(p Pipeline, defaultValue any) any {
	definition := toAny(p.Spec).(map[string]any)
	param := map[string]any{"name": "branch"}
	if defaultValue != nil {
		param["default"] = defaultValue
	}
	definition["params"] = []any{param}
	return definition
}

func TestBuildRetryOnlyLatestExcludesKnownDifferentTargets(t *testing.T) {
	p, source, policy := safetyFixture()
	p.definition = targetDefinition(p, "main")
	source.recordedDefinition = p.definition
	source.Spec.Params = []Param{{Name: "repository", Value: "example"}}
	policy.Spec.RetryOnlyLatest = true
	// An unresolved parameter comes first; a later known mismatch still proves
	// that this candidate is not the source's target.
	policy.Spec.TargetParameters = []string{"repository", "branch"}
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	source.Metadata.CreationTimestamp = now.Add(-time.Hour)
	var candidate PipelineRun
	candidate.Metadata.Name = "newer"
	candidate.Metadata.CreationTimestamp = now
	candidate.Spec.PipelineRef.Name = p.Metadata.Name
	candidate.Spec.Params = []Param{{Name: "branch", Value: "feature"}}
	plan, err := BuildWithNewerRuns(p, source, policy, []PipelineRun{candidate}, now)
	if err != nil || plan.Decision != "recover" {
		t.Fatalf("known different branch must exclude candidate: %+v %v", plan, err)
	}
	candidate.Spec.Params[0].Value = "main"
	plan, err = BuildWithNewerRuns(p, source, policy, []PipelineRun{candidate}, now)
	if err != nil || plan.Decision != "refuse" || !contains(plan.RefusalReason, "cannot resolve target parameter") {
		t.Fatalf("otherwise matching unresolved target must refuse: %+v %v", plan, err)
	}
}
