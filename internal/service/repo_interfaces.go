package service

import (
	"context"

	"github.com/RomanKovalev007/organization_service/internal/domain"
)

type DepartmentRepo interface {
	Create(ctx context.Context, dept *domain.Department) (*domain.Department, error)
	GetByID(ctx context.Context, id int64) (*domain.Department, error)
	GetChildren(ctx context.Context, parentID int64) ([]*domain.Department, error)
	Update(ctx context.Context, id int64, updates map[string]any) (*domain.Department, error)
	Delete(ctx context.Context, id int64) error
	IsDescendant(ctx context.Context, deptID, targetID int64) (bool, error)
	ReparentChildren(ctx context.Context, fromParentID, toParentID int64) error
}

type EmployeeRepo interface {
	Create(ctx context.Context, emp *domain.Employee) (*domain.Employee, error)
	GetByDepartment(ctx context.Context, departmentID int64) ([]domain.Employee, error)
	ReassignAll(ctx context.Context, fromDeptID, toDeptID int64) error
}

type TxRunner interface {
	RunInTx(ctx context.Context, fn func(deptRepo DepartmentRepo, empRepo EmployeeRepo) error) error
}
