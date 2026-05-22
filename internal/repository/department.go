package repository

import (
	"context"
	"fmt"

	"github.com/RomanKovalev007/organization_service/internal/apperr"
	"github.com/RomanKovalev007/organization_service/internal/domain"
	"gorm.io/gorm"
)

type DepartmentRepo struct {
	db *gorm.DB
}

func NewDepartmentRepo(db *gorm.DB) *DepartmentRepo {
	return &DepartmentRepo{db: db}
}

func (r *DepartmentRepo) WithTx(tx *gorm.DB) *DepartmentRepo {
	return &DepartmentRepo{db: tx}
}

func (r *DepartmentRepo) Create(ctx context.Context, dept *domain.Department) (*domain.Department, error) {
	res := r.db.WithContext(ctx).Create(dept)
	if res.Error != nil {
		return nil, fmt.Errorf("create department: %w", wrapDBError(res.Error))
	}
	return dept, nil
}

func (r *DepartmentRepo) GetByID(ctx context.Context, id int64) (*domain.Department, error) {
	var dept domain.Department
	res := r.db.WithContext(ctx).First(&dept, id)
	if res.Error != nil {
		return nil, fmt.Errorf("get department: %w", wrapDBError(res.Error))
	}
	return &dept, nil
}

func (r *DepartmentRepo) GetChildren(ctx context.Context, parentID int64) ([]domain.Department, error) {
	var depts []domain.Department
	res := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&depts)
	if res.Error != nil {
		return nil, fmt.Errorf("get children: %w", wrapDBError(res.Error))
	}
	return depts, nil
}

func (r *DepartmentRepo) Update(ctx context.Context, id int64, updates map[string]any) (*domain.Department, error) {
	res := r.db.WithContext(ctx).
		Model(&domain.Department{}).
		Where("id = ?", id).
		Updates(updates)
	if res.Error != nil {
		return nil, fmt.Errorf("update department: %w", wrapDBError(res.Error))
	}
	if res.RowsAffected == 0 {
		return nil, apperr.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *DepartmentRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&domain.Department{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete department: %w", wrapDBError(res.Error))
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *DepartmentRepo) IsDescendant(ctx context.Context, deptID, targetID int64) (bool, error) {
	var count int64
	res := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE tree AS (
			SELECT id FROM departments WHERE id = ?
			UNION ALL
			SELECT d.id FROM departments d
			JOIN tree t ON d.parent_id = t.id
		)
		SELECT COUNT(*) FROM tree WHERE id = ?
	`, deptID, targetID).Scan(&count)
	if res.Error != nil {
		return false, fmt.Errorf("is descendant: %w", res.Error)
	}
	return count > 0, nil
}
