package diffing

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type DiffType string

const (
	DiffTypeAdded       DiffType = "added"
	DiffTypeRemoved     DiffType = "removed"
	DiffTypeModified    DiffType = "modified"
	DiffTypeTypeChanged DiffType = "type_changed"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityBreaking Severity = "breaking"
)

type Diff struct {
	Path     string      `json:"path"`
	Type     DiffType    `json:"type"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
	Severity Severity    `json:"severity"`
}

type Config struct {
	IgnorePaths      []string
	TreatArraysAsSet bool
}

type Engine struct {
	config Config
}

func NewEngine(config Config) *Engine {
	return &Engine{config: config}
}

func (e *Engine) Compare(a, b interface{}) ([]Diff, error) {
	var diffs []Diff
	e.compareRecursive("", a, b, &diffs)
	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Path < diffs[j].Path
	})
	return diffs, nil
}

func (e *Engine) compareRecursive(path string, a, b interface{}, diffs *[]Diff) {
	if a == nil && b == nil {
		return
	}

	if a == nil && b != nil {
		*diffs = append(*diffs, Diff{
			Path:     path,
			Type:     DiffTypeAdded,
			NewValue: b,
			Severity: SeverityInfo,
		})
		return
	}

	if a != nil && b == nil {
		*diffs = append(*diffs, Diff{
			Path:     path,
			Type:     DiffTypeRemoved,
			OldValue: a,
			Severity: SeverityWarning,
		})
		return
	}

	aKind := reflectKind(a)
	bKind := reflectKind(b)

	if aKind != bKind {
		*diffs = append(*diffs, Diff{
			Path:     path,
			Type:     DiffTypeTypeChanged,
			OldValue: a,
			NewValue: b,
			Severity: SeverityBreaking,
		})
		return
	}

	switch aVal := a.(type) {
	case map[string]interface{}:
		e.compareObjects(path, aVal, b.(map[string]interface{}), diffs)
	case []interface{}:
		e.compareArrays(path, aVal, b.([]interface{}), diffs)
	default:
		if !reflect.DeepEqual(a, b) {
			*diffs = append(*diffs, Diff{
				Path:     path,
				Type:     DiffTypeModified,
				OldValue: a,
				NewValue: b,
				Severity: SeverityInfo,
			})
		}
	}
}

func (e *Engine) compareObjects(path string, a, b map[string]interface{}, diffs *[]Diff) {
	keys := mergedKeys(a, b)

	for _, key := range keys {
		currentPath := buildPath(path, key)
		if e.shouldIgnore(currentPath) {
			continue
		}

		aVal, aExists := a[key]
		bVal, bExists := b[key]

		if !aExists {
			*diffs = append(*diffs, Diff{
				Path:     currentPath,
				Type:     DiffTypeAdded,
				NewValue: bVal,
				Severity: SeverityInfo,
			})
			continue
		}

		if !bExists {
			*diffs = append(*diffs, Diff{
				Path:     currentPath,
				Type:     DiffTypeRemoved,
				OldValue: aVal,
				Severity: SeverityWarning,
			})
			continue
		}

		e.compareRecursive(currentPath, aVal, bVal, diffs)
	}
}

func (e *Engine) compareArrays(path string, a, b []interface{}, diffs *[]Diff) {
	if e.config.TreatArraysAsSet {
		e.compareArraysAsSet(path, a, b, diffs)
		return
	}

	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	for i := 0; i < maxLen; i++ {
		currentPath := fmt.Sprintf("%s[%d]", path, i)

		if i >= len(a) {
			*diffs = append(*diffs, Diff{
				Path:     currentPath,
				Type:     DiffTypeAdded,
				NewValue: b[i],
				Severity: SeverityInfo,
			})
			continue
		}

		if i >= len(b) {
			*diffs = append(*diffs, Diff{
				Path:     currentPath,
				Type:     DiffTypeRemoved,
				OldValue: a[i],
				Severity: SeverityWarning,
			})
			continue
		}

		e.compareRecursive(currentPath, a[i], b[i], diffs)
	}
}

func (e *Engine) compareArraysAsSet(path string, a, b []interface{}, diffs *[]Diff) {
	aSet := make(map[string]interface{})
	bSet := make(map[string]interface{})

	for _, item := range a {
		aSet[fmt.Sprintf("%v", item)] = item
	}
	for _, item := range b {
		bSet[fmt.Sprintf("%v", item)] = item
	}

	for key, val := range aSet {
		if _, exists := bSet[key]; !exists {
			*diffs = append(*diffs, Diff{
				Path:     path + "[]",
				Type:     DiffTypeRemoved,
				OldValue: val,
				Severity: SeverityWarning,
			})
		}
	}

	for key, val := range bSet {
		if _, exists := aSet[key]; !exists {
			*diffs = append(*diffs, Diff{
				Path:     path + "[]",
				Type:     DiffTypeAdded,
				NewValue: val,
				Severity: SeverityInfo,
			})
		}
	}
}

func (e *Engine) shouldIgnore(path string) bool {
	for _, p := range e.config.IgnorePaths {
		if p == path || strings.HasPrefix(path, p+".") || strings.HasPrefix(path, p+"[") {
			return true
		}
	}
	return false
}

func buildPath(base, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

func mergedKeys(a, b map[string]interface{}) []string {
	seen := make(map[string]bool)
	var keys []string

	for k := range a {
		if !seen[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	for k := range b {
		if !seen[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}

	sort.Strings(keys)
	return keys
}

func reflectKind(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	default:
		return reflect.TypeOf(v).String()
	}
}
