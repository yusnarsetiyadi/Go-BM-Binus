package model

import (
	"bm_binus/internal/abstraction"

	"gorm.io/gorm"
)

type UserEntity struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleId   int    `json:"role_id"`
	IsDelete bool   `json:"is_delete"`
}

// UserEntityModel ...
type UserEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	UserEntity

	abstraction.Entity

	Role RoleEntityModel `json:"role" gorm:"foreignKey:RoleId"`

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (UserEntityModel) TableName() string {
	return "user"
}

type UserCountDataModel struct {
	Count int `json:"count"`
}

func (m *UserEntityModel) BeforeUpdate(tx *gorm.DB) (err error) {
	// m.UpdatedAt = general.NowLocal()
	return
}

func (m *UserEntityModel) BeforeCreate(tx *gorm.DB) (err error) {
	// m.CreatedAt = *general.Now()
	return
}
