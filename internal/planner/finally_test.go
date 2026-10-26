package planner

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"testing"
	"time"
)

func TestFinallyExportedResultRecovery(t *testing.T) {
	cases := []struct {
		name, policy, reference string
		result                  any
		want                    string
		refuse                  bool
	}{
		{"reused string result", "only", "$(finally.audit.results.note)", "checked", "checked", false},
		{"embedded string result", "only", "audit:$(finally.audit.results.note)", "checked", "audit:checked", false},
		{"rerun result stays referenced", "parallel", "$(finally.audit.results.note)", "checked", "$(finally.audit.results.note)", false},
		{"missing result", "only", "$(finally.audit.results.note)", nil, "", true},
		{"array result", "only", "$(finally.audit.results.note)", []any{"checked"}, "", true},
		{"object result", "only", "$(finally.audit.results.note)", map[string]any{"note": "checked"}, "", true},
		{"status reference", "only", "$(finally.audit.status)", "checked", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := "../../testdata/finally-recovery"
			dir := t.TempDir()
			pipeline, err := readRaw(filepath.Join(root, "pipeline.json"))
			if err != nil {
				t.Fatal(err)
			}
			run, err := readRaw(filepath.Join(root, "run.json"))
			if err != nil {
				t.Fatal(err)
			}
			trs, err := readRaw(filepath.Join(root, "taskruns.json"))
			if err != nil {
				t.Fatal(err)
			}
			spec := asMap(pipeline["spec"])
			audit := asMap(asList(spec["finally"])[1])
			asMap(audit["taskSpec"])["results"] = []any{map[string]any{"name": "note", "type": "string"}}
			spec["results"] = []any{map[string]any{"name": "audit-note", "value": tc.reference}}
			asMap(run["status"])["pipelineSpec"] = toAny(spec)
			if tc.result != nil {
				asMap(asMap(asList(trs["items"])[3])["status"])["results"] = []any{map[string]any{"name": "note", "value": tc.result}}
			}
			for name, value := range map[string]any{"pipeline": pipeline, "run": run, "taskruns": trs} {
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, name+".json"), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			data, err := CreateRecoveryRun(filepath.Join(dir, "pipeline.json"), filepath.Join(dir, "run.json"), filepath.Join(root, tc.policy+"-policy.json"), filepath.Join(dir, "taskruns.json"), "", nil, true)
			if tc.refuse {
				if err == nil || !strings.Contains(err.Error(), `inherited task "audit"`) || len(data) != 0 {
					t.Fatalf("unresolvable finally reference must refuse without a manifest: %s %v", data, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var out map[string]any
			if err := yaml.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			recoverySpec := asMap(asMap(out["spec"])["pipelineSpec"])
			want := tc.want
			if tc.policy == "only" {
				publishers := asList(recoverySpec["tasks"])
				if len(publishers) != 1 {
					t.Fatalf("want one ordinary result publisher, got %v", publishers)
				}
				publisher := asMap(publishers[0])
				publisherSpec := asMap(publisher["taskSpec"])
				step := asMap(asList(publisherSpec["steps"])[0])
				encoded := asMap(asList(step["env"])[0])["value"].(string)
				decoded, err := base64.StdEncoding.DecodeString(encoded)
				if err != nil || string(decoded) != tc.result {
					t.Fatalf("publisher must preserve source bytes: %s %v", decoded, err)
				}
				want = strings.ReplaceAll(tc.reference, "$(finally.audit.results.note)", "$(tasks."+publisher["name"].(string)+".results.value-0)")
				if asMap(asMap(out["metadata"])["annotations"])[annotationResultPublishers] != publisher["name"] {
					t.Fatalf("publisher annotation missing: %v", out)
				}
			}
			if got := asMap(asList(recoverySpec["results"])[0])["value"]; got != want {
				t.Fatalf("exported finally result = %v, want %q", got, want)
			}
			present := false
			for _, task := range asList(recoverySpec["finally"]) {
				present = present || asMap(task)["name"] == "audit"
			}
			if present != (tc.policy == "parallel") {
				t.Fatalf("audit presence must match reuse/rerun: %v", recoverySpec)
			}
		})
	}
}

func TestPipelineResultPublishersGroupByProducerAndAvoidNameCollisions(t *testing.T) {
	node := map[string]any{"results": []any{
		map[string]any{"name": "ordinary", "value": "$(tasks.build.results.image)"},
		map[string]any{"name": "final", "value": "$(finally.audit.results.note)"},
		map[string]any{"name": "repeated", "value": "$(finally.audit.results.note)"},
		map[string]any{"name": "second", "value": "$(finally.audit.results.other)"},
		map[string]any{"name": "remaining", "value": "$(finally.notify.results.receipt)"},
	}}
	publishers, err := publishInheritedPipelineResults(node, map[string]TaskPlan{
		"build": {Results: map[string]any{"image": "source-image"}},
		"audit": {Results: map[string]any{"note": "checked", "other": "\n'\"$()\\\n"}},
	}, map[string][]string{"retryctl-reused-results": nil, "build": nil, "audit": nil, "notify": nil})
	if err != nil {
		t.Fatal(err)
	}
	if len(publishers) != 2 || asMap(publishers[0])["name"] != "retryctl-reused-results-1" || asMap(publishers[1])["name"] != "retryctl-reused-results-2" {
		t.Fatalf("publishers must be per producer with distinct names: %v", publishers)
	}
	if len(asList(asMap(asMap(publishers[0])["taskSpec"])["results"])) != 1 || len(asList(asMap(asMap(publishers[1])["taskSpec"])["results"])) != 2 {
		t.Fatalf("duplicate result references must share published outputs: %v", publishers)
	}
	results := asList(node["results"])
	if asMap(results[1])["value"] != asMap(results[2])["value"] || asMap(results[4])["value"] != "$(finally.notify.results.receipt)" {
		t.Fatalf("deduplication or remaining reference changed: %v", results)
	}
}

func TestFinallyRecoveryGeneration(t *testing.T) {
	root := "../../testdata/finally-recovery"
	tests := []struct {
		policy          string
		selected        []string
		ordinary, final string
	}{
		{"only", nil, "", "notify"},
		{"upstream", nil, "prepare,build", "notify"},
		{"parallel", nil, "", "notify,audit"},
		{"only", []string{"audit"}, "", "notify,audit"},
	}
	for _, tt := range tests {
		t.Run(tt.policy+strings.Join(tt.selected, ""), func(t *testing.T) {
			path := func(name string) string { return filepath.Join(root, name+".json") }
			data, err := CreateRecoveryRun(path("pipeline"), path("run"), path(tt.policy+"-policy"), path("taskruns"), "", tt.selected, true)
			if err != nil {
				t.Fatal(err)
			}
			var out map[string]any
			if err := yaml.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			spec := asMap(asMap(out["spec"])["pipelineSpec"])
			names := func(raw any) string {
				var names []string
				for _, item := range asList(raw) {
					names = append(names, asMap(item)["name"].(string))
				}
				return strings.Join(names, ",")
			}
			ordinary := names(spec["tasks"])
			if tt.ordinary == "" {
				anchor := asMap(asMap(out["metadata"])["annotations"])[annotationFinallyAnchor]
				if ordinary != anchor {
					t.Fatalf("missing skipped final-phase anchor: %v", spec)
				}
				when := asMap(asList(asMap(asList(spec["tasks"])[0])["when"])[0])
				if when["input"] != "false" || asList(when["values"])[0] != "true" {
					t.Fatalf("anchor can execute: %v", when)
				}
				ordinary = ""
			}
			if ordinary != tt.ordinary || names(spec["finally"]) != tt.final {
				t.Fatalf("unexpected phases: %+v", spec)
			}
			if tt.ordinary == "" {
				params := asList(asMap(asList(spec["finally"])[0])["params"])
				if asMap(params[0])["value"] != "source-image" {
					t.Fatalf("inherited result not substituted: %v", params)
				}
			}
			specs := asList(asMap(out["spec"])["taskRunSpecs"])
			if len(specs) != len(strings.FieldsFunc(tt.ordinary, func(r rune) bool { return r == ',' }))+len(asList(spec["finally"])) {
				t.Fatalf("taskRunSpecs for inherited tasks retained: %v", specs)
			}
			annotations := asMap(asMap(out["metadata"])["annotations"])
			if tt.final == "notify" && !strings.Contains(annotations[annotationInheritedRuns].(string), "audit=source-audit") {
				t.Fatalf("finally reuse provenance missing: %v", annotations)
			}
		})
	}
}

func TestFinallyUsesOrdinaryClosureAndBlockRules(t *testing.T) {
	var p Pipeline
	p.Metadata.Name = "finally"
	p.Spec.Tasks = []PipelineTask{{Name: "build"}}
	p.Spec.Finally = []PipelineTask{{Name: "notify"}, {Name: "audit"}}
	var r PipelineRun
	r.Metadata.Name = "source"
	r.Spec.PipelineRef.Name = p.Metadata.Name
	r.Status.ChildReferences = []ChildReference{{PipelineTask: "build", Status: "Succeeded"}, {PipelineTask: "notify", Status: "Failed"}, {PipelineTask: "audit", Status: "Succeeded"}}
	var policy Policy
	policy.Metadata.Name = "policy"
	policy.Spec.PipelineRef = p.Metadata.Name
	policy.Spec.Tasks = map[string]TaskRule{"notify": {RetryWith: []string{"audit"}}, "audit": {BlocksResume: true}}
	plan, err := Build(p, recorded(r, p), policy, time.Now())
	if err != nil || plan.Decision != "recover" || !warningContains(plan.Warnings, "blocks-resume-task-rerun") {
		t.Fatalf("successful blocked finally task should rerun with warning: %+v %v", plan, err)
	}
	policy.Spec.Tasks["notify"] = TaskRule{BlocksResume: true}
	plan, err = Build(p, recorded(r, p), policy, time.Now())
	if err != nil || plan.Decision != "refuse" || !strings.Contains(plan.RefusalReason, "blocks resume") {
		t.Fatalf("failed blocked finally task should refuse: %+v %v", plan, err)
	}
}

func TestFinallyAnchorDoesNotReuseOriginalTaskName(t *testing.T) {
	root := "../../testdata/finally-recovery"
	dir := t.TempDir()
	for _, name := range []string{"pipeline", "run", "taskruns", "only-policy"} {
		data, err := os.ReadFile(filepath.Join(root, name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		data = []byte(strings.ReplaceAll(string(data), "audit", "retryctl-finally-anchor"))
		if err := os.WriteFile(filepath.Join(dir, name+".json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	path := func(name string) string { return filepath.Join(dir, name+".json") }
	data, err := CreateRecoveryRun(path("pipeline"), path("run"), path("only-policy"), path("taskruns"), "", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	anchor := asMap(asMap(out["metadata"])["annotations"])[annotationFinallyAnchor]
	if anchor != "retryctl-finally-anchor-1" {
		t.Fatalf("generated anchor collides with original task: %v", anchor)
	}
}
