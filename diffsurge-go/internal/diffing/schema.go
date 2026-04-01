package diffing

import (
	"context"
	"encoding/json"
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

func (sc *SchemaComparer) CompareContents(oldContent, newContent interface{}) ([]Diff, []BreakingChange, error) {
	oldSpec, err := loadOpenAPISpecFromContent(oldContent)
	if err != nil {
		return nil, nil, fmt.Errorf("loading old schema content: %w", err)
	}

	newSpec, err := loadOpenAPISpecFromContent(newContent)
	if err != nil {
		return nil, nil, fmt.Errorf("loading new schema content: %w", err)
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
	sc.compareRequestBody(path, oldOp.RequestBody, newOp.RequestBody, diffs, breaking)
}

func (sc *SchemaComparer) compareRequestBody(path string, oldBody, newBody *openapi3.RequestBodyRef, diffs *[]Diff, breaking *[]BreakingChange) {
	if oldBody == nil && newBody == nil {
		return
	}

	reqPath := path + ".requestBody"

	// Request body added where there was none
	if oldBody == nil && newBody != nil {
		if newBody.Value != nil && newBody.Value.Required {
			*breaking = append(*breaking, BreakingChange{
				Path:        reqPath,
				Type:        "request_body_added_required",
				Description: "A required request body has been added",
				Severity:    SeverityBreaking,
			})
		} else {
			*diffs = append(*diffs, Diff{
				Path:     reqPath,
				Type:     DiffTypeAdded,
				NewValue: "requestBody",
				Severity: SeverityInfo,
			})
		}
		return
	}

	// Request body removed
	if oldBody != nil && newBody == nil {
		*diffs = append(*diffs, Diff{
			Path:     reqPath,
			Type:     DiffTypeRemoved,
			OldValue: "requestBody",
			Severity: SeverityInfo,
		})
		return
	}

	// Both exist — compare request body schemas across content types
	if oldBody.Value != nil && newBody.Value != nil {
		// Check if required status changed
		if !oldBody.Value.Required && newBody.Value.Required {
			*breaking = append(*breaking, BreakingChange{
				Path:        reqPath,
				Type:        "request_body_became_required",
				Description: "Request body changed from optional to required",
				Severity:    SeverityBreaking,
			})
		}

		if oldBody.Value.Content != nil && newBody.Value.Content != nil {
			for contentType, oldMediaType := range oldBody.Value.Content {
				newMediaType := newBody.Value.Content[contentType]
				if newMediaType == nil {
					continue
				}
				if oldMediaType.Schema != nil && newMediaType.Schema != nil &&
					oldMediaType.Schema.Value != nil && newMediaType.Schema.Value != nil {
					schemaPath := fmt.Sprintf("%s.content.%s.schema", reqPath, contentType)
					sc.compareSchemaObject(schemaPath, oldMediaType.Schema.Value, newMediaType.Schema.Value, diffs, breaking)
				}
			}
		}
	}
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

	// Detect new required parameters (breaking) and new optional parameters (info)
	for name, newParam := range newMap {
		paramPath := fmt.Sprintf("%s.params.%s", path, name)
		oldParam, exists := oldMap[name]
		if !exists {
			if newParam.Required {
				*breaking = append(*breaking, BreakingChange{
					Path:        paramPath,
					Type:        "required_param_added",
					Description: fmt.Sprintf("New required parameter '%s' added", name),
					Severity:    SeverityBreaking,
				})
			} else {
				*diffs = append(*diffs, Diff{
					Path:     paramPath,
					Type:     DiffTypeAdded,
					NewValue: name,
					Severity: SeverityInfo,
				})
			}
			continue
		}

		// Parameter exists in both — check for breaking changes within it
		// Check if parameter became required
		if !oldParam.Required && newParam.Required {
			*breaking = append(*breaking, BreakingChange{
				Path:        paramPath,
				Type:        "param_became_required",
				Description: fmt.Sprintf("Parameter '%s' changed from optional to required", name),
				Severity:    SeverityBreaking,
			})
		}

		// Check parameter type change
		if oldParam.Schema != nil && newParam.Schema != nil &&
			oldParam.Schema.Value != nil && newParam.Schema.Value != nil {
			oldType := oldParam.Schema.Value.Type
			newType := newParam.Schema.Value.Type
			if oldType != nil && newType != nil && oldType.Slice()[0] != newType.Slice()[0] {
				*breaking = append(*breaking, BreakingChange{
					Path:        paramPath,
					Type:        "param_type_changed",
					Description: fmt.Sprintf("Parameter '%s' type changed from '%s' to '%s'", name, oldType.Slice()[0], newType.Slice()[0]),
					Severity:    SeverityBreaking,
				})
			}
		}
	}

	// Detect removed parameters (breaking)
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
			sc.compareResponseContent(respPath, oldResp.Value, newRespRef.Value, diffs, breaking)
		}
	}
}

