package diffing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_Compare_SimpleFieldChange(t *testing.T) {
	a := map[string]interface{}{
		"name": "John",
		"age":  float64(30),
	}
	b := map[string]interface{}{
		"name": "Jane",
		"age":  float64(30),
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "name", diffs[0].Path)
	assert.Equal(t, DiffTypeModified, diffs[0].Type)
	assert.Equal(t, "John", diffs[0].OldValue)
	assert.Equal(t, "Jane", diffs[0].NewValue)
}

func TestEngine_Compare_FieldAdded(t *testing.T) {
	a := map[string]interface{}{
		"name": "John",
	}
	b := map[string]interface{}{
		"name":  "John",
		"email": "john@test.com",
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "email", diffs[0].Path)
	assert.Equal(t, DiffTypeAdded, diffs[0].Type)
	assert.Equal(t, "john@test.com", diffs[0].NewValue)
}

func TestEngine_Compare_FieldRemoved(t *testing.T) {
	a := map[string]interface{}{
		"name":  "John",
		"email": "john@test.com",
	}
	b := map[string]interface{}{
		"name": "John",
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "email", diffs[0].Path)
	assert.Equal(t, DiffTypeRemoved, diffs[0].Type)
	assert.Equal(t, "john@test.com", diffs[0].OldValue)
}

func TestEngine_Compare_NestedChange(t *testing.T) {
	a := map[string]interface{}{
		"user": map[string]interface{}{
			"address": map[string]interface{}{
				"city": "New York",
			},
		},
	}
	b := map[string]interface{}{
		"user": map[string]interface{}{
			"address": map[string]interface{}{
				"city": "San Francisco",
			},
		},
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "user.address.city", diffs[0].Path)
	assert.Equal(t, DiffTypeModified, diffs[0].Type)
}

func TestEngine_Compare_TypeChange(t *testing.T) {
	a := map[string]interface{}{
		"count": "42",
	}
	b := map[string]interface{}{
		"count": float64(42),
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "count", diffs[0].Path)
	assert.Equal(t, DiffTypeTypeChanged, diffs[0].Type)
	assert.Equal(t, SeverityBreaking, diffs[0].Severity)
}

func TestEngine_Compare_ArrayOrdered(t *testing.T) {
	a := map[string]interface{}{
		"tags": []interface{}{"go", "rust", "python"},
	}
	b := map[string]interface{}{
		"tags": []interface{}{"go", "java", "python"},
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "tags[1]", diffs[0].Path)
	assert.Equal(t, DiffTypeModified, diffs[0].Type)
	assert.Equal(t, "rust", diffs[0].OldValue)
	assert.Equal(t, "java", diffs[0].NewValue)
}

func TestEngine_Compare_ArrayLengthChange(t *testing.T) {
	a := map[string]interface{}{
		"items": []interface{}{"a", "b"},
	}
	b := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "items[2]", diffs[0].Path)
	assert.Equal(t, DiffTypeAdded, diffs[0].Type)
}

func TestEngine_Compare_NullValues(t *testing.T) {
	a := map[string]interface{}{
		"name":  "John",
		"email": nil,
	}
	b := map[string]interface{}{
		"name":  "John",
		"email": "john@test.com",
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "email", diffs[0].Path)
	assert.Equal(t, DiffTypeAdded, diffs[0].Type)
}

func TestEngine_Compare_BothNull(t *testing.T) {
	engine := NewEngine(Config{})
	diffs, err := engine.Compare(nil, nil)
	require.NoError(t, err)
	assert.Empty(t, diffs)
}

func TestEngine_Compare_IgnorePaths(t *testing.T) {
	a := map[string]interface{}{
		"name":      "John",
		"timestamp": "2024-01-01",
		"metadata": map[string]interface{}{
			"internal": "secret",
		},
	}
	b := map[string]interface{}{
		"name":      "Jane",
		"timestamp": "2024-02-01",
		"metadata": map[string]interface{}{
			"internal": "changed",
		},
	}

	engine := NewEngine(Config{
		IgnorePaths: []string{"timestamp", "metadata"},
	})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)
	assert.Equal(t, "name", diffs[0].Path)
}

func TestEngine_Compare_NoDifferences(t *testing.T) {
	a := map[string]interface{}{
		"name": "John",
		"age":  float64(30),
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, a)
	require.NoError(t, err)
	assert.Empty(t, diffs)
}

