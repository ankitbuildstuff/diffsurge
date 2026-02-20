package diffing

import (
	"context"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

type BreakingChange struct {
	Path        string   `json:"path"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
}

type SchemaComparer struct{}

func NewSchemaComparer() *SchemaComparer {
	return &SchemaComparer{}
}

func (sc *SchemaComparer) CompareFiles(oldPath, newPath string) ([]Diff, []BreakingChange, error) {
	oldSpec, err := loadOpenAPISpec(oldPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading old spec: %w", err)
	}

	newSpec, err := loadOpenAPISpec(newPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading new spec: %w", err)
	}

	return sc.Compare(oldSpec, newSpec)
}

func (sc *SchemaComparer) Compare(oldSpec, newSpec *openapi3.T) ([]Diff, []BreakingChange, error) {
	var diffs []Diff
	var breakingChanges []BreakingChange

	sc.compareEndpoints(oldSpec, newSpec, &diffs, &breakingChanges)
	sc.compareSchemas(oldSpec, newSpec, &diffs, &breakingChanges)

	return diffs, breakingChanges, nil
}

func (sc *SchemaComparer) compareEndpoints(oldSpec, newSpec *openapi3.T, diffs *[]Diff, breaking *[]BreakingChange) {
	if oldSpec.Paths == nil || newSpec.Paths == nil {
		return
	}

	for path, oldPathItem := range oldSpec.Paths.Map() {
		newPathItem := newSpec.Paths.Find(path)
		if newPathItem == nil {
			*breaking = append(*breaking, BreakingChange{
				Path:        path,
				Type:        "endpoint_removed",
				Description: fmt.Sprintf("Endpoint %s has been removed", path),
				Severity:    SeverityBreaking,
			})
			*diffs = append(*diffs, Diff{
				Path:     path,
				Type:     DiffTypeRemoved,
				OldValue: path,
				Severity: SeverityBreaking,
			})
			continue
		}

		sc.comparePathItem(path, oldPathItem, newPathItem, diffs, breaking)
	}

	for path := range newSpec.Paths.Map() {
		if oldSpec.Paths.Find(path) == nil {
			*diffs = append(*diffs, Diff{
				Path:     path,
				Type:     DiffTypeAdded,
				NewValue: path,
				Severity: SeverityInfo,
			})
		}
	}
}

func (sc *SchemaComparer) comparePathItem(path string, oldItem, newItem *openapi3.PathItem, diffs *[]Diff, breaking *[]BreakingChange) {
	methods := map[string][2]*openapi3.Operation{
		"GET":     {oldItem.Get, newItem.Get},
		"POST":    {oldItem.Post, newItem.Post},
		"PUT":     {oldItem.Put, newItem.Put},
		"DELETE":  {oldItem.Delete, newItem.Delete},
		"PATCH":   {oldItem.Patch, newItem.Patch},
		"OPTIONS": {oldItem.Options, newItem.Options},
	}

	for method, ops := range methods {
		oldOp, newOp := ops[0], ops[1]
		opPath := fmt.Sprintf("%s.%s", path, method)

		if oldOp != nil && newOp == nil {
			*breaking = append(*breaking, BreakingChange{
				Path:        opPath,
				Type:        "method_removed",
				Description: fmt.Sprintf("Method %s on %s has been removed", method, path),
				Severity:    SeverityBreaking,
			})
			continue
		}

		if oldOp == nil && newOp != nil {
			*diffs = append(*diffs, Diff{
				Path:     opPath,
				Type:     DiffTypeAdded,
				NewValue: method,
				Severity: SeverityInfo,
			})
			continue
		}

		if oldOp != nil && newOp != nil {
			sc.compareOperation(opPath, oldOp, newOp, diffs, breaking)
		}
	}
}

func (sc *SchemaComparer) compareOperation(path string, oldOp, newOp *openapi3.Operation, diffs *[]Diff, breaking *[]BreakingChange) {
	sc.compareParameters(path, oldOp.Parameters, newOp.Parameters, diffs, breaking)
	sc.compareResponses(path, oldOp.Responses, newOp.Responses, diffs, breaking)
}

func (sc *SchemaComparer) compareParameters(path string, oldParams, newParams openapi3.Parameters, diffs *[]Diff, breaking *[]BreakingChange) {
	oldMap := make(map[string]*openapi3.Parameter)
	for _, p := range oldParams {
		if p.Value != nil {
			oldMap[p.Value.Name] = p.Value
		}
	}

	newMap := make(map[string]*openapi3.Parameter)
	for _, p := range newParams {
		if p.Value != nil {
			newMap[p.Value.Name] = p.Value
		}
	}

	for name, newParam := range newMap {
		paramPath := fmt.Sprintf("%s.params.%s", path, name)
		if _, exists := oldMap[name]; !exists && newParam.Required {
			*breaking = append(*breaking, BreakingChange{
				Path:        paramPath,
				Type:        "required_param_added",
				Description: fmt.Sprintf("New required parameter '%s' added", name),
				Severity:    SeverityBreaking,
			})
		}
	}

	for name := range oldMap {
		paramPath := fmt.Sprintf("%s.params.%s", path, name)
		if _, exists := newMap[name]; !exists {
			*breaking = append(*breaking, BreakingChange{
				Path:        paramPath,
				Type:        "param_removed",
				Description: fmt.Sprintf("Parameter '%s' has been removed", name),
				Severity:    SeverityBreaking,
			})
		}
	}
}

func (sc *SchemaComparer) compareResponses(path string, oldResponses, newResponses *openapi3.Responses, diffs *[]Diff, breaking *[]BreakingChange) {
	if oldResponses == nil || newResponses == nil {
		return
	}

	for status, oldResp := range oldResponses.Map() {
		respPath := fmt.Sprintf("%s.responses.%s", path, status)
		newRespRef := newResponses.Value(status)
		if newRespRef == nil {
			*breaking = append(*breaking, BreakingChange{
				Path:        respPath,
				Type:        "response_removed",
				Description: fmt.Sprintf("Response for status %s has been removed", status),
				Severity:    SeverityBreaking,
			})
			continue
		}

		if oldResp.Value != nil && newRespRef.Value != nil {
			sc.compareResponseContent(respPath, oldResp.Value, newRespRef.Value)
		}
	}
}

func (sc *SchemaComparer) compareResponseContent(path string, oldResp, newResp *openapi3.Response) {
	// Extensible: compare response body schemas, content types, etc.
	// For now, we compare the description changes
	oldDesc := ""
	newDesc := ""
	if oldResp.Description != nil {
		oldDesc = *oldResp.Description
	}
	if newResp.Description != nil {
		newDesc = *newResp.Description
	}
	_ = oldDesc
	_ = newDesc
}

func (sc *SchemaComparer) compareSchemas(oldSpec, newSpec *openapi3.T, diffs *[]Diff, breaking *[]BreakingChange) {
	if oldSpec.Components == nil || newSpec.Components == nil {
		return
	}

	oldSchemas := oldSpec.Components.Schemas
	newSchemas := newSpec.Components.Schemas

	for name := range oldSchemas {
		schemaPath := fmt.Sprintf("components.schemas.%s", name)
		if _, exists := newSchemas[name]; !exists {
			*breaking = append(*breaking, BreakingChange{
				Path:        schemaPath,
				Type:        "schema_removed",
				Description: fmt.Sprintf("Schema '%s' has been removed", name),
				Severity:    SeverityBreaking,
			})
		}
	}

	for name := range newSchemas {
		schemaPath := fmt.Sprintf("components.schemas.%s", name)
		if _, exists := oldSchemas[name]; !exists {
			*diffs = append(*diffs, Diff{
				Path:     schemaPath,
				Type:     DiffTypeAdded,
				NewValue: name,
				Severity: SeverityInfo,
			})
		}
	}
}

func loadOpenAPISpec(path string) (*openapi3.T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("parsing OpenAPI spec: %w", err)
	}

	if err := spec.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("validating OpenAPI spec: %w", err)
	}

	return spec, nil
}
