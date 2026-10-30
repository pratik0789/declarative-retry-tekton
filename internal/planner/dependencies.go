package planner

import (
	"encoding/json"
	"fmt"
	"sort"
)

// executionDependencies reads the complete task definitions, including result
// references in conditions and embedded Tasks. YAML list order is not an edge.
func executionDependencies(spec any) (map[string][]string, error) {
	upstream := map[string][]string{}
	ordinary := asList(asMap(spec)["tasks"])
	final := asList(asMap(spec)["finally"])
	for index, raw := range append(append([]any(nil), ordinary...), final...) {
		task := asMap(raw)
		name, _ := task["name"].(string)
		if name == "" {
			return nil, fmt.Errorf("Pipeline task has no name")
		}
		if _, exists := upstream[name]; exists {
			return nil, fmt.Errorf("duplicate pipeline task %q", name)
		}
		parents := map[string]bool{}
		// Tekton schedules the final phase after all ordinary tasks finish.
		if index >= len(ordinary) {
			for _, predecessor := range ordinary {
				parent, _ := asMap(predecessor)["name"].(string)
				parents[parent] = true
			}
		}
		for _, dependency := range asList(task["runAfter"]) {
			parents[fmt.Sprint(dependency)] = true
		}
		encoded, err := json.Marshal(task)
		if err != nil {
			return nil, fmt.Errorf("read execution dependencies of %q: %w", name, err)
		}
		for _, match := range taskReference.FindAllStringSubmatch(string(encoded), -1) {
			parents[match[1]] = true
		}
		upstream[name] = nil
		for parent := range parents {
			upstream[name] = append(upstream[name], parent)
		}
		sort.Strings(upstream[name])
	}
	state := map[string]int{}
	var visit func(string) error
	visit = func(name string) error {
		if state[name] == 1 {
			return fmt.Errorf("Pipeline execution dependencies contain a cycle at %q", name)
		}
		if state[name] == 2 {
			return nil
		}
		state[name] = 1
		for _, parent := range upstream[name] {
			if _, exists := upstream[parent]; !exists {
				return fmt.Errorf("task %q has an execution dependency on unknown task %q", name, parent)
			}
			if err := visit(parent); err != nil {
				return err
			}
		}
		state[name] = 2
		return nil
	}
	for _, name := range sortedTaskNames(upstream) {
		if err := visit(name); err != nil {
			return nil, err
		}
	}
	return upstream, nil
}

func sortedTaskNames(upstream map[string][]string) []string {
	names := make([]string, 0, len(upstream))
	for name := range upstream {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// validateRetryDirections checks each policy root separately. Checking only
// adjacent declarations would allow parallel branches to introduce a descendant
// of the original root indirectly.
func validateRetryDirections(upstream map[string][]string, rules map[string]TaskRule) error {
	downstream := map[string][]string{}
	for _, name := range sortedTaskNames(upstream) {
		for _, parent := range upstream[name] {
			downstream[parent] = append(downstream[parent], name)
		}
	}
	for _, root := range sortedTaskNames(upstream) {
		if _, declared := rules[root]; !declared {
			continue
		}
		later := map[string]bool{}
		var markLater func(string)
		markLater = func(name string) {
			for _, child := range downstream[name] {
				if !later[child] {
					later[child] = true
					markLater(child)
				}
			}
		}
		markLater(root)
		seen := map[string]bool{}
		var visit func(string) error
		visit = func(name string) error {
			if seen[name] {
				return nil
			}
			seen[name] = true
			if later[name] {
				return fmt.Errorf("retryWith closure for task %q includes later task %q; only earlier or parallel tasks are allowed", root, name)
			}
			next := append([]string(nil), rules[name].RetryWith...)
			sort.Strings(next)
			for _, dependency := range next {
				if err := visit(dependency); err != nil {
					return err
				}
			}
			return nil
		}
		if err := visit(root); err != nil {
			return err
		}
	}
	return nil
}
