package model

import "bm_binus/internal/abstraction"

type StatusEntity struct {
	Name string `json:"name"`
}

// StatusEntityModel ...
type StatusEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	StatusEntity

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (StatusEntityModel) TableName() string {
	return "status"
}

type StatusCountDataModel struct {
	Count int `json:"count"`
}
