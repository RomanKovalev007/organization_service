package service

import (
	"context"
	"testing"

	"github.com/RomanKovalev007/organization_service/internal/apperr"
	"github.com/RomanKovalev007/organization_service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

// CreateDepartment

func TestCreateDepartment_Success(t *testing.T) {
	dept := &domain.Department{ID: 1, Name: "Engineering"}
	svc := NewService(
		&mockDeptRepo{CreateFn: func(_ context.Context, d *domain.Department) (*domain.Department, error) {
			return dept, nil
		}},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	res, err := svc.CreateDepartment(ctx, &domain.Department{Name: "Engineering"})
	require.NoError(t, err)
	assert.Equal(t, dept, res)
}

func TestCreateDepartment_ParentNotFound(t *testing.T) {
	svc := NewService(
		&mockDeptRepo{CreateFn: func(_ context.Context, _ *domain.Department) (*domain.Department, error) {
			return nil, apperr.ErrNotFound
		}},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	_, err := svc.CreateDepartment(ctx, &domain.Department{Name: "Sub"})
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeNotFound, appErr.Code)
}

func TestCreateDepartment_AlreadyExists(t *testing.T) {
	svc := NewService(
		&mockDeptRepo{CreateFn: func(_ context.Context, _ *domain.Department) (*domain.Department, error) {
			return nil, apperr.ErrAlreadyExists
		}},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	_, err := svc.CreateDepartment(ctx, &domain.Department{Name: "Dup"})
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeAlreadyExists, appErr.Code)
}

//CreateEmployee

func TestCreateEmployee_Success(t *testing.T) {
	emp := &domain.Employee{ID: 1, DepartmentID: 10, FullName: "Ivan", Position: "Dev"}
	svc := NewService(
		&mockDeptRepo{},
		&mockEmpRepo{CreateFn: func(_ context.Context, e *domain.Employee) (*domain.Employee, error) {
			return emp, nil
		}},
		&mockTxRunner{},
	)

	res, err := svc.CreateEmployee(ctx, &domain.Employee{DepartmentID: 10, FullName: "Ivan", Position: "Dev"})
	require.NoError(t, err)
	assert.Equal(t, emp, res)
}

func TestCreateEmployee_DepartmentNotFound(t *testing.T) {
	svc := NewService(
		&mockDeptRepo{},
		&mockEmpRepo{CreateFn: func(_ context.Context, _ *domain.Employee) (*domain.Employee, error) {
			return nil, apperr.ErrNotFound
		}},
		&mockTxRunner{},
	)

	_, err := svc.CreateEmployee(ctx, &domain.Employee{DepartmentID: 99})
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeNotFound, appErr.Code)
}

// GetDepartment

func TestGetDepartment_NotFound(t *testing.T) {
	svc := NewService(
		&mockDeptRepo{GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
			return nil, apperr.ErrNotFound
		}},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	_, err := svc.GetDepartment(ctx, 1, 1, false)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeNotFound, appErr.Code)
}

func TestGetDepartment_Depth1_NoChildren(t *testing.T) {
	dept := &domain.Department{ID: 1, Name: "Root"}
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
				return dept, nil
			},
			GetChildrenFn: func(_ context.Context, _ int64) ([]*domain.Department, error) {
				return nil, nil
			},
		},
		&mockEmpRepo{
			GetByDepartmentFn: func(_ context.Context, _ int64) ([]domain.Employee, error) {
				return nil, nil
			},
		},
		&mockTxRunner{},
	)

	tree, err := svc.GetDepartment(ctx, 1, 1, true)
	require.NoError(t, err)
	assert.Equal(t, dept.ID, tree.ID)
	assert.Empty(t, tree.Children)
	assert.Empty(t, tree.Employees)
}

func TestGetDepartment_Depth1_WithChildren(t *testing.T) {
	root := &domain.Department{ID: 1, Name: "Root"}
	child := &domain.Department{ID: 2, Name: "Child", ParentID: &root.ID}
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, id int64) (*domain.Department, error) {
				if id == 1 {
					return root, nil
				}
				return child, nil
			},
			GetChildrenFn: func(_ context.Context, parentID int64) ([]*domain.Department, error) {
				if parentID == 1 {
					return []*domain.Department{child}, nil
				}
				return nil, nil
			},
		},
		&mockEmpRepo{
			GetByDepartmentFn: func(_ context.Context, _ int64) ([]domain.Employee, error) {
				return nil, nil
			},
		},
		&mockTxRunner{},
	)

	tree, err := svc.GetDepartment(ctx, 1, 1, false)
	require.NoError(t, err)
	assert.Equal(t, root.ID, tree.ID)
	require.Len(t, tree.Children, 1)
	assert.Equal(t, child.ID, tree.Children[0].ID)
}

// UpdateDepartment

func TestUpdateDepartment_NothingToUpdate(t *testing.T) {
	svc := NewService(&mockDeptRepo{}, &mockEmpRepo{}, &mockTxRunner{})

	_, err := svc.UpdateDepartment(ctx, 1, nil, nil)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidInput, appErr.Code)
}

func TestUpdateDepartment_SelfParent(t *testing.T) {
	svc := NewService(&mockDeptRepo{}, &mockEmpRepo{}, &mockTxRunner{})

	id := int64(1)
	_, err := svc.UpdateDepartment(ctx, 1, nil, &id)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidInput, appErr.Code)
}

