package service

import (
	"context"

	"github.com/RomanKovalev007/organization_service/internal/domain"
)

type mockDeptRepo struct {
	CreateFn          func(ctx context.Context, dept *domain.Department) (*domain.Department, error)
	GetByIDFn         func(ctx context.Context, id int64) (*domain.Department, error)
	GetChildrenFn     func(ctx context.Context, parentID int64) ([]*domain.Department, error)
	UpdateFn          func(ctx context.Context, id int64, updates map[string]any) (*domain.Department, error)
	DeleteFn          func(ctx context.Context, id int64) error
	IsDescendantFn    func(ctx context.Context, deptID, targetID int64) (bool, error)
	ReparentChildrenFn func(ctx context.Context, fromParentID, toParentID int64) error
}

func (m *mockDeptRepo) Create(ctx context.Context, dept *domain.Department) (*domain.Department, error) {
	return m.CreateFn(ctx, dept)
}
func (m *mockDeptRepo) GetByID(ctx context.Context, id int64) (*domain.Department, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *mockDeptRepo) GetChildren(ctx context.Context, parentID int64) ([]*domain.Department, error) {
	return m.GetChildrenFn(ctx, parentID)
}
func (m *mockDeptRepo) Update(ctx context.Context, id int64, updates map[string]any) (*domain.Department, error) {
	return m.UpdateFn(ctx, id, updates)
}
func (m *mockDeptRepo) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}
func (m *mockDeptRepo) IsDescendant(ctx context.Context, deptID, targetID int64) (bool, error) {
	return m.IsDescendantFn(ctx, deptID, targetID)
}
func (m *mockDeptRepo) ReparentChildren(ctx context.Context, fromParentID, toParentID int64) error {
	return m.ReparentChildrenFn(ctx, fromParentID, toParentID)
}

type mockEmpRepo struct {
	CreateFn          func(ctx context.Context, emp *domain.Employee) (*domain.Employee, error)
	GetByDepartmentFn func(ctx context.Context, departmentID int64) ([]domain.Employee, error)
	ReassignAllFn     func(ctx context.Context, fromDeptID, toDeptID int64) error
}

func (m *mockEmpRepo) Create(ctx context.Context, emp *domain.Employee) (*domain.Employee, error) {
	return m.CreateFn(ctx, emp)
}
func (m *mockEmpRepo) GetByDepartment(ctx context.Context, departmentID int64) ([]domain.Employee, error) {
	return m.GetByDepartmentFn(ctx, departmentID)
}
func (m *mockEmpRepo) ReassignAll(ctx context.Context, fromDeptID, toDeptID int64) error {
	return m.ReassignAllFn(ctx, fromDeptID, toDeptID)
}

type mockTxRunner struct {
	deptRepo DepartmentRepo
	empRepo  EmployeeRepo
}

func (m *mockTxRunner) RunInTx(ctx context.Context, fn func(DepartmentRepo, EmployeeRepo) error) error {
	return fn(m.deptRepo, m.empRepo)
}
