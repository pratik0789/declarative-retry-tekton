package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"testing"
	"time"
)

func TestRecoveryRequiresTerminalSourceAndChildren(t *testing.T) {
	for _, status := range []string{"Unknown", "True", ""} {
		p, r, policy := safetyFixture()
		r = recorded(r, p)
		if status == "" {
			r.Status.Conditions = nil
		} else {
			r.Status.Conditions = []Condition{{Type: "Succeeded", Status: status}}
		}
		plan, err := Build(p, r, policy, time.Now())
		if err != nil || plan.Decision != "refuse" || !strings.Contains(plan.RefusalReason, "terminal Succeeded=False") {
			t.Fatalf("%q: %+v %v", status, plan, err)
		}
	}
	p, r, policy := safetyFixture()
	r = recorded(r, p)
	r.Status.ChildReferences = append(r.Status.ChildReferences, ChildReference{Name: "active-finally", PipelineTask: "notify", Status: "Unknown"})
	plan, err := Build(p, r, policy, time.Now())
	if err != nil || plan.Decision != "refuse" || !strings.Contains(plan.RefusalReason, "terminal status") {
		t.Fatalf("active child: %+v %v", plan, err)
	}
}

func TestWarningsCoverFullTaskDefinitions(t *testing.T) {
	for _, field := range []string{"when", "taskSpec", "params"} {
		for _, namespace := range []string{"tasks", "finally"} {
			t.Run(field+"/"+namespace, func(t *testing.T) {
				p, r, policy := safetyFixture()
				raw := toAny(p.Spec).(map[string]any)
				consumer := raw["tasks"].([]any)[1].(map[string]any)
				ref := "$(" + namespace + ".build.results.flag)"
				switch field {
				case "when":
					consumer[field] = []any{map[string]any{"input": ref, "operator": "in", "values": []any{"yes"}}}
				case "taskSpec":
					consumer[field] = map[string]any{"steps": []any{map[string]any{"name": "s", "script": "echo " + ref}}}
				case "params":
					consumer[field] = []any{map[string]any{"name": "flag", "value": ref}}
				}
				p.definition = raw
				r.recordedDefinition = raw
				r.Status.Conditions = []Condition{{Type: "Succeeded", Status: "False"}}
				plan, err := Build(p, r, policy, time.Now())
				if err != nil || plan.Decision != "recover" || !warningContains(plan.Warnings, "shared-result-across-closure") {
					t.Fatalf("%+v %v", plan, err)
				}
			})
		}
	}
}

