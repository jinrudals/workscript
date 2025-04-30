package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jinrudals/workscript/builder"
)

type MergeCommand struct {
	StagesPath      string
	InformationPath string
	OutputPath      string
}

// Run executes the merge command logic
func (m *MergeCommand) Run() error {
	stages, info, err := m.Prepare()
	if err != nil {
		return fmt.Errorf("prepare error: %w", err)
	}

	// Set global values
	if globals, ok := info["global"].(map[string]interface{}); ok {
		if err := builder.GetInstance().Set(globals); err != nil {
			return fmt.Errorf("failed to set global values: %w", err)
		}
	}

	// Run build
	b, err := builder.NewBuilder(stages, info)
	if err != nil {
		return fmt.Errorf("builder init error: %w", err)
	}
	refined, err := b.Build()
	if err != nil {
		return fmt.Errorf("builder execution error: %w", err)
	}

	return m.Write(refined)
}

func (m *MergeCommand) Prepare() (map[string]map[string]interface{}, map[string]interface{}, error) {
	// Read stages
	stageBytes, err := os.ReadFile(m.StagesPath)
	if err != nil {
		return nil, nil, err
	}
	var stages map[string]map[string]interface{}
	if err := json.Unmarshal(stageBytes, &stages); err != nil {
		return nil, nil, err
	}

	// Read information
	infoBytes, err := os.ReadFile(m.InformationPath)
	if err != nil {
		return nil, nil, err
	}
	var info map[string]interface{}
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return nil, nil, err
	}

	return stages, info, nil
}

func (m *MergeCommand) Write(result map[string]map[string]interface{}) error {
	f, err := os.Create(m.OutputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(result); err != nil {
		return err
	}

	fmt.Printf("Merge complete → %s\n", m.OutputPath)
	return nil
}
