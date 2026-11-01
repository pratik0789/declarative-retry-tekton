package planner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
	"time"

	"sigs.k8s.io/yaml"
)

var artifactRef = regexp.MustCompile(`\$\(tasks\.([a-z0-9-]+)\.results\.([A-Z0-9_]*ARTIFACT)\)`)

// Offline study over Konflux container-build-catalog OCI trusted-artifact Pipelines;
// set CBC to a checkout and OUT to the output path. After the artifacts expire, an
// executing task that reads an artifact result of a reused producer gets a dangling
// reference.
func TestKonfluxExpiryStudy(t *testing.T) {
	root := os.Getenv("CBC")
	if root == "" {
		t.Skip("CBC not set")
	}
	files, _ := filepath.Glob(filepath.Join(root, "pipelines/*-oci-ta*/*.yaml"))
	type row struct {
		Pipeline, Failed, Model string
		Dangling                map[string]bool
		Reruns                  map[string]int
	}
	var out []row
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
		consumes := map[string]map[string]bool{}
		for _, field := range []string{"tasks", "finally"} {
			for _, raw := range asList(asMap(def)[field]) {
				m := asMap(raw)
				blob, _ := json.Marshal(m)
				for _, match := range artifactRef.FindAllStringSubmatch(string(blob), -1) {
					name := m["name"].(string)
					if consumes[name] == nil {
						consumes[name] = map[string]bool{}
					}
					consumes[name][match[1]] = true
				}
			}
		}
		up, err := executionDependencies(def)
		if err != nil {
			t.Fatal(err)
		}
		empty := Policy{}
		empty.Metadata.Name = "empty"
		empty.Spec.PipelineRef = pipeline.Metadata.Name
		for _, model := range []string{"upstream", "independent"} {
			for _, task := range pipeline.Spec.Tasks {
				run := syntheticRunK(pipeline, task.Name, model, up)
				completed := map[string]bool{}
				for _, c := range run.Status.ChildReferences {
					if c.Status == "Succeeded" {
						completed[c.PipelineTask] = true
					}
				}
				all := append(append([]PipelineTask(nil), pipeline.Spec.Tasks...), pipeline.Spec.Finally...)
				dangling := func(rerun map[string]bool) bool {
					for _, tk := range all {
						executes := rerun[tk.Name] || !completed[tk.Name]
						if !executes {
							continue
						}
						for producer := range consumes[tk.Name] {
							if completed[producer] && !rerun[producer] {
								return true
							}
						}
					}
					return false
				}
				r := row{pipeline.Metadata.Name, task.Name, model, map[string]bool{}, map[string]int{}}
				failedOnly := map[string]bool{task.Name: true}
				r.Dangling["failed-only"], r.Reruns["failed-only"] = dangling(failedOnly), 1
				anc := map[string]bool{task.Name: true}
				for a := range ancestorsK(task.Name, up) {
					if completed[a] {
						anc[a] = true
					}
				}
				r.Dangling["all-ancestors"], r.Reruns["all-ancestors"] = dangling(anc), len(anc)
				sel := map[string]bool{}
				var p Plan
				for {
					var s []string
					for k := range sel {
						s = append(s, k)
					}
					sort.Strings(s)
					p, err = BuildWithSelection(pipeline, run, empty, nil, s, time.Now())
					if err != nil || p.Decision != "recover" {
						t.Fatalf("%s %s: %v %s", f, task.Name, err, p.RefusalReason)
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
				r.Dangling["warning-rule"], r.Reruns["warning-rule"] = dangling(warn), len(warn)
				full := map[string]bool{}
				for c := range completed {
					full[c] = true
				}
				full[task.Name] = true
				r.Dangling["full-restart"], r.Reruns["full-restart"] = false, len(full)
				// retryWith derived from the documented artifact flows: each consumer retries with its producers.
				flow := empty
				flow.Metadata.Name = "artifact-flows"
				flow.Spec.Tasks = map[string]TaskRule{}
				for consumer, producers := range consumes {
					var list []string
					for producer := range producers {
						list = append(list, producer)
					}
					sort.Strings(list)
					flow.Spec.Tasks[consumer] = TaskRule{RetryWith: list}
				}
				pf, err := BuildWithSelection(pipeline, run, flow, nil, nil, time.Now())
				if err != nil || pf.Decision != "recover" {
					t.Fatalf("artifact-flow policy %s %s: %v %s", f, task.Name, err, pf.RefusalReason)
				}
				flowRerun := map[string]bool{}
				for _, tp := range pf.Tasks {
					if tp.Action == "rerun" {
						flowRerun[tp.Name] = true
					}
				}
				r.Dangling["artifact-flow-policy"], r.Reruns["artifact-flow-policy"] = dangling(flowRerun), len(flowRerun)
				// resumeWithin shorter than the artifact expiry refuses every recovery after expiry.
				policy := empty
				policy.Metadata.Name = "window"
				policy.Spec.ResumeWithin = "96h"
				late := run
				done := time.Now().Add(-6 * 24 * time.Hour)
				late.Status.CompletionTime = &done
				pw, _ := BuildWithSelection(pipeline, late, policy, nil, nil, time.Now())
				r.Dangling["resumeWithin-policy"], r.Reruns["resumeWithin-policy"] = false, 0
				if pw.Decision != "refuse" {
					t.Fatalf("expected window refusal: %s", pw.RefusalReason)
				}
				out = append(out, r)
			}
		}
	}
	data, _ := json.MarshalIndent(out, "", " ")
	os.WriteFile(os.Getenv("OUT"), data, 0o644)
	fmt.Println("rows", len(out))
}
