package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RomanKovalev007/organization_service/internal/apperr"
	"github.com/RomanKovalev007/organization_service/internal/domain"
	"github.com/RomanKovalev007/organization_service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func do(t *testing.T, handler http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, target, &buf)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) apperr.Error {
	t.Helper()
	var e apperr.Error
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&e))
	return e
}

// --- CreateDepartment ---

func TestCreateDepartment_Success(t *testing.T) {
	svc := &mockSvc{
		CreateDepartmentFn: func(_ context.Context, d *domain.Department) (*domain.Department, error) {
			d.ID = 1
			return d, nil
		},
	}
	rec := do(t, newTestMux(svc), http.MethodPost, "/departments/", map[string]any{"name": "Engineering"})
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp domain.Department
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, "Engineering", resp.Name)
}

func TestCreateDepartment_EmptyName(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodPost, "/departments/", map[string]any{"name": "  "})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, apperr.CodeInvalidInput, decodeError(t, rec).Code)
}

func TestCreateDepartment_NameTooLong(t *testing.T) {
	name := string(make([]byte, 201))
	rec := do(t, newTestMux(&mockSvc{}), http.MethodPost, "/departments/", map[string]any{"name": name})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, apperr.CodeInvalidInput, decodeError(t, rec).Code)
}

func TestCreateDepartment_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()
	newTestMux(&mockSvc{}).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateDepartment_AlreadyExists(t *testing.T) {
	svc := &mockSvc{
		CreateDepartmentFn: func(_ context.Context, _ *domain.Department) (*domain.Department, error) {
			return nil, apperr.New(apperr.CodeAlreadyExists, "already exists")
		},
	}
	rec := do(t, newTestMux(svc), http.MethodPost, "/departments/", map[string]any{"name": "Dup"})
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, apperr.CodeAlreadyExists, decodeError(t, rec).Code)
}

func TestCreateDepartment_ParentNotFound(t *testing.T) {
	svc := &mockSvc{
		CreateDepartmentFn: func(_ context.Context, _ *domain.Department) (*domain.Department, error) {
			return nil, apperr.New(apperr.CodeNotFound, "parent not found")
		},
	}
	rec := do(t, newTestMux(svc), http.MethodPost, "/departments/", map[string]any{"name": "Sub", "parent_id": 99})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- CreateEmployee ---

func TestCreateEmployee_Success(t *testing.T) {
	svc := &mockSvc{
		CreateEmployeeFn: func(_ context.Context, e *domain.Employee) (*domain.Employee, error) {
			e.ID = 5
			return e, nil
		},
	}
	rec := do(t, newTestMux(svc), http.MethodPost, "/departments/1/employees/",
		map[string]any{"full_name": "Ivan Petrov", "position": "Developer"})
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp domain.Employee
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, int64(5), resp.ID)
}

func TestCreateEmployee_InvalidDeptID(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodPost, "/departments/abc/employees/",
		map[string]any{"full_name": "Ivan", "position": "Dev"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateEmployee_EmptyFullName(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodPost, "/departments/1/employees/",
		map[string]any{"full_name": "", "position": "Dev"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, apperr.CodeInvalidInput, decodeError(t, rec).Code)
}

func TestCreateEmployee_EmptyPosition(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodPost, "/departments/1/employees/",
		map[string]any{"full_name": "Ivan", "position": "  "})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, apperr.CodeInvalidInput, decodeError(t, rec).Code)
}

func TestCreateEmployee_DeptNotFound(t *testing.T) {
	svc := &mockSvc{
		CreateEmployeeFn: func(_ context.Context, _ *domain.Employee) (*domain.Employee, error) {
			return nil, apperr.New(apperr.CodeNotFound, "department not found")
		},
	}
	rec := do(t, newTestMux(svc), http.MethodPost, "/departments/1/employees/",
		map[string]any{"full_name": "Ivan", "position": "Dev"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- GetDepartment ---

func TestGetDepartment_Success(t *testing.T) {
	svc := &mockSvc{
		GetDepartmentFn: func(_ context.Context, id int64, _ int, _ bool) (*domain.DepartmentTree, error) {
			return &domain.DepartmentTree{Department: domain.Department{ID: id, Name: "Root"}}, nil
		},
	}
	rec := do(t, newTestMux(svc), http.MethodGet, "/departments/1", nil)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.DepartmentTree
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, int64(1), resp.ID)
}

func TestGetDepartment_InvalidID(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodGet, "/departments/abc", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetDepartment_DepthOutOfRange(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodGet, "/departments/1?depth=6", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, apperr.CodeInvalidInput, decodeError(t, rec).Code)
}

func TestGetDepartment_DepthZero(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodGet, "/departments/1?depth=0", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetDepartment_NotFound(t *testing.T) {
	svc := &mockSvc{
		GetDepartmentFn: func(_ context.Context, _ int64, _ int, _ bool) (*domain.DepartmentTree, error) {
			return nil, apperr.New(apperr.CodeNotFound, "not found")
		},
	}
	rec := do(t, newTestMux(svc), http.MethodGet, "/departments/99", nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- UpdateDepartment ---

func TestUpdateDepartment_Success(t *testing.T) {
	svc := &mockSvc{
		UpdateDepartmentFn: func(_ context.Context, id int64, name *string, _ *int64) (*domain.Department, error) {
			return &domain.Department{ID: id, Name: *name}, nil
		},
	}
	rec := do(t, newTestMux(svc), http.MethodPatch, "/departments/1", map[string]any{"name": "New Name"})
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.Department
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "New Name", resp.Name)
}

func TestUpdateDepartment_InvalidID(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodPatch, "/departments/abc", map[string]any{"name": "X"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateDepartment_EmptyName(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodPatch, "/departments/1", map[string]any{"name": "  "})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, apperr.CodeInvalidInput, decodeError(t, rec).Code)
}

func TestUpdateDepartment_NotFound(t *testing.T) {
	svc := &mockSvc{
		UpdateDepartmentFn: func(_ context.Context, _ int64, _ *string, _ *int64) (*domain.Department, error) {
			return nil, apperr.New(apperr.CodeNotFound, "not found")
		},
	}
	rec := do(t, newTestMux(svc), http.MethodPatch, "/departments/99", map[string]any{"name": "X"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- DeleteDepartment ---

func TestDeleteDepartment_Cascade_Success(t *testing.T) {
	svc := &mockSvc{
		DeleteDepartmentFn: func(_ context.Context, _ int64, _ service.DeleteMode, _ *int64) error {
			return nil
		},
	}
	rec := do(t, newTestMux(svc), http.MethodDelete, "/departments/1?mode=cascade", nil)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteDepartment_Reassign_Success(t *testing.T) {
	svc := &mockSvc{
		DeleteDepartmentFn: func(_ context.Context, _ int64, _ service.DeleteMode, _ *int64) error {
			return nil
		},
	}
	rec := do(t, newTestMux(svc), http.MethodDelete, "/departments/1?mode=reassign&reassign_to_department_id=2", nil)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteDepartment_InvalidMode(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodDelete, "/departments/1?mode=wrong", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, apperr.CodeInvalidInput, decodeError(t, rec).Code)
}

func TestDeleteDepartment_InvalidID(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodDelete, "/departments/abc?mode=cascade", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteDepartment_InvalidReassignID(t *testing.T) {
	rec := do(t, newTestMux(&mockSvc{}), http.MethodDelete, "/departments/1?mode=reassign&reassign_to_department_id=abc", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
