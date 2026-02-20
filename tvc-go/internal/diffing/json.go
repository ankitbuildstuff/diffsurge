package diffing

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadJSON(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	return ParseJSON(data)
}

func ParseJSON(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return result, nil
}

func DiffJSONFiles(oldPath, newPath string, config Config) ([]Diff, error) {
	oldData, err := LoadJSON(oldPath)
	if err != nil {
		return nil, fmt.Errorf("loading old file: %w", err)
	}

	newData, err := LoadJSON(newPath)
	if err != nil {
		return nil, fmt.Errorf("loading new file: %w", err)
	}

	engine := NewEngine(config)
	return engine.Compare(oldData, newData)
}

func DiffJSONBytes(oldData, newData []byte, config Config) ([]Diff, error) {
	oldParsed, err := ParseJSON(oldData)
	if err != nil {
		return nil, fmt.Errorf("parsing old JSON: %w", err)
	}

	newParsed, err := ParseJSON(newData)
	if err != nil {
		return nil, fmt.Errorf("parsing new JSON: %w", err)
	}

	engine := NewEngine(config)
	return engine.Compare(oldParsed, newParsed)
}
