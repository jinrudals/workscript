package builder

import (
	"encoding/json"
	"fmt"

	"github.com/jinrudals/workscript/builder/utils"
)

type Builder struct {
	OriginalStages map[string]*Stage
	Targets        []*Target
	CombinedStages map[string]map[string]interface{}
}

func NewBuilder(
	stages map[string]map[string]interface{},
	info map[string]interface{}) (*Builder, error) {
	originalStgages := make(map[string]*Stage, len(stages))

	// step 1. Set gloval values
	if gvals, ok := info["global"].(map[string]interface{}); ok {
		if err := GetInstance().Set(gvals); err != nil {
			return nil, fmt.Errorf("failed to set global values: %w", err)
		}
	}

	for key, val := range stages {
		val["name"] = key
		originalStgages[key] = NewStage(val)
	}
	targets := []*Target{}
	if tmap, ok := info["targets"].(map[string]interface{}); ok {
		for name, raw := range tmap {
			tvals, ok := raw.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid target : %s", name)
			}
			stageCopies := make(map[string]*Stage)
			for k, v := range originalStgages {
				cp := *v
				cp.Variables = utils.CopyStringMap(v.Variables)
				cp.Run = utils.CopyStringMap(v.Run)
				cp.Post = utils.CopyStringMap(v.Post)
				cp.Before = append([]string{}, v.Before...)
				cp.After = append([]string{}, v.After...)
				cp.Requires = append([]string{}, v.Requires...)
				stageCopies[k] = &cp
			}
			target := NewTarget(name, stageCopies, tvals)
			targets = append(targets, target)
		}
	}
	return &Builder{
		OriginalStages: originalStgages,
		Targets:        targets,
		CombinedStages: make(map[string]map[string]interface{}),
	}, nil
}

func (b *Builder) Build() (map[string]map[string]interface{}, error) {
	outputs := []map[string]interface{}{}

	for _, target := range b.Targets {
		target.RemoveUnusedStages()
		target.AddRequiredStages()
		target.ApplyTargetSpecificVariables()
		target.ResolveStageVariables()
		target.ApplyCrossVariables()
		target.RemoveNotExistsDependencies()
		target.ChangeNames()
		target.ResolveCommandVariables()

		jsonBytes, err := target.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize target %s: %w", target.Name, err)
		}

		var stageOutputs []map[string]interface{}

		if err := json.Unmarshal(jsonBytes, &stageOutputs); err != nil {
			return nil, fmt.Errorf("json unmarshal error for target %s: %w", target.Name, err)
		}

		outputs = append(outputs, stageOutputs...)
	}

	refined := make(map[string]map[string]interface{}, len(outputs))
	for _, each := range outputs {
		if nameRaw, ok := each["name"]; ok {
			if nameStr, ok := nameRaw.(string); ok {
				refined[nameStr] = each
			}
		}
	}
	b.CombinedStages = refined
	return refined, nil
}
