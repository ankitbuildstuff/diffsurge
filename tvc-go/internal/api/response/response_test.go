package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, map[string]string{"key": "value"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `"key":"value"`)
}

func TestJSON_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	Created(w, map[string]int{"id": 1})
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"id":1`)
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	NoContent(w)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestPaginated(t *testing.T) {
	w := httptest.NewRecorder()
	data := []string{"a", "b"}
	meta := PaginationMeta{NextCursor: "abc", HasMore: true}
	Paginated(w, data, meta)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp paginatedResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Pagination.HasMore)
	assert.Equal(t, "abc", resp.Pagination.NextCursor)
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	BadRequest(w, "bad input")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "bad input")
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	NotFound(w, "Project")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Project not found")
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	Unauthorized(w, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication required")

	w = httptest.NewRecorder()
	Unauthorized(w, "custom msg")
	assert.Contains(t, w.Body.String(), "custom msg")
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	Forbidden(w, "")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")

	w = httptest.NewRecorder()
	Forbidden(w, "no access")
	assert.Contains(t, w.Body.String(), "no access")
}

func TestConflict(t *testing.T) {
	w := httptest.NewRecorder()
	Conflict(w, "already exists")
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "already exists")
}

func TestRateLimited(t *testing.T) {
	w := httptest.NewRecorder()
	RateLimited(w)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	InternalError(w)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPayloadTooLarge(t *testing.T) {
	w := httptest.NewRecorder()
	PayloadTooLarge(w, "1MB")
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Contains(t, w.Body.String(), "1MB")
}

func TestValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	ValidationError(w, []FieldError{{Field: "name", Message: "required"}})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "name")
	assert.Contains(t, w.Body.String(), "required")
}
