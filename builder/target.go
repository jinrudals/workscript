package builder

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jinrudals/workscript/builder/utils"
)

type Target struct {
	Name           string
	OriginalStages map[string]*Stage
	Stages         map[string]*Stage
	Kwargs         map[string]interface{}
}

func NewTarget(
	name string,
	stages map[string]*Stage,
	kwargs map[string]interface{},
) *Target {
	return &Target{
		Name:           name,
		OriginalStages: stages,
		Kwargs:         kwargs,
		Stages:         map[string]*Stage{},
	}
}

func (t *Target) RemoveUnusedStages() {
	gv := GetInstance()
	for k, v := range t.OriginalStages {
		if enabled, ok := t.Kwargs[k]; ok {
			if b, ok := enabled.(bool); ok && b {
				t.Stages[k] = v
			}
		} else if b, ok := gv.run[k]; ok && b {
			t.Stages[k] = v
		}
	}
}

func (t *Target) AddRequiredStages() {
	for _, stage := range t.Stages {
		for _, req := range stage.Requires {
			if _, ok := t.Stages[req]; !ok {
				if original, exists := t.OriginalStages[req]; exists {
					t.Stages[req] = original
				}
			}
		}
	}
}

func (t *Target) ApplyTargetSpecificVariables() {
	for key, value := range t.Kwargs {
		valStr := fmt.Sprintf("%v", value)
		if strings.Contains(key, ":") {
			parts := strings.SplitN(key, ":", 2)
			stageName, envName := parts[0], parts[1]
			if stage, ok := t.Stages[stageName]; ok {
				stage.Variables[envName] = valStr
			}
		} else {
			for _, stage := range t.Stages {
				stage.Variables[key] = valStr
			}
		}
	}
}

// ApplyCrossVariables resolves @{stage.var} references
func (t *Target) ApplyCrossVariables() {
	for _, stage := range t.Stages {
		if err := stage.ApplyCrossVariables(t.Stages); err != nil {
			log.Printf("cross-variable error in %s: %v", stage.Name, err)
		}
	}
}

// ChangeNames scopes stage names with target prefix
func (t *Target) ChangeNames() {
	for _, stage := range t.Stages {
		stage.SetName(t.Name)
	}
}

// ResolveStageVariables resolves @{var} within each stage
func (t *Target) ResolveStageVariables() {
	for _, stage := range t.Stages {
		stage.ResolveSelfVariables()
	}
}

// ResolveCommandVariables replaces ${var} in commands
func (t *Target) ResolveCommandVariables() {
	for _, stage := range t.Stages {
		if err := stage.ResolveCommandVariables(); err != nil {
			log.Printf("command-variable error in %s: %v", stage.Name, err)
		}
	}
}

// RemoveNotExistsDependencies removes dangling before/after edges
func (t *Target) RemoveNotExistsDependencies() {
	tKeys := make(map[string]*Stage, len(t.Stages))
	for k, v := range t.Stages {
		tKeys[k] = v
	}
	for _, stage := range t.Stages {
		stage.Before = utils.FilterExisting(stage.Before, tKeys)
		stage.After = utils.FilterExisting(stage.After, tKeys)
	}
}

// ToJSON marshals all final stage data with target name
func (t *Target) ToJSON() ([]byte, error) {
	var outputs []map[string]interface{}
	for _, stage := range t.Stages {
		data, err := stage.ToJSON()
		if err != nil {
			return nil, err
		}
		var out map[string]interface{}
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		out["target"] = t.Name
		outputs = append(outputs, out)
	}
	return json.Marshal(outputs)
}

func (t *Target) String() string {
	return fmt.Sprintf("Target(%s)", t.Name)
}
