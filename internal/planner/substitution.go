package planner

import (
	"fmt"
	"reflect"
	"strings"
)

// Tekton saves status.pipelineSpec after substituting Pipeline parameters,
// pipelineRun/pipeline context and Workspace binding flags. A current string
// matches a saved one if it is equal, or equal after applying this run's own
// values for exactly those variables. Array and object values are not
// substituted, so expansions they cause still refuse.
func runSubstitutions(pipelineName string, r PipelineRun, recorded any) map[string]string {
	values := map[string]any{}
	spec, _ := recorded.(map[string]any)
	for _, raw := range asList(spec["params"]) {
		param := asMap(raw)
		name, _ := param["name"].(string)
		if value, ok := param["default"]; ok && value != nil {
			values[name] = value
		}
	}
	for _, param := range r.Spec.Params {
		values[param.Name] = param.Value
	}
	subs := map[string]string{}
	for name, value := range values {
		text, ok := value.(string)
		if !ok {
			continue
		}
		for _, form := range []string{"params.%s", "params['%s']", `params["%s"]`} {
			subs["$("+fmt.Sprintf(form, name)+")"] = text
		}
	}
	subs["$(context.pipelineRun.name)"] = r.Metadata.Name
	subs["$(context.pipelineRun.namespace)"] = r.Metadata.Namespace
	subs["$(context.pipelineRun.uid)"] = r.Metadata.UID
	subs["$(context.pipeline.name)"] = pipelineName
	bound := map[string]bool{}
	for _, binding := range r.Spec.Workspaces {
		if name, ok := binding["name"].(string); ok {
			bound[name] = true
		}
	}
	for _, raw := range asList(spec["workspaces"]) {
		name, _ := asMap(raw)["name"].(string)
		subs["$(workspaces."+name+".bound)"] = fmt.Sprint(bound[name])
	}
	return subs
}

func substitute(text string, subs map[string]string) string {
	if !strings.Contains(text, "$(") {
		return text
	}
	pairs := make([]string, 0, 2*len(subs))
	for variable, value := range subs {
		pairs = append(pairs, variable, value)
	}
	return strings.NewReplacer(pairs...).Replace(text)
}

func definitionMatchesUnder(current, recorded any, subs map[string]string) bool {
	switch typed := current.(type) {
	case map[string]any:
		other, ok := recorded.(map[string]any)
		if !ok || len(typed) != len(other) {
			return false
		}
		for key, value := range typed {
			if otherValue, exists := other[key]; !exists || !definitionMatchesUnder(value, otherValue, subs) {
				return false
			}
		}
		return true
	case []any:
		other, ok := recorded.([]any)
		if !ok || len(typed) != len(other) {
			return false
		}
		for index := range typed {
			if !definitionMatchesUnder(typed[index], other[index], subs) {
				return false
			}
		}
		return true
	case string:
		other, ok := recorded.(string)
		return ok && (typed == other || substitute(typed, subs) == other)
	default:
		return reflect.DeepEqual(current, recorded)
	}
}
