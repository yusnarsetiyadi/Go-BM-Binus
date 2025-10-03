package model

import (
	"bm_binus/internal/abstraction"

	"gorm.io/gorm"
)

type NotificationEntity struct {
	Title     string `json:"title"`
	Message   string `json:"message"`
	IsRead    bool   `json:"is_read"`
	UserId    int    `json:"user_id"`
	RequestId int    `json:"request_id"`
}

// NotificationEntityModel ...
type NotificationEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	NotificationEntity

	abstraction.EntityJustCreated

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (NotificationEntityModel) TableName() string {
	return "notification"
}

type NotificationCountDataModel struct {
	CountTotal  int `json:"count_total"`
	CountRead   int `json:"count_read"`
	CountUnread int `json:"count_unread"`
}

func (m *NotificationEntityModel) BeforeCreate(tx *gorm.DB) (err error) {
	// m.CreatedAt = *general.Now()
	return
}