// Independent oracle: boolean Floyd-Warshall, not the planner's recursive
// nearest-retained traversal. These are structural transformations of abstract
// partitions, not claims that every partition is a realizable failed run.
func warshall(a [][]bool) [][]bool {
	n := len(a)
	out := make([][]bool, n)
	for i := range a {
		out[i] = append([]bool(nil), a[i]...)
	}
	for k := 0; k < n; k++ {
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				out[i][j] = out[i][j] || (out[i][k] && out[k][j])
			}
		}
	}
	return out
}
func TestExhaustiveRecoveryProjection(t *testing.T) {
	const n = 5
	dir := t.TempDir()
	path := func(s string) string { return filepath.Join(dir, s) }
	write := func(name string, v any) {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path(name), b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	edges := [][2]int{}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			edges = append(edges, [2]int{i, j})
		}
	}
	cases := 0
	for mask := 0; mask < 1<<len(edges); mask++ {
		matrix := make([][]bool, n)
		tasks := make([]any, n)
		for i := 0; i < n; i++ {
			matrix[i] = make([]bool, n)
			tasks[i] = map[string]any{"name": fmt.Sprintf("t%d", i), "taskSpec": map[string]any{"results": []any{map[string]any{"name": "out"}}, "steps": []any{map[string]any{"name": "s", "image": "alpine:3.20", "script": "echo ok"}}}}
		}
		for bit, e := range edges {
			if mask&(1<<bit) == 0 {
				continue
			}
			i, j := e[0], e[1]
			matrix[i][j] = true
			task := tasks[j].(map[string]any)
			parent := fmt.Sprintf("t%d", i)
			ref := "$(tasks." + parent + ".results.out)"
			switch (i + j) % 3 {
			case 0:
				task["runAfter"] = appendList(task["runAfter"], parent)
			case 1:
				task["params"] = appendList(task["params"], map[string]any{"name": parent, "value": ref})
			case 2:
				task["when"] = appendList(task["when"], map[string]any{"input": ref, "operator": "in", "values": []any{"ok"}})
			}
		}
		spec := map[string]any{"tasks": tasks}
		write("pipeline.json", map[string]any{"kind": "Pipeline", "metadata": map[string]any{"name": "projection"}, "spec": spec})
		write("policy.json", map[string]any{"metadata": map[string]any{"name": "empty"}, "spec": map[string]any{"pipelineRef": "projection", "tasks": map[string]any{}}})
		original := warshall(matrix)
		for kept := 1; kept < 1<<n; kept++ {
			failed := 0
			for i := 0; i < n; i++ {
				if kept&(1<<i) != 0 {
					failed = i
				}
			}
			children := []any{}
			selected := []string{}
			for i := 0; i < n; i++ {
				name := fmt.Sprintf("t%d", i)
				status := "Succeeded"
				if i == failed {
					status = "Failed"
				}
				children = append(children, map[string]any{"name": name + "-run", "pipelineTaskName": name, "status": status})
				if kept&(1<<i) != 0 {
					selected = append(selected, name)
				}
			}
			run := map[string]any{"metadata": map[string]any{"name": "source"}, "spec": map[string]any{"pipelineRef": map[string]any{"name": "projection"}}, "status": map[string]any{"conditions": []any{map[string]any{"type": "Succeeded", "status": "False"}}, "pipelineSpec": spec, "childReferences": children}}
			write("run.json", run)
			trs := []any{}
			for i := 0; i < n; i++ {
				name := fmt.Sprintf("t%d", i)
				status := "True"
				if i == failed {
					status = "False"
				}
				trs = append(trs, map[string]any{"metadata": map[string]any{"name": name + "-run", "labels": map[string]any{"tekton.dev/pipelineRun": "source", "tekton.dev/pipelineTask": name}}, "status": map[string]any{"conditions": []any{map[string]any{"type": "Succeeded", "status": status}}, "results": []any{map[string]any{"name": "out", "value": "ok"}}}})
			}
			write("taskruns.json", map[string]any{"items": trs})
			manifest, err := CreateRecoveryRun(path("pipeline.json"), path("run.json"), path("policy.json"), path("taskruns.json"), "", selected, true)
			if err != nil {
				t.Fatalf("graph %d kept %d: %v", mask, kept, err)
			}
			var out map[string]any
			if err = yaml.Unmarshal(manifest, &out); err != nil {
				t.Fatal(err)
			}
			projected := make([][]bool, n)
			for i := range projected {
				projected[i] = make([]bool, n)
			}
			rawTasks := asList(asMap(asMap(out["spec"])["pipelineSpec"])["tasks"])
			if len(rawTasks) != len(selected) {
				t.Fatalf("retained count %d != %d", len(rawTasks), len(selected))
			}
			for _, raw := range rawTasks {
				task := asMap(raw)
				var j int
				fmt.Sscanf(task["name"].(string), "t%d", &j)
				encoded, _ := json.Marshal(task)
				for i := 0; i < n; i++ {
					name := fmt.Sprintf("t%d", i)
					edge := false
					for _, parent := range asList(task["runAfter"]) {
						edge = edge || parent == name
					}
					edge = edge || strings.Contains(string(encoded), "$(tasks."+name+".results.out)")
					if edge {
						if kept&(1<<i) == 0 {
							t.Fatalf("dangling reference")
						}
						projected[i][j] = true
					}
				}
			}
			actual := warshall(projected)
			for i := 0; i < n; i++ {
				for j := 0; j < n; j++ {
					if kept&(1<<i) != 0 && kept&(1<<j) != 0 && actual[i][j] != original[i][j] {
						t.Fatalf("graph=%d kept=%d pair=%d,%d got=%v want=%v", mask, kept, i, j, actual[i][j], original[i][j])
					}
				}
			}
			cases++
		}
	}
	t.Logf("Verified %d transformations: 1024 forward DAGs x 31 nonempty retained sets; mixed runAfter/params/when edges", cases)
}
func appendList(raw any, v any) []any {
	if raw == nil {
		return []any{v}
	}
	return append(raw.([]any), v)
}
