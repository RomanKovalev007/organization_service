package handler

import (
	"context"
	"net/http"

	"github.com/RomanKovalev007/organization_service/internal/domain"
	"github.com/RomanKovalev007/organization_service/internal/service"
)

type mockSvc struct {
	CreateDepartmentFn func(ctx context.Context, dept *domain.Department) (*domain.Department, error)
	CreateEmployeeFn   func(ctx context.Context, emp *domain.Employee) (*domain.Employee, error)
	GetDepartmentFn    func(ctx context.Context, id int64, depth int, includeEmployees bool) (*domain.DepartmentTree, error)
	UpdateDepartmentFn func(ctx context.Context, id int64, name *string, parentID *int64) (*domain.Department, error)
	DeleteDepartmentFn func(ctx context.Context, id int64, mode service.DeleteMode, reassignTo *int64) error
}

func (m *mockSvc) CreateDepartment(ctx context.Context, dept *domain.Department) (*domain.Department, error) {
	return m.CreateDepartmentFn(ctx, dept)
}
func (m *mockSvc) CreateEmployee(ctx context.Context, emp *domain.Employee) (*domain.Employee, error) {
	return m.CreateEmployeeFn(ctx, emp)
}
func (m *mockSvc) GetDepartment(ctx context.Context, id int64, depth int, includeEmployees bool) (*domain.DepartmentTree, error) {
	return m.GetDepartmentFn(ctx, id, depth, includeEmployees)
}
func (m *mockSvc) UpdateDepartment(ctx context.Context, id int64, name *string, parentID *int64) (*domain.Department, error) {
	return m.UpdateDepartmentFn(ctx, id, name, parentID)
}
func (m *mockSvc) DeleteDepartment(ctx context.Context, id int64, mode service.DeleteMode, reassignTo *int64) error {
	return m.DeleteDepartmentFn(ctx, id, mode, reassignTo)
}

func newTestMux(svc departmentService) http.Handler {
	h := New(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /departments/", h.CreateDepartment)
	mux.HandleFunc("POST /departments/{id}/employees/", h.CreateEmployee)
	mux.HandleFunc("GET /departments/{id}", h.GetDepartment)
	mux.HandleFunc("PATCH /departments/{id}", h.UpdateDepartment)
	mux.HandleFunc("DELETE /departments/{id}", h.DeleteDepartment)
	return mux
}
