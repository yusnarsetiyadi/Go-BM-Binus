package model

import (
	"bm_binus/internal/abstraction"

	"gorm.io/gorm"
)

type CommentEntity struct {
	RequestId int    `json:"request_id"`
	Comment   string `json:"comment"`
	IsDelete  bool   `json:"is_delete"`
}

// CommentEntityModel ...
type CommentEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	CommentEntity

	abstraction.EntityWithBy

	CreateBy UserEntityModel `json:"create_by" gorm:"foreignKey:CreatedBy"`
	UpdateBy UserEntityModel `json:"update_by" gorm:"foreignKey:UpdatedBy"`

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (CommentEntityModel) TableName() string {
	return "comment"
}

type CommentCountDataModel struct {
	Count int `json:"count"`
}

func (m *CommentEntityModel) BeforeUpdate(tx *gorm.DB) (err error) {
	m.UpdatedBy = &m.Context.Auth.ID
	return
}

func (m *CommentEntityModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.CreatedBy = m.Context.Auth.ID
	return
}
