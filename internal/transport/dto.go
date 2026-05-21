package transport

import (
	"time"

	"github.com/RomanKovalev007/organization_service/internal/domain"
)

type DeleteMode string

const (
	ModeCascade  DeleteMode = "cascade"
	ModeReassign DeleteMode = "reassign"
)

// Request DTOs (JSON body)

type CreateDepartmentReq struct {
	Name     string  `json:"name"`
	ParentID *int64  `json:"parent_id,omitempty"`
}

type CreateEmployeeReq struct {
	FullName string     `json:"full_name"`
	Position string     `json:"position"`
	HiredAt  *time.Time `json:"hired_at,omitempty"`
}

type UpdateDepartmentReq struct {
	Name     *string `json:"name,omitempty"`
	ParentID *int64  `json:"parent_id,omitempty"`
}

// Query-param DTOs

type GetDepartmentReq struct {
	Depth            int
	IncludeEmployees bool
}

type DeleteDepartmentReq struct {
	Mode                   DeleteMode
	ReassignToDepartmentID *int64
}

// Response DTOs

// DepartmentNode is the recursive tree node returned by GET /departments/{id}
type DepartmentNode struct {
	ID        int64              `json:"id"`
	Name      string             `json:"name"`
	ParentID  *int64             `json:"parent_id,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	Employees []domain.Employee  `json:"employees,omitempty"`
	Children  []DepartmentNode   `json:"children,omitempty"`
}
