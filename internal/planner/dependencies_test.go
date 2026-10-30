package planner

import (
	"strings"
	"testing"
	"time"
)

func TestRetryWithDirectionRestriction(t *testing.T) {
	base := []PipelineTask{
		{Name: "a"}, {Name: "b"},
		{Name: "c", RunAfter: []string{"a"}},
		{Name: "f", RunAfter: []string{"c"}},
	}
	tests := []struct {
		name     string
		tasks    []PipelineTask
		rules    map[string]TaskRule
		selected []string
		want     string
		reason   string
	}{
		{name: "earlier task", rules: map[string]TaskRule{"f": {RetryWith: []string{"a"}}}, want: "recover"},
		{name: "parallel task", rules: map[string]TaskRule{"f": {RetryWith: []string{"b"}}}, want: "recover"},
		{name: "later task in unreachable entry", rules: map[string]TaskRule{"a": {RetryWith: []string{"c"}}}, want: "refuse", reason: `task "a" includes later task "c"`},
		{name: "indirect later task through parallel branch", rules: map[string]TaskRule{
			"f": {RetryWith: []string{"a"}}, "a": {RetryWith: []string{"b"}}, "b": {RetryWith: []string{"c"}},
		}, want: "refuse", reason: `task "a" includes later task "c"`},
		{name: "indirect violation with multiple selections", selected: []string{"b", "a"}, rules: map[string]TaskRule{
			"a": {RetryWith: []string{"b"}}, "b": {RetryWith: []string{"c"}},
		}, want: "refuse", reason: `task "a" includes later task "c"`},
		{name: "explicitly selected later task", selected: []string{"a", "c"}, want: "recover"},
		{name: "YAML order is not execution order", tasks: []PipelineTask{base[3], base[2], base[1], base[0]}, rules: map[string]TaskRule{
			"f": {RetryWith: []string{"a"}},
		}, want: "recover"},
		{name: "result-only ordering", tasks: []PipelineTask{
			{Name: "a"}, {Name: "b"}, {Name: "c", Params: []Param{{Name: "input", Value: "$(tasks.a.results.output)"}}}, {Name: "f"},
		}, rules: map[string]TaskRule{"a": {RetryWith: []string{"c"}}}, want: "refuse", reason: `task "a" includes later task "c"`},
		{name: "when result ordering", tasks: []PipelineTask{
			{Name: "a"}, {Name: "b"}, {Name: "c", When: []map[string]any{{"input": "$(tasks.a.results.output)", "operator": "in", "values": []any{"yes"}}}}, {Name: "f"},
		}, rules: map[string]TaskRule{"a": {RetryWith: []string{"c"}}}, want: "refuse", reason: `task "a" includes later task "c"`},
		{name: "embedded task reference", tasks: []PipelineTask{
			{Name: "a"}, {Name: "b"}, {Name: "c", TaskSpec: map[string]any{"steps": []any{map[string]any{"script": "echo $(tasks.a.results.output)"}}}}, {Name: "f"},
		}, rules: map[string]TaskRule{"a": {RetryWith: []string{"c"}}}, want: "refuse", reason: `task "a" includes later task "c"`},
		{name: "unresolved execution dependency", tasks: []PipelineTask{{Name: "a", RunAfter: []string{"missing"}}, {Name: "f"}}, want: "refuse", reason: `unknown task "missing"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Pipeline
			p.Metadata.Name = "direction"
			p.Spec.Tasks = tt.tasks
			if p.Spec.Tasks == nil {
				p.Spec.Tasks = base
			}
			var r PipelineRun
			r.Metadata.Name = "source"
			r.Spec.PipelineRef.Name = p.Metadata.Name
			for _, task := range p.Spec.Tasks {
				status := "Succeeded"
				if task.Name == "f" {
					status = "Failed"
				}
				r.Status.ChildReferences = append(r.Status.ChildReferences, ChildReference{PipelineTask: task.Name, Status: status})
			}
			var policy Policy
			policy.Metadata.Name = "policy"
			policy.Spec.PipelineRef = p.Metadata.Name
			policy.Spec.Tasks = tt.rules
			plan, err := BuildWithSelection(p, recorded(r, p), policy, nil, tt.selected, time.Now())
			if err != nil || plan.Decision != tt.want || !strings.Contains(plan.RefusalReason, tt.reason) {
				t.Fatalf("plan=%+v err=%v; want %s / %q", plan, err, tt.want, tt.reason)
			}
			if tt.name == "explicitly selected later task" {
				for _, task := range plan.Tasks {
					if (task.Name == "a" || task.Name == "c" || task.Name == "f") && task.Action != "rerun" {
						t.Fatalf("selected/failed task not rerun: %+v", task)
					}
				}
			}
		})
	}
}
