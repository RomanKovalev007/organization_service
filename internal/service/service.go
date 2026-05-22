package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/RomanKovalev007/organization_service/internal/apperr"
	"github.com/RomanKovalev007/organization_service/internal/domain"
)

type DeleteMode string

const (
	DeleteModeCascade  DeleteMode = "cascade"
	DeleteModeReassign DeleteMode = "reassign"
)

type Service struct {
	deptRepo DepartmentRepo
	empRepo  EmployeeRepo
	txRunner TxRunner
}

func NewService(dr DepartmentRepo, er EmployeeRepo, tx TxRunner) *Service {
	return &Service{
		deptRepo: dr,
		empRepo:  er,
		txRunner: tx,
	}
}

func (s *Service) CreateDepartment(ctx context.Context, dept *domain.Department) (*domain.Department, error) {
    resDept, err := s.deptRepo.Create(ctx, dept)
    if err != nil {
        if errors.Is(err, apperr.ErrAlreadyExists) {
            return nil, apperr.New(apperr.CodeAlreadyExists, "department with the same name already exists in this parent")
        }
        if errors.Is(err, apperr.ErrNotFound) {
            return nil, apperr.New(apperr.CodeNotFound, "parent department not found")
        }
        return nil, apperr.New(apperr.CodeInternalError, err.Error())
    }
    slog.Info("department created", "id", resDept.ID, "name", resDept.Name)
    return resDept, nil
}

func (s *Service) CreateEmployee(ctx context.Context, emp *domain.Employee) (*domain.Employee, error) {
	res, err := s.empRepo.Create(ctx, emp)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, apperr.New(apperr.CodeNotFound, "department not found")
		}
		return nil, apperr.New(apperr.CodeInternalError, err.Error())
	}
	slog.Info("employee created", "id", res.ID, "department_id", res.DepartmentID)
	return res, nil
}

func (s *Service) GetDepartment(ctx context.Context, id int64, depth int, includeEmployees bool) (*domain.DepartmentTree, error) {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, apperr.New(apperr.CodeNotFound, "department not found")
		}
		return nil, apperr.New(apperr.CodeInternalError, err.Error())
	}

	return s.recursiveGetDept(ctx, dept, depth, 0, includeEmployees)
}

func (s *Service) UpdateDepartment(ctx context.Context, id int64, name *string, parentID *int64) (*domain.Department, error) {
	if parentID != nil {
		if *parentID == id {
			return nil, apperr.New(apperr.CodeInvalidInput, "department cannot be its own parent")
		}

		_, err := s.deptRepo.GetByID(ctx, *parentID)
		if err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				return nil, apperr.New(apperr.CodeNotFound, "new parent department not found")
			}
			return nil, apperr.New(apperr.CodeInternalError, err.Error())
		}

		isDesc, err := s.deptRepo.IsDescendant(ctx, id, *parentID)
		if err != nil {
			return nil, apperr.New(apperr.CodeInternalError, err.Error())
		}
		if isDesc {
			return nil, apperr.New(apperr.CodeInvalidInput, "cannot move department into its own subtree")
		}
	}

	updates := map[string]any{}
	if name != nil {
		updates["name"] = *name
	}
	if parentID != nil {
		updates["parent_id"] = *parentID
	}

	if len(updates) == 0 {
		return nil, apperr.New(apperr.CodeInvalidInput, "nothing to update")
	}

	res, err := s.deptRepo.Update(ctx, id, updates)
	if err != nil {
		if errors.Is(err, apperr.ErrAlreadyExists) {
			return nil, apperr.New(apperr.CodeAlreadyExists, "department with the same name already exists in this parent")
		} else if errors.Is(err, apperr.ErrNotFound) {
			return nil, apperr.New(apperr.CodeNotFound, "department not found")
		}
		return nil, apperr.New(apperr.CodeInternalError, err.Error())
	}
	slog.Info("department updated", "id", res.ID)
	return res, nil
}

func (s *Service) DeleteDepartment(ctx context.Context, id int64, mode DeleteMode, reassignTo *int64) error {
	_, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return apperr.New(apperr.CodeNotFound, "department not found")
		}
		return apperr.New(apperr.CodeInternalError, err.Error())
	}

	switch mode {
	case DeleteModeCascade:
		if err := s.deptRepo.Delete(ctx, id); err != nil {
			return apperr.New(apperr.CodeInternalError, err.Error())
		}
		slog.Info("department deleted", "id", id, "mode", "cascade")

	case DeleteModeReassign:
		if reassignTo == nil {
			return apperr.New(apperr.CodeInvalidInput, "reassign_to_department_id is required for reassign mode")
		}
		if *reassignTo == id {
			return apperr.New(apperr.CodeInvalidInput, "cannot reassign to the department being deleted")
		}

		_, err := s.deptRepo.GetByID(ctx, *reassignTo)
		if err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				return apperr.New(apperr.CodeNotFound, "reassign target department not found")
			}
			return apperr.New(apperr.CodeInternalError, err.Error())
		}

		if err := s.txRunner.RunInTx(ctx,
			func(deptRepo DepartmentRepo, empRepo EmployeeRepo) error {
				isDesc, err := deptRepo.IsDescendant(ctx, id, *reassignTo)
				if err != nil {
					return fmt.Errorf("is descendant: %w", err)
				}
				if isDesc {
					return apperr.New(apperr.CodeInvalidInput, "reassign target is inside the subtree of the department being deleted")
				}
				if err := empRepo.ReassignAll(ctx, id, *reassignTo); err != nil {
					return fmt.Errorf("reassign employees: %w", err)
				}
				if err := deptRepo.ReparentChildren(ctx, id, *reassignTo); err != nil {
					return fmt.Errorf("reparent children: %w", err)
				}
				if err := deptRepo.Delete(ctx, id); err != nil {
					return fmt.Errorf("delete department: %w", err)
				}
				return nil
			}); err != nil {
			if errors.As(err, new(*apperr.Error)) {
				return err
			}
			return apperr.New(apperr.CodeInternalError, err.Error())
		}
		slog.Info("department deleted", "id", id, "mode", "reassign", "reassign_to", *reassignTo)

	default:
		return apperr.New(apperr.CodeInvalidInput, "mode must be 'cascade' or 'reassign'")
	}

	return nil
}

func (s *Service) recursiveGetDept(ctx context.Context, dept *domain.Department, depth int, curr int, includeEmployees bool) (*domain.DepartmentTree, error) {
	var emps []domain.Employee
	if includeEmployees {
		var err error
		emps, err = s.empRepo.GetByDepartment(ctx, dept.ID)
		if err != nil {
			return nil, apperr.New(apperr.CodeInternalError, fmt.Sprintf("failed to get employees for department %d: %s", dept.ID, err.Error()))
		}
	}

	if curr == depth {
		return &domain.DepartmentTree{
			Department: *dept,
			Employees:  emps,
		}, nil
	}

	children, err := s.deptRepo.GetChildren(ctx, dept.ID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternalError, fmt.Sprintf("failed to get children for department %d: %s", dept.ID, err.Error()))
	}

	childrenTree := make([]*domain.DepartmentTree, 0, len(children))
	for _, ch := range children {
		cht, err := s.recursiveGetDept(ctx, ch, depth, curr+1, includeEmployees)
		if err != nil {
			return nil, err
		}
		childrenTree = append(childrenTree, cht)
	}

	return &domain.DepartmentTree{
		Department: *dept,
		Employees:  emps,
		Children:   childrenTree,
	}, nil
}
