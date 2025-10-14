package model

import (
	"bm_binus/internal/abstraction"
	"time"

	"gorm.io/gorm"
)

type RequestEntity struct {
	UserId           int       `json:"user_id"`
	EventName        string    `json:"event_name"`
	EventLocation    string    `json:"event_location"`
	EventDateStart   time.Time `json:"event_date_start"`
	EventDateEnd     time.Time `json:"event_date_end"`
	Description      string    `json:"description"`
	EventTypeId      int       `json:"event_type_id"`
	CountParticipant int       `json:"count_participant"`
	StatusId         int       `json:"status_id"`
	IsDelete         bool      `json:"is_delete"`
}

// RequestEntityModel ...
type RequestEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	RequestEntity

	abstraction.Entity

	EventType EventTypeEntityModel `json:"event_type" gorm:"foreignKey:EventTypeId"`
	Status    StatusEntityModel    `json:"status" gorm:"foreignKey:StatusId"`

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (RequestEntityModel) TableName() string {
	return "request"
}

type RequestCountDataModel struct {
	Count int `json:"count"`
}

func (m *RequestEntityModel) BeforeUpdate(tx *gorm.DB) (err error) {
	// m.UpdatedAt = general.NowLocal()
	return
}

func (m *RequestEntityModel) BeforeCreate(tx *gorm.DB) (err error) {
	// m.CreatedAt = *general.Now()
	return
}
