package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"sigs.k8s.io/yaml"
)

var unsafeTasks = map[string]bool{"create-github-release": true, "create-pyxis-image": true, "make-repo-public": true, "publish-pyxis-repository": true, "publish-to-mrrc": true, "push-artifacts-to-storage": true, "push-rpm-data-to-pyxis": true, "push-rpms-to-pulp": true, "push-snapshot": true, "rh-sign-image-cosign": true}

func taskFile(raw map[string]any) string {
	ref, _ := raw["taskRef"].(map[string]any)
	params, _ := ref["params"].([]any)
	for _, p := range params {
		pm := p.(map[string]any)
		if pm["name"] == "pathInRepo" {
			return filepath.Base(filepath.Dir(pm["value"].(string)))
		}
	}
	return ""
}

type cell struct {
	Pipeline, Failed, Model string
	Completed               int
	Rules                   map[string]map[string]any
}

// Offline study over Konflux release-service-catalog managed Pipelines; set RSC
// to a checkout and OUT to the output path. Labels are the ten tasks marked
// retry-unsafe in konflux-ci/release-service-utils pull request 1109.
func TestKonfluxReleaseStudy(t *testing.T) {
	root := os.Getenv("RSC")
	if root == "" {
		t.Skip("RSC not set")
	}
	files, _ := filepath.Glob(filepath.Join(root, "pipelines/managed/*/*.yaml"))
	var out []cell
	for _, f := range files {
		data, _ := os.ReadFile(f)
		var pipeline Pipeline
		if err := yaml.Unmarshal(data, &pipeline); err != nil || len(pipeline.Spec.Tasks) == 0 {
			continue
		}
		def := pipeline.definition
		if def == nil {
			def = toAny(pipeline.Spec)
		}
		unsafe := map[string]bool{}
		for _, field := range []string{"tasks", "finally"} {
			for _, raw := range asList(asMap(def)[field]) {
				m := asMap(raw)
				if unsafeTasks[taskFile(m)] {
					unsafe[m["name"].(string)] = true
				}
			}
		}
		if len(unsafe) == 0 {
			continue
		}
		policy := Policy{}
		policy.Metadata.Name = "labels"
		policy.Spec.PipelineRef = pipeline.Metadata.Name
		policy.Spec.Tasks = map[string]TaskRule{}
		for name := range unsafe {
			policy.Spec.Tasks[name] = TaskRule{BlocksResume: true, NoRepeat: true}
		}
		empty := Policy{}
		empty.Metadata.Name = "empty"
		empty.Spec.PipelineRef = pipeline.Metadata.Name
		up, err := executionDependencies(def)
		if err != nil {
			t.Fatal(err)
		}
		for _, model := range []string{"upstream", "independent"} {
			for _, task := range pipeline.Spec.Tasks {
				run := syntheticRunK(pipeline, task.Name, model, up)
				completed := map[string]bool{}
				for _, c := range run.Status.ChildReferences {
					if c.Status == "Succeeded" {
						completed[c.PipelineTask] = true
					}
				}
				anc := ancestorsK(task.Name, up)
				rules := map[string]map[string]any{}
				score := func(name string, reruns map[string]bool, refused bool) {
					risk := 0
					for r := range reruns {
						if unsafe[r] {
							risk++
						}
					}
					rules[name] = map[string]any{"reruns": len(reruns), "unsafeReexecuted": risk, "refused": refused}
				}
				full := map[string]bool{task.Name: true}
				for c := range completed {
					full[c] = true
				}
				score("full-restart", full, false)
				ancestors := map[string]bool{task.Name: true}
				for a := range anc {
					if completed[a] {
						ancestors[a] = true
					}
				}
				score("all-ancestors", ancestors, false)
				score("failed-only", map[string]bool{task.Name: true}, false)
				sel := map[string]bool{}
				var p Plan
				for {
					var s []string
					for k := range sel {
						s = append(s, k)
					}
					sort.Strings(s)
					p, err = BuildWithSelection(pipeline, run, empty, nil, s, time.Now())
					if err != nil {
						t.Fatal(err)
					}
					if p.Decision != "recover" {
						t.Fatalf("%s %s: %s", f, task.Name, p.RefusalReason)
					}
					inh := map[string]bool{}
					for _, tp := range p.Tasks {
						if tp.Action == "inherit" {
							inh[tp.Name] = true
						}
					}
					added := false
					for _, w := range p.Warnings {
						for _, x := range w.Tasks {
							if inh[x] && !sel[x] {
								sel[x], added = true, true
							}
						}
					}
					if !added {
						break
					}
				}
				warn := map[string]bool{}
				for _, tp := range p.Tasks {
					if tp.Action == "rerun" {
						warn[tp.Name] = true
					}
				}
				score("warning-rule", warn, false)
				pp, err := BuildWithSelection(pipeline, run, policy, nil, nil, time.Now())
				if err != nil {
					t.Fatal(err)
				}
				decl := map[string]bool{}
				for _, tp := range pp.Tasks {
					if tp.Action == "rerun" {
						decl[tp.Name] = true
					}
				}
				if pp.Decision == "refuse" {
					decl = map[string]bool{}
				}
				score("label-policy", decl, pp.Decision == "refuse")
				out = append(out, cell{pipeline.Metadata.Name, task.Name, model, len(completed), rules})
			}
		}
	}
	data, _ := json.MarshalIndent(out, "", " ")
	os.WriteFile(os.Getenv("OUT"), data, 0o644)
	fmt.Println("cells", len(out))
}

func ancestorsK(name string, up map[string][]string) map[string]bool {
	seen := map[string]bool{}
	var walk func(string)
	walk = func(c string) {
		for _, p := range up[c] {
			if !seen[p] {
				seen[p] = true
				walk(p)
			}
		}
	}
	walk(name)
	return seen
}

func syntheticRunK(pipeline Pipeline, failed, model string, up map[string][]string) PipelineRun {
	var workspaces []map[string]any
	for _, w := range pipeline.Spec.Workspaces {
		workspaces = append(workspaces, map[string]any{"name": w.Name, "persistentVolumeClaim": map[string]any{"claimName": w.Name}})
	}
	completed := ancestorsK(failed, up)
	if model == "independent" {
		down := map[string]bool{}
		for ch := true; ch; {
			ch = false
			for task, ps := range up {
				if down[task] {
					continue
				}
				for _, p := range ps {
					if p == failed || down[p] {
						down[task], ch = true, true
						break
					}
				}
			}
		}
		completed = map[string]bool{}
		for _, task := range pipeline.Spec.Tasks {
			if task.Name != failed && !down[task.Name] {
				completed[task.Name] = true
			}
		}
	}
	var run PipelineRun
	run.Metadata.Name = "replay-" + failed
	run.Status.Conditions = []Condition{{Type: "Succeeded", Status: "False"}}
	run.Spec.PipelineRef.Name = pipeline.Metadata.Name
	run.Spec.Workspaces = workspaces
	run.recordedDefinition = pipeline.definition
	for _, task := range pipeline.Spec.Tasks {
		if completed[task.Name] {
			run.Status.ChildReferences = append(run.Status.ChildReferences, ChildReference{Name: run.Metadata.Name + "-" + task.Name, PipelineTask: task.Name, Status: "Succeeded"})
		}
	}
	run.Status.ChildReferences = append(run.Status.ChildReferences, ChildReference{Name: run.Metadata.Name + "-" + failed, PipelineTask: failed, Status: "Failed"})
	return run
}
