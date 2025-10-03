package model

import (
	"bm_binus/internal/abstraction"

	"gorm.io/gorm"
)

type FileEntity struct {
	RequestId int    `json:"request_id"`
	File      string `json:"file"`
	FileName  string `json:"file_name"`
	IsDelete  bool   `json:"is_delete"`
}

// FileEntityModel ...
type FileEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	FileEntity

	abstraction.Entity

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (FileEntityModel) TableName() string {
	return "file"
}

type FileCountDataModel struct {
	Count int `json:"count"`
}

func (m *FileEntityModel) BeforeUpdate(tx *gorm.DB) (err error) {
	// m.UpdatedAt = general.NowLocal()
	return
}

func (m *FileEntityModel) BeforeCreate(tx *gorm.DB) (err error) {
	// m.CreatedAt = *general.Now()
	return
}
