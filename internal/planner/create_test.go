package planner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestCreateRequiresWarningConfirmation(t *testing.T) {
	_, err := CreateRecoveryRun(
		"../../testdata/result-conflict/pipeline.yaml",
		"../../testdata/result-conflict/run.yaml",
		"../../testdata/result-conflict/policy.yaml",
		"",
		"",
		nil,
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
		"",
		nil,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), "shared-state-warnings-accepted: \"true\"") {
		t.Fatalf("accepted warning annotation missing:\n%s", manifest)
	}
}

const probePipeline = `
apiVersion: tekton.dev/v1
kind: Pipeline
metadata: {name: probe}
spec:
  params: [{name: repo}]
  workspaces: [{name: shared}]
  tasks:
  - name: clone
    params: [{name: url, value: $(params.repo)}]
    workspaces: [{name: src, workspace: shared, subPath: code}]
    taskRef: {name: git-clone}
  - name: build
    runAfter: [clone]
    retries: 2
    timeout: 30m
    params: [{name: revision, value: "rev-$(tasks.clone.results.commit)"}]
    workspaces: [{name: src, workspace: shared, subPath: code}]
    taskRef: {name: kaniko}
  finally:
  - name: notify
    params: [{name: commit, value: $(tasks.clone.results.commit)}]
    taskRef: {name: notify}
`

const probePolicy = `
apiVersion: retry.tekton.dev/v1alpha1
kind: PipelineRetryPolicy
metadata: {name: probe-policy, generation: 7}
spec: {pipelineRef: probe, tasks: {}}
`

const probeTaskRuns = `
items:
- metadata:
    name: probe-1-clone
    labels: {tekton.dev/pipelineRun: probe-1, tekton.dev/pipelineTask: clone}
  status:
    conditions: [{type: Succeeded, status: "True"}]
    results: [{name: commit, value: a1b2c3}]
- metadata:
    name: probe-1-build
    labels: {tekton.dev/pipelineRun: probe-1, tekton.dev/pipelineTask: build}
  status:
    conditions: [{type: Succeeded, status: "False"}]
`

func probeRun(binding string) string {
	return `
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata: {name: probe-1, namespace: ci, uid: abc}
spec:
  pipelineRef: {name: probe}
  params: [{name: repo, value: https://example.invalid/r.git}]
  taskRunTemplate: {serviceAccountName: builder}
  workspaces:
  - name: shared
    ` + binding + `
`
}

