package planner

import "testing"

func TestDefinitionMatchesUnderRunSubstitutions(t *testing.T) {
	current := map[string]any{
		"params":     []any{map[string]any{"name": "release"}, map[string]any{"name": "channel", "default": "stable"}},
		"workspaces": []any{map[string]any{"name": "shared"}},
		"tasks": []any{map[string]any{
			"name":     "scan",
			"params":   []any{map[string]any{"name": "run", "value": "$(context.pipelineRun.name)"}, map[string]any{"name": "tag", "value": "$(params.release)-$(params['channel'])"}},
			"taskSpec": map[string]any{"steps": []any{map[string]any{"script": "echo $(params.image) $(workspaces.shared.bound)"}}},
		}},
	}
	saved := map[string]any{
		"params":     current["params"],
		"workspaces": current["workspaces"],
		"tasks": []any{map[string]any{
			"name":     "scan",
			"params":   []any{map[string]any{"name": "run", "value": "src-run"}, map[string]any{"name": "tag", "value": "v1-stable"}},
			"taskSpec": map[string]any{"steps": []any{map[string]any{"script": "echo $(params.image) true"}}},
		}},
	}
	var run PipelineRun
	run.Metadata.Name = "src-run"
	run.Spec.Params = []Param{{Name: "release", Value: "v1"}}
	run.Spec.Workspaces = []map[string]any{{"name": "shared"}}
	subs := runSubstitutions("p", run, saved)
	if !definitionMatchesUnder(current, saved, subs) {
		t.Fatal("substituted parameters, context and Workspace flags should match")
	}
	run.Spec.Params = []Param{{Name: "release", Value: "v2"}}
	if definitionMatchesUnder(current, saved, runSubstitutions("p", run, saved)) {
		t.Fatal("a different parameter value must not match")
	}
	run.Spec.Params = []Param{{Name: "release", Value: "v1"}}
	saved["tasks"].([]any)[0].(map[string]any)["taskSpec"] = map[string]any{"steps": []any{map[string]any{"script": "echo changed true"}}}
	if definitionMatchesUnder(current, saved, subs) {
		t.Fatal("an edited script must not match")
	}
	array := map[string]any{"args": []any{"$(params.flags[*])"}}
	expanded := map[string]any{"args": []any{"-a", "-b"}}
	run.Spec.Params = []Param{{Name: "flags", Value: []any{"-a", "-b"}}}
	if definitionMatchesUnder(array, expanded, runSubstitutions("p", run, map[string]any{})) {
		t.Fatal("array expansion is not substituted and must refuse")
	}
}
