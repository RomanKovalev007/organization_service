package handler

import (
	"time"
)


// Request DTOs (JSON body)

type CreateDepartmentReq struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id,omitempty"`
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
	Mode                   string
	ReassignToDepartmentID *int64
}