func TestEngine_Compare_EmptyObjects(t *testing.T) {
	a := map[string]interface{}{}
	b := map[string]interface{}{}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	assert.Empty(t, diffs)
}

func TestEngine_Compare_ComplexNested(t *testing.T) {
	a := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"name": "Alice",
				"roles": []interface{}{"admin", "user"},
			},
		},
	}
	b := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"name": "Alice",
				"roles": []interface{}{"admin", "moderator"},
			},
		},
	}

	engine := NewEngine(Config{})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)
	require.Len(t, diffs, 1)

	assert.Equal(t, "users[0].roles[1]", diffs[0].Path)
	assert.Equal(t, "user", diffs[0].OldValue)
	assert.Equal(t, "moderator", diffs[0].NewValue)
}

func TestEngine_Compare_ArrayAsSet(t *testing.T) {
	a := map[string]interface{}{
		"tags": []interface{}{"go", "rust", "python"},
	}
	b := map[string]interface{}{
		"tags": []interface{}{"python", "go", "java"},
	}

	engine := NewEngine(Config{TreatArraysAsSet: true})
	diffs, err := engine.Compare(a, b)
	require.NoError(t, err)

	hasRemoved := false
	hasAdded := false
	for _, d := range diffs {
		if d.Type == DiffTypeRemoved && d.OldValue == "rust" {
			hasRemoved = true
		}
		if d.Type == DiffTypeAdded && d.NewValue == "java" {
			hasAdded = true
		}
	}
	assert.True(t, hasRemoved, "should detect 'rust' removed")
	assert.True(t, hasAdded, "should detect 'java' added")
}

func TestTextFormatter_Format(t *testing.T) {
	diffs := []Diff{
		{Path: "name", Type: DiffTypeModified, OldValue: "old", NewValue: "new", Severity: SeverityInfo},
		{Path: "email", Type: DiffTypeRemoved, OldValue: "test@test.com", Severity: SeverityWarning},
	}

	formatter := NewTextFormatter(false)
	output, err := formatter.Format(diffs)
	require.NoError(t, err)
	assert.Contains(t, output, "2 difference(s)")
	assert.Contains(t, output, "name")
	assert.Contains(t, output, "email")
}

func TestTextFormatter_NoDiffs(t *testing.T) {
	formatter := NewTextFormatter(false)
	output, err := formatter.Format([]Diff{})
	require.NoError(t, err)
	assert.Contains(t, output, "No differences found")
}

func TestJSONFormatter_Format(t *testing.T) {
	diffs := []Diff{
		{Path: "name", Type: DiffTypeModified, OldValue: "old", NewValue: "new", Severity: SeverityInfo},
	}

	formatter := NewJSONFormatter(true)
	output, err := formatter.Format(diffs)
	require.NoError(t, err)
	assert.Contains(t, output, `"total_changes": 1`)
	assert.Contains(t, output, `"has_breaking": false`)
}

func BenchmarkEngine_Compare_Small(b *testing.B) {
	obj1 := generateTestObject(10)
	obj2 := generateTestObject(10)
	obj2.(map[string]interface{})["field_0"] = "changed"

	engine := NewEngine(Config{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = engine.Compare(obj1, obj2)
	}
}

func BenchmarkEngine_Compare_Medium(b *testing.B) {
	obj1 := generateTestObject(100)
	obj2 := generateTestObject(100)
	obj2.(map[string]interface{})["field_0"] = "changed"

	engine := NewEngine(Config{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = engine.Compare(obj1, obj2)
	}
}

func BenchmarkEngine_Compare_Large(b *testing.B) {
	obj1 := generateTestObject(1000)
	obj2 := generateTestObject(1000)
	obj2.(map[string]interface{})["field_0"] = "changed"

	engine := NewEngine(Config{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = engine.Compare(obj1, obj2)
	}
}

func generateTestObject(fields int) interface{} {
	obj := make(map[string]interface{})
	for i := 0; i < fields; i++ {
		key := "field_" + string(rune('0'+i%10)) + string(rune('0'+i/10%10)) + string(rune('0'+i/100%10))
		if i < fields/3 {
			obj[key] = "value_" + key
		} else if i < 2*fields/3 {
			obj[key] = float64(i)
		} else {
			obj[key] = map[string]interface{}{
				"nested": "value_" + key,
			}
		}
	}
	return obj
}
