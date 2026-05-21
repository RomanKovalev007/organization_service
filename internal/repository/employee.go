package repository

import (
	"context"
	"fmt"

	"github.com/RomanKovalev007/organization_service/internal/domain"
	"gorm.io/gorm"
)

type EmployeeRepo struct {
	db *gorm.DB
}

func NewEmployeeRepo(db *gorm.DB) *EmployeeRepo {
	return &EmployeeRepo{db: db}
}

func (r *EmployeeRepo) WithTx(tx *gorm.DB) *EmployeeRepo {
	return &EmployeeRepo{db: tx}
}

func (r *EmployeeRepo) Create(ctx context.Context, emp *domain.Employee) (*domain.Employee, error) {
	res := r.db.WithContext(ctx).Create(emp)
	if res.Error != nil {
		return nil, fmt.Errorf("create employee: %w", wrapDBError(res.Error))
	}
	return emp, nil
}

func (r *EmployeeRepo) GetByDepartment(ctx context.Context, departmentID int64) ([]domain.Employee, error) {
	var emps []domain.Employee
	res := r.db.WithContext(ctx).
		Where("department_id = ?", departmentID).
		Order("created_at ASC").
		Find(&emps)
	if res.Error != nil {
		return nil, fmt.Errorf("get employees: %w", wrapDBError(res.Error))
	}
	return emps, nil
}

func (r *EmployeeRepo) ReassignAll(ctx context.Context, fromDeptID, toDeptID int64) error {
	res := r.db.WithContext(ctx).
		Model(&domain.Employee{}).
		Where("department_id = ?", fromDeptID).
		Update("department_id", toDeptID)
	if res.Error != nil {
		return fmt.Errorf("reassign employees: %w", wrapDBError(res.Error))
	}
	return nil
}