func TestUpdateDepartment_ParentNotFound(t *testing.T) {
	parentID := int64(99)
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
				return nil, apperr.ErrNotFound
			},
		},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	_, err := svc.UpdateDepartment(ctx, 1, nil, &parentID)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeNotFound, appErr.Code)
}

func TestUpdateDepartment_CycleDetected(t *testing.T) {
	parentID := int64(2)
	dept := &domain.Department{ID: 1, Name: "Root"}
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
				return dept, nil
			},
			IsDescendantFn: func(_ context.Context, _, _ int64) (bool, error) {
				return true, nil
			},
		},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	_, err := svc.UpdateDepartment(ctx, 1, nil, &parentID)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidInput, appErr.Code)
}

func TestUpdateDepartment_Success(t *testing.T) {
	name := "New Name"
	updated := &domain.Department{ID: 1, Name: name}
	svc := NewService(
		&mockDeptRepo{
			UpdateFn: func(_ context.Context, _ int64, _ map[string]any) (*domain.Department, error) {
				return updated, nil
			},
		},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	res, err := svc.UpdateDepartment(ctx, 1, &name, nil)
	require.NoError(t, err)
	assert.Equal(t, name, res.Name)
}

// DeleteDepartment

func TestDeleteDepartment_NotFound(t *testing.T) {
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
				return nil, apperr.ErrNotFound
			},
		},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	err := svc.DeleteDepartment(ctx, 1, DeleteModeCascade, nil)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeNotFound, appErr.Code)
}

func TestDeleteDepartment_Cascade_Success(t *testing.T) {
	dept := &domain.Department{ID: 1, Name: "Root"}
	deleted := false
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
				return dept, nil
			},
			DeleteFn: func(_ context.Context, _ int64) error {
				deleted = true
				return nil
			},
		},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	err := svc.DeleteDepartment(ctx, 1, DeleteModeCascade, nil)
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestDeleteDepartment_Reassign_NoTarget(t *testing.T) {
	dept := &domain.Department{ID: 1}
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
				return dept, nil
			},
		},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	err := svc.DeleteDepartment(ctx, 1, DeleteModeReassign, nil)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidInput, appErr.Code)
}

func TestDeleteDepartment_Reassign_SelfTarget(t *testing.T) {
	dept := &domain.Department{ID: 1}
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
				return dept, nil
			},
		},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	id := int64(1)
	err := svc.DeleteDepartment(ctx, 1, DeleteModeReassign, &id)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidInput, appErr.Code)
}

func TestDeleteDepartment_Reassign_TargetNotFound(t *testing.T) {
	reassignTo := int64(99)
	svc := NewService(
		&mockDeptRepo{
			GetByIDFn: func(_ context.Context, id int64) (*domain.Department, error) {
				if id == 1 {
					return &domain.Department{ID: 1}, nil
				}
				return nil, apperr.ErrNotFound
			},
		},
		&mockEmpRepo{},
		&mockTxRunner{},
	)

	err := svc.DeleteDepartment(ctx, 1, DeleteModeReassign, &reassignTo)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeNotFound, appErr.Code)
}

func TestDeleteDepartment_Reassign_TargetInSubtree(t *testing.T) {
	reassignTo := int64(2)
	deptRepo := &mockDeptRepo{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
			return &domain.Department{ID: 1}, nil
		},
		IsDescendantFn: func(_ context.Context, _, _ int64) (bool, error) {
			return true, nil
		},
		ReparentChildrenFn: func(_ context.Context, _, _ int64) error { return nil },
		DeleteFn: func(_ context.Context, _ int64) error { return nil },
	}
	empRepo := &mockEmpRepo{
		ReassignAllFn: func(_ context.Context, _, _ int64) error { return nil },
	}
	svc := NewService(deptRepo, empRepo, &mockTxRunner{deptRepo: deptRepo, empRepo: empRepo})

	err := svc.DeleteDepartment(ctx, 1, DeleteModeReassign, &reassignTo)
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidInput, appErr.Code)
}

func TestDeleteDepartment_Reassign_Success(t *testing.T) {
	reassignTo := int64(2)
	reassigned := false
	reparented := false
	deleted := false

	deptRepo := &mockDeptRepo{
		GetByIDFn: func(_ context.Context, _ int64) (*domain.Department, error) {
			return &domain.Department{ID: 1}, nil
		},
		IsDescendantFn: func(_ context.Context, _, _ int64) (bool, error) {
			return false, nil
		},
		ReparentChildrenFn: func(_ context.Context, _, _ int64) error {
			reparented = true
			return nil
		},
		DeleteFn: func(_ context.Context, _ int64) error {
			deleted = true
			return nil
		},
	}
	empRepo := &mockEmpRepo{
		ReassignAllFn: func(_ context.Context, _, _ int64) error {
			reassigned = true
			return nil
		},
	}
	svc := NewService(deptRepo, empRepo, &mockTxRunner{deptRepo: deptRepo, empRepo: empRepo})

	err := svc.DeleteDepartment(ctx, 1, DeleteModeReassign, &reassignTo)
	require.NoError(t, err)
	assert.True(t, reassigned)
	assert.True(t, reparented)
	assert.True(t, deleted)
}