func (sc *SchemaComparer) compareResponseContent(path string, oldResp, newResp *openapi3.Response, diffs *[]Diff, breaking *[]BreakingChange) {
	// Compare response description changes
	oldDesc := ""
	newDesc := ""
	if oldResp.Description != nil {
		oldDesc = *oldResp.Description
	}
	if newResp.Description != nil {
		newDesc = *newResp.Description
	}
	if oldDesc != newDesc {
		*diffs = append(*diffs, Diff{
			Path:     path + ".description",
			Type:     DiffTypeModified,
			OldValue: oldDesc,
			NewValue: newDesc,
			Severity: SeverityInfo,
		})
	}

	// Compare response body schemas across content types
	if oldResp.Content != nil && newResp.Content != nil {
		for contentType, oldMediaType := range oldResp.Content {
			newMediaType := newResp.Content[contentType]
			if newMediaType == nil {
				*breaking = append(*breaking, BreakingChange{
					Path:        fmt.Sprintf("%s.content.%s", path, contentType),
					Type:        "response_content_type_removed",
					Description: fmt.Sprintf("Response content type '%s' has been removed", contentType),
					Severity:    SeverityBreaking,
				})
				continue
			}
			// Compare the response body schema
			if oldMediaType.Schema != nil && newMediaType.Schema != nil &&
				oldMediaType.Schema.Value != nil && newMediaType.Schema.Value != nil {
				schemaPath := fmt.Sprintf("%s.content.%s.schema", path, contentType)
				sc.compareSchemaObject(schemaPath, oldMediaType.Schema.Value, newMediaType.Schema.Value, diffs, breaking)
			}
		}
	}
}

func (sc *SchemaComparer) compareSchemas(oldSpec, newSpec *openapi3.T, diffs *[]Diff, breaking *[]BreakingChange) {
	if oldSpec.Components == nil || newSpec.Components == nil {
		return
	}

	oldSchemas := oldSpec.Components.Schemas
	newSchemas := newSpec.Components.Schemas

	// Check for removed schemas
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

	// Check for added schemas
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

	// Deep compare schemas that exist in both versions
	for name, oldSchemaRef := range oldSchemas {
		newSchemaRef, exists := newSchemas[name]
		if !exists {
			continue // already handled above
		}
		if oldSchemaRef.Value == nil || newSchemaRef.Value == nil {
			continue
		}
		schemaPath := fmt.Sprintf("components.schemas.%s", name)
		sc.compareSchemaObject(schemaPath, oldSchemaRef.Value, newSchemaRef.Value, diffs, breaking)
	}
}

