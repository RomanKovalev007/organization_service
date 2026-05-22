package domain

import "time"

type Department struct {
	ID        int64     `gorm:"primaryKey;column:id"         json:"id"`
	Name      string    `gorm:"column:name;not null"         json:"name"`
	ParentID  *int64    `gorm:"column:parent_id"             json:"parent_id,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

type Employee struct {
	ID           int64      `gorm:"primaryKey;column:id"              json:"id"`
	DepartmentID int64      `gorm:"column:department_id;not null"     json:"department_id"`
	FullName     string     `gorm:"column:full_name;not null"         json:"full_name"`
	Position     string     `gorm:"column:position;not null"          json:"position"`
	HiredAt      *time.Time `gorm:"column:hired_at"                   json:"hired_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"  json:"created_at"`
}

type DepartmentTree struct {
	Department
	Employees []Employee
	Children  []*DepartmentTree
}