func createProbe(t *testing.T, pipeline, run, taskRuns string) (map[string]any, error) {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{"pipeline.yaml": pipeline, "run.yaml": run, "policy.yaml": probePolicy, "taskruns.yaml": taskRuns}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	path := func(name string) string { return filepath.Join(dir, name) }
	var definition struct {
		Spec any `json:"spec"`
	}
	if err := yaml.Unmarshal([]byte(pipeline), &definition); err != nil {
		t.Fatal(err)
	}
	recordedSpec, _ := json.Marshal(definition.Spec)
	if err := os.WriteFile(path("run.yaml"), []byte(run+"status:\n  pipelineSpec: "+string(recordedSpec)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := CreateRecoveryRun(path("pipeline.yaml"), path("run.yaml"), path("policy.yaml"), path("taskruns.yaml"), "", nil, true)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := yaml.Unmarshal(manifest, &out); err != nil {
		t.Fatal(err)
	}
	return out, nil
}

func TestCreateSubstitutesInheritedResultsAndPreservesFields(t *testing.T) {
	out, err := createProbe(t, probePipeline, probeRun("persistentVolumeClaim: {claimName: shared-pvc}"), probeTaskRuns)
	if err != nil {
		t.Fatal(err)
	}
	spec := asMap(out["spec"])
	pipelineSpec := asMap(spec["pipelineSpec"])
	tasks := asList(pipelineSpec["tasks"])
	if len(tasks) != 1 {
		t.Fatalf("want only build in recovery run, got %v", tasks)
	}
	build := asMap(tasks[0])
	if _, ok := build["runAfter"]; ok {
		t.Errorf("runAfter to inherited clone was not pruned: %v", build["runAfter"])
	}
	if build["retries"] != float64(2) || build["timeout"] != "30m" {
		t.Errorf("task fields dropped: %v", build)
	}
	if got := asMap(asList(build["params"])[0])["value"]; got != "rev-a1b2c3" {
		t.Errorf("inherited result not substituted: %v", got)
	}
	if got := asMap(asList(build["workspaces"])[0])["subPath"]; got != "code" {
		t.Errorf("subPath dropped: %v", got)
	}
	if got := asMap(asList(asMap(asList(pipelineSpec["finally"])[0])["params"])[0])["value"]; got != "a1b2c3" {
		t.Errorf("finally result not substituted: %v", got)
	}
	if len(asList(pipelineSpec["params"])) != 1 || len(asList(spec["params"])) != 1 {
		t.Errorf("pipeline or run params dropped: %v / %v", pipelineSpec["params"], spec["params"])
	}
	if asMap(spec["taskRunTemplate"])["serviceAccountName"] != "builder" {
		t.Errorf("taskRunTemplate dropped: %v", spec)
	}
	if _, ok := spec["pipelineRef"]; ok {
		t.Errorf("pipelineRef must be replaced by pipelineSpec")
	}
	metadata := asMap(out["metadata"])
	annotations := asMap(metadata["annotations"])
	if annotations[annotationPolicyGeneration] != "7" || annotations[annotationInheritedRuns] != "clone=probe-1-clone" || annotations[annotationSourcePipeline] != "probe" {
		t.Errorf("provenance annotations missing: %v", annotations)
	}
	if metadata["namespace"] != "ci" {
		t.Errorf("namespace not preserved: %v", metadata)
	}
}

func TestCreateRefusesUnresolvableInheritedResult(t *testing.T) {
	taskRuns := strings.Replace(probeTaskRuns, "results: [{name: commit, value: a1b2c3}]", "results: []", 1)
	_, err := createProbe(t, probePipeline, probeRun("persistentVolumeClaim: {claimName: shared-pvc}"), taskRuns)
	if err == nil || !strings.Contains(err.Error(), `inherited task "clone"`) {
		t.Fatalf("expected unresolved-reference refusal, got %v", err)
	}
}

func TestCreateWarnsButProceedsForEphemeralWorkspace(t *testing.T) {
	for _, binding := range []string{"volumeClaimTemplate: {spec: {accessModes: [ReadWriteOnce]}}", "emptyDir: {}"} {
		out, err := createProbe(t, probePipeline, probeRun(binding), probeTaskRuns)
		if err != nil {
			t.Fatalf("%s: shared-state findings must not refuse recovery: %v", binding, err)
		}
		codes, _ := asMap(asMap(out["metadata"])["annotations"])[annotationWarningCodes].(string)
		if !strings.Contains(codes, "ephemeral-workspace-across-closure") {
			t.Fatalf("%s: expected ephemeral-workspace warning, got %q", binding, codes)
		}
	}
}

func TestCreateKeepsOrderingThroughInheritedTasks(t *testing.T) {
	pipeline := `
apiVersion: tekton.dev/v1
kind: Pipeline
metadata: {name: probe}
spec:
  tasks:
  - {name: clone, taskRef: {name: t}}
  - {name: build, runAfter: [clone], taskRef: {name: t}}
  - {name: scan, runAfter: [build], taskRef: {name: t}}
  - {name: migrate, runAfter: [scan], taskRef: {name: t}}
  - {name: deploy, runAfter: [migrate], taskRef: {name: t}}
`
	run := `
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata: {name: probe-1}
spec: {pipelineRef: {name: probe}}
`
	taskRuns := `
items:
- metadata: {name: c, labels: {tekton.dev/pipelineRun: probe-1, tekton.dev/pipelineTask: clone}}
  status: {conditions: [{type: Succeeded, status: "True"}]}
- metadata: {name: b, labels: {tekton.dev/pipelineRun: probe-1, tekton.dev/pipelineTask: build}}
  status: {conditions: [{type: Succeeded, status: "True"}]}
- metadata: {name: s, labels: {tekton.dev/pipelineRun: probe-1, tekton.dev/pipelineTask: scan}}
  status: {conditions: [{type: Succeeded, status: "True"}]}
- metadata: {name: m, labels: {tekton.dev/pipelineRun: probe-1, tekton.dev/pipelineTask: migrate}}
  status: {conditions: [{type: Succeeded, status: "True"}]}
- metadata: {name: d, labels: {tekton.dev/pipelineRun: probe-1, tekton.dev/pipelineTask: deploy}}
  status: {conditions: [{type: Succeeded, status: "False"}]}
`
	dir := t.TempDir()
	var definition struct {
		Spec any `json:"spec"`
	}
	if err := yaml.Unmarshal([]byte(pipeline), &definition); err != nil {
		t.Fatal(err)
	}
	recordedSpec, _ := json.Marshal(definition.Spec)
	policy := "apiVersion: retry.tekton.dev/v1alpha1\nkind: PipelineRetryPolicy\nmetadata: {name: p}\nspec: {pipelineRef: probe, tasks: {scan: {retryWith: [build]}}}\n"
	files := map[string]string{"pipeline.yaml": pipeline, "run.yaml": run + "status:\n  pipelineSpec: " + string(recordedSpec) + "\n", "policy.yaml": policy, "taskruns.yaml": taskRuns}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	path := func(name string) string { return filepath.Join(dir, name) }
	manifest, err := CreateRecoveryRun(path("pipeline.yaml"), path("run.yaml"), path("policy.yaml"), path("taskruns.yaml"), "", []string{"scan"}, true)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := yaml.Unmarshal(manifest, &out); err != nil {
		t.Fatal(err)
	}
	for _, raw := range asList(asMap(asMap(out["spec"])["pipelineSpec"])["tasks"]) {
		task := asMap(raw)
		if task["name"] == "deploy" {
			if got := asList(task["runAfter"]); len(got) != 1 || got[0] != "scan" {
				t.Fatalf("deploy must still wait for the rerun scan through inherited migrate, got runAfter=%v", got)
			}
			return
		}
	}
	t.Fatal("deploy missing from recovery run")
}
