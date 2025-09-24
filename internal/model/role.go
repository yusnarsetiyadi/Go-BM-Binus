package model

import "bm_binus/internal/abstraction"

type RoleEntity struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDelete    bool   `json:"is_delete"`
}

// RoleEntityModel ...
type RoleEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	RoleEntity

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (RoleEntityModel) TableName() string {
	return "role"
}

type RoleCountDataModel struct {
	Count int `json:"count"`
}
