package planner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"sigs.k8s.io/yaml"
)

func TestPolicyContentIdentitySurvivesPresentationAndDetectsHonestEdits(t *testing.T) {
	p, r, policy := safetyFixture()
	r = recorded(r, p)
	plan, err := Build(p, r, policy, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var snapshot Policy
	if err := json.Unmarshal([]byte(`{"metadata":{"name":"copy"},"spec":`+string(plan.PolicySpec)+`}`), &snapshot); err != nil {
		t.Fatal(err)
	}
	copyPlan, err := Build(p, r, snapshot, time.Now())
	if err != nil || copyPlan.PolicySpecSHA256 != plan.PolicySpecSHA256 {
		t.Fatalf("metadata changed content identity: %+v %v", copyPlan, err)
	}
	if got := fmt.Sprintf("sha256:%x", sha256.Sum256(plan.PolicySpec)); got != plan.PolicySpecSHA256 {
		t.Fatalf("snapshot not bound: %s", got)
	}
	var reordered map[string]any
	if err := json.Unmarshal(plan.PolicySpec, &reordered); err != nil {
		t.Fatal(err)
	}
	presentation, err := yaml.Marshal(map[string]any{"metadata": map[string]any{"name": policy.Metadata.Name, "generation": policy.Metadata.Generation}, "spec": reordered})
	if err != nil {
		t.Fatal(err)
	}
	var parsed Policy
	if err := yaml.Unmarshal(presentation, &parsed); err != nil {
		t.Fatal(err)
	}
	parsedPlan, err := Build(p, r, parsed, time.Now())
	if err != nil || parsedPlan.PolicySpecSHA256 != plan.PolicySpecSHA256 {
		t.Fatalf("presentation changed identity: %+v %v", parsedPlan, err)
	}
	parsed.Spec.ResumeWithin = "1h"
	editedPlan, err := Build(p, r, parsed, time.Now())
	if err != nil || editedPlan.PolicySpecSHA256 == plan.PolicySpecSHA256 {
		t.Fatalf("honest content edit retained identity: %+v %v", editedPlan, err)
	}
	if parsed.Metadata.Name != policy.Metadata.Name || parsed.Metadata.Generation != policy.Metadata.Generation {
		t.Fatal("edit probe changed name/generation")
	}
}

func TestGeneratedRunRetainsVerifiableEffectivePolicy(t *testing.T) {
	out, err := createProbe(t, probePipeline, probeRun("persistentVolumeClaim: {claimName: shared-pvc}"), probeTaskRuns)
	if err != nil {
		t.Fatal(err)
	}
	annotations := asMap(asMap(out["metadata"])["annotations"])
	snapshot, ok := annotations[annotationPolicySpec].(string)
	if !ok || !json.Valid([]byte(snapshot)) {
		t.Fatalf("missing policy snapshot: %v", annotations)
	}
	if got := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(snapshot))); annotations[annotationPolicySpecSHA256] != got {
		t.Fatalf("wrong policy hash: %v", annotations)
	}
	var spec map[string]any
	if err := json.Unmarshal([]byte(snapshot), &spec); err != nil {
		t.Fatal(err)
	}
	if spec["pipelineRef"] != "probe" || spec["tasks"] == nil || spec["retryOnlyLatest"] != false {
		t.Fatalf("effective rules missing: %s", snapshot)
	}
}
