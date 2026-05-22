package repository

import (
	"context"

	"github.com/RomanKovalev007/organization_service/internal/service"
	"gorm.io/gorm"
)

type TxManager struct {
	db       *gorm.DB
	deptRepo *DepartmentRepo
	empRepo  *EmployeeRepo
}

func NewTxManager(db *gorm.DB, dr *DepartmentRepo, er *EmployeeRepo) *TxManager {
	return &TxManager{db: db, deptRepo: dr, empRepo: er}
}

func (m *TxManager) RunInTx(ctx context.Context, fn func(service.DepartmentRepo, service.EmployeeRepo) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(m.deptRepo.WithTx(tx), m.empRepo.WithTx(tx))
	})
}
