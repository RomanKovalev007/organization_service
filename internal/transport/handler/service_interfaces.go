package handler

import (
	"context"

	"github.com/RomanKovalev007/organization_service/internal/domain"
	"github.com/RomanKovalev007/organization_service/internal/service"
)

type departmentService interface {
	CreateDepartment(ctx context.Context, dept *domain.Department) (*domain.Department, error)
	CreateEmployee(ctx context.Context, emp *domain.Employee) (*domain.Employee, error)
	GetDepartment(ctx context.Context, id int64, depth int, includeEmployees bool) (*domain.DepartmentTree, error)
	UpdateDepartment(ctx context.Context, id int64, name *string, parentID *int64) (*domain.Department, error)
	DeleteDepartment(ctx context.Context, id int64, mode service.DeleteMode, reassignTo *int64) error
}
