package transport

import (
	"time"
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

type DepartmentResp struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    ParentID  *int64    `json:"parent_id,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

type EmployeeResp struct {
    ID           int64      `json:"id"`
    DepartmentID int64      `json:"department_id"`
    FullName     string     `json:"full_name"`
    Position     string     `json:"position"`
    HiredAt      *time.Time `json:"hired_at,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
}

type DepartmentNode struct {
    DepartmentResp
    Employees []EmployeeResp  `json:"employees,omitempty"`
    Children  []DepartmentNode `json:"children,omitempty"`
}
