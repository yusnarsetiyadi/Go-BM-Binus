package model

import "bm_binus/internal/abstraction"

type EventTypeEntity struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	IsDelete bool   `json:"is_delete"`
}

// EventTypeEntityModel ...
type EventTypeEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	EventTypeEntity

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (EventTypeEntityModel) TableName() string {
	return "event_type"
}

type EventTypeCountDataModel struct {
	Count int `json:"count"`
}
