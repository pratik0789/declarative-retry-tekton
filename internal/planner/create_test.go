package planner

import (
	"strings"
	"testing"
)

func TestCreateRequiresWarningConfirmation(t *testing.T) {
	_, err := CreateRecoveryRun(
		"../../testdata/result-conflict/pipeline.yaml",
		"../../testdata/result-conflict/run.yaml",
		"../../testdata/result-conflict/policy.yaml",
		"",
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "explicit confirmation") {
		t.Fatalf("expected confirmation error, got %v", err)
	}
}

func TestCreateRecordsAcceptedWarnings(t *testing.T) {
	manifest, err := CreateRecoveryRun(
		"../../testdata/result-conflict/pipeline.yaml",
		"../../testdata/result-conflict/run.yaml",
		"../../testdata/result-conflict/policy.yaml",
		"",
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), "shared-state-warnings-accepted: \"true\"") {
		t.Fatalf("accepted warning annotation missing:\n%s", manifest)
	}
}
