package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jinrudals/workscript/runner"
)

// Run executes the DAG in "all" or "command" mode.
func Run(stagesPath string, only string, withDeps bool, maxWorkers int, mode string) error {
	data, err := os.ReadFile(stagesPath)
	if err != nil {
		return fmt.Errorf("failed to read merged file: %w", err)
	}

	var allStages map[string]map[string]interface{}
	if err := json.Unmarshal(data, &allStages); err != nil {
		return fmt.Errorf("failed to unmarshal stages: %w", err)
	}

	stages := allStages

	// If --only is used, extract subgraph
	if only != "" {
		if _, ok := allStages[only]; !ok {
			return fmt.Errorf("stage '%s' not found in merged file", only)
		}

		selected := map[string]struct{}{}

		if withDeps {
			var collectWithDeps func(name string)
			collectWithDeps = func(name string) {
				if _, ok := selected[name]; ok {
					return
				}
				selected[name] = struct{}{}
				if node, ok := allStages[name]; ok {
					if afterList, ok := node["after"].([]interface{}); ok {
						for _, a := range afterList {
							collectWithDeps(a.(string))
						}
					}
				}
			}
			collectWithDeps(only)
		}

		selected[only] = struct{}{}
		filtered := make(map[string]map[string]interface{})
		for k := range selected {
			filtered[k] = allStages[k]
		}

		// Clean up irrelevant before/after refs
		for _, v := range filtered {
			v["before"] = filterStrings(v["before"], selected)
			v["after"] = filterStrings(v["after"], selected)
		}

		stages = filtered
	}

	// Ensure logs directory
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("failed to create logs dir: %w", err)
	}

	r := runner.NewRunnerFromMap(stages)
	r.Launch(maxWorkers, mode)
	return nil
}

// Post runs the DAG in post-only mode
func Post(stagesPath string, only string, withDeps bool, maxWorkers int) error {
	return Run(stagesPath, only, withDeps, maxWorkers, "post")
}

// filterStrings filters []interface{} by keys present in 'allowed'
func filterStrings(raw interface{}, allowed map[string]struct{}) []string {
	var result []string
	list, ok := raw.([]interface{})
	if !ok {
		return result
	}
	for _, val := range list {
		str := fmt.Sprintf("%v", val)
		if _, ok := allowed[str]; ok {
			result = append(result, str)
		}
	}
	return result
}