// compareSchemaObject performs deep comparison of two OpenAPI schema objects,
// detecting property-level breaking changes: type changes, required field
// additions/removals, property removals, and enum narrowing.
func (sc *SchemaComparer) compareSchemaObject(path string, oldSchema, newSchema *openapi3.Schema, diffs *[]Diff, breaking *[]BreakingChange) {
	// 1. Detect type changes
	if oldSchema.Type != nil && newSchema.Type != nil {
		oldTypes := oldSchema.Type.Slice()
		newTypes := newSchema.Type.Slice()
		if len(oldTypes) > 0 && len(newTypes) > 0 && oldTypes[0] != newTypes[0] {
			*breaking = append(*breaking, BreakingChange{
				Path:        path,
				Type:        "type_changed",
				Description: fmt.Sprintf("Type changed from '%s' to '%s'", oldTypes[0], newTypes[0]),
				Severity:    SeverityBreaking,
			})
			*diffs = append(*diffs, Diff{
				Path:     path + ".type",
				Type:     DiffTypeTypeChanged,
				OldValue: oldTypes[0],
				NewValue: newTypes[0],
				Severity: SeverityBreaking,
			})
		}
	}

	// 2. Detect required field changes
	oldRequired := toStringSet(oldSchema.Required)
	newRequired := toStringSet(newSchema.Required)

	// New required fields added (breaking if the property existed before as optional)
	for field := range newRequired {
		if !oldRequired[field] {
			reqPath := fmt.Sprintf("%s.required.%s", path, field)
			// If the property existed before but was optional, it's breaking
			// If the property is entirely new AND required, it's also breaking
			_, existedBefore := oldSchema.Properties[field]
			if existedBefore {
				*breaking = append(*breaking, BreakingChange{
					Path:        reqPath,
					Type:        "field_became_required",
					Description: fmt.Sprintf("Field '%s' changed from optional to required", field),
					Severity:    SeverityBreaking,
				})
			} else {
				*breaking = append(*breaking, BreakingChange{
					Path:        reqPath,
					Type:        "required_field_added",
					Description: fmt.Sprintf("New required field '%s' added", field),
					Severity:    SeverityBreaking,
				})
			}
		}
	}

	// Required fields removed (info — relaxing the contract is non-breaking)
	for field := range oldRequired {
		if !newRequired[field] {
			// Only report if the field still exists (just no longer required)
			if _, stillExists := newSchema.Properties[field]; stillExists {
				*diffs = append(*diffs, Diff{
					Path:     fmt.Sprintf("%s.required.%s", path, field),
					Type:     DiffTypeModified,
					OldValue: "required",
					NewValue: "optional",
					Severity: SeverityInfo,
				})
			}
		}
	}

	// 3. Detect property changes
	for propName, oldPropRef := range oldSchema.Properties {
		propPath := fmt.Sprintf("%s.properties.%s", path, propName)
		newPropRef, exists := newSchema.Properties[propName]
		if !exists {
			// Property removed
			severity := SeverityWarning
			changeType := "property_removed"
			desc := fmt.Sprintf("Property '%s' has been removed", propName)
			if oldRequired[propName] {
				severity = SeverityBreaking
				changeType = "required_property_removed"
				desc = fmt.Sprintf("Required property '%s' has been removed", propName)
			}
			*breaking = append(*breaking, BreakingChange{
				Path:        propPath,
				Type:        changeType,
				Description: desc,
				Severity:    severity,
			})
			*diffs = append(*diffs, Diff{
				Path:     propPath,
				Type:     DiffTypeRemoved,
				OldValue: propName,
				Severity: severity,
			})
			continue
		}

		// Property exists in both — recursively compare the property schemas
		if oldPropRef.Value != nil && newPropRef.Value != nil {
			sc.compareSchemaObject(propPath, oldPropRef.Value, newPropRef.Value, diffs, breaking)
		}
	}

	// Detect new properties added
	for propName := range newSchema.Properties {
		propPath := fmt.Sprintf("%s.properties.%s", path, propName)
		if _, exists := oldSchema.Properties[propName]; !exists {
			*diffs = append(*diffs, Diff{
				Path:     propPath,
				Type:     DiffTypeAdded,
				NewValue: propName,
				Severity: SeverityInfo,
			})
		}
	}

	// 4. Detect enum changes
	if len(oldSchema.Enum) > 0 || len(newSchema.Enum) > 0 {
		sc.compareEnums(path, oldSchema.Enum, newSchema.Enum, diffs, breaking)
	}

	// 5. Compare items schema (for array types)
	if oldSchema.Items != nil && newSchema.Items != nil &&
		oldSchema.Items.Value != nil && newSchema.Items.Value != nil {
		sc.compareSchemaObject(path+".items", oldSchema.Items.Value, newSchema.Items.Value, diffs, breaking)
	}
}

// compareEnums detects added and removed enum values.
// Removing an enum value is breaking (narrowing the contract).
// Adding an enum value is non-breaking (widening the contract).
func (sc *SchemaComparer) compareEnums(path string, oldEnum, newEnum []interface{}, diffs *[]Diff, breaking *[]BreakingChange) {
	oldSet := make(map[string]bool)
	for _, v := range oldEnum {
		oldSet[fmt.Sprintf("%v", v)] = true
	}
	newSet := make(map[string]bool)
	for _, v := range newEnum {
		newSet[fmt.Sprintf("%v", v)] = true
	}

	for val := range oldSet {
		if !newSet[val] {
			*breaking = append(*breaking, BreakingChange{
				Path:        fmt.Sprintf("%s.enum", path),
				Type:        "enum_value_removed",
				Description: fmt.Sprintf("Enum value '%s' has been removed", val),
				Severity:    SeverityBreaking,
			})
		}
	}

	for val := range newSet {
		if !oldSet[val] {
			*diffs = append(*diffs, Diff{
				Path:     fmt.Sprintf("%s.enum", path),
				Type:     DiffTypeAdded,
				NewValue: val,
				Severity: SeverityInfo,
			})
		}
	}
}

// toStringSet converts a slice of strings to a set (map) for O(1) lookup.
func toStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

func loadOpenAPISpec(path string) (*openapi3.T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return loadOpenAPISpecFromData(data)
}

func loadOpenAPISpecFromContent(content interface{}) (*openapi3.T, error) {
	switch value := content.(type) {
	case string:
		return loadOpenAPISpecFromData([]byte(value))
	case []byte:
		return loadOpenAPISpecFromData(value)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("encoding schema content: %w", err)
		}
		return loadOpenAPISpecFromData(data)
	}
}

func loadOpenAPISpecFromData(data []byte) (*openapi3.T, error) {
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
