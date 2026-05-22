package handler

type CreateDepartmentReq struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id,omitempty"`
}

type CreateEmployeeReq struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at,omitempty"`
}

type UpdateDepartmentReq struct {
	Name     *string `json:"name,omitempty"`
	ParentID *int64  `json:"parent_id,omitempty"`
}

