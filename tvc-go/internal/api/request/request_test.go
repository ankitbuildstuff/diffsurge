package request

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSON(t *testing.T) {
	body := `{"name":"test"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	var dst struct{ Name string }
	err := ParseJSON(r, 0, &dst)
	require.NoError(t, err)
	assert.Equal(t, "test", dst.Name)
}

func TestParseJSON_Invalid(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
	var dst struct{}
	err := ParseJSON(r, 0, &dst)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestPathUUID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.SetPathValue("id", "550e8400-e29b-41d4-a716-446655440000")
	id, err := PathUUID(r, "id")
	require.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", id.String())
}

func TestPathUUID_Missing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	_, err := PathUUID(r, "id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestPathUUID_Invalid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.SetPathValue("id", "not-a-uuid")
	_, err := PathUUID(r, "id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID")
}

func TestQueryString(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test?key=val", nil)
	assert.Equal(t, "val", QueryString(r, "key", "default"))
	assert.Equal(t, "default", QueryString(r, "missing", "default"))
}

func TestQueryInt(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test?n=42&bad=abc", nil)
	assert.Equal(t, 42, QueryInt(r, "n", 0))
	assert.Equal(t, 0, QueryInt(r, "bad", 0))
	assert.Equal(t, 10, QueryInt(r, "missing", 10))
}

func TestQueryTime(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test?t=2026-01-01T00:00:00Z&bad=nope", nil)
	result := QueryTime(r, "t")
	require.NotNil(t, result)
	assert.Equal(t, 2026, result.Year())

	assert.Nil(t, QueryTime(r, "bad"))
	assert.Nil(t, QueryTime(r, "missing"))
}

func TestQueryStringSlice(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test?m=GET,POST,PUT", nil)
	result := QueryStringSlice(r, "m")
	assert.Equal(t, []string{"GET", "POST", "PUT"}, result)
	assert.Nil(t, QueryStringSlice(r, "missing"))
}

func TestQueryIntSlice(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test?s=200,404,500", nil)
	result := QueryIntSlice(r, "s")
	assert.Equal(t, []int{200, 404, 500}, result)
	assert.Nil(t, QueryIntSlice(r, "missing"))
}

func TestDrainBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("body content"))
	DrainBody(r)

	r = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Body = nil
	DrainBody(r)
}
