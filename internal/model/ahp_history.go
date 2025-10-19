package model

import (
	"bm_binus/internal/abstraction"

	"gorm.io/gorm"
)

type AhpHistoryEntity struct {
	Kriteria             string `json:"kriteria"`
	KriteriaComparison   string `json:"kriteria_comparison"`
	Alternatif           string `json:"alternatif"`
	AlternatifComparison string `json:"alternatif_comparison"`
	PriorityGlobal       string `json:"priority_global"`
	ReferenceRequest     int    `json:"reference_request"`
	IsDelete             bool   `json:"is_delete"`
}

// AhpHistoryEntityModel ...
type AhpHistoryEntityModel struct {
	ID int `json:"id" param:"id" form:"id" validate:"number,min=1" gorm:"primaryKey;autoIncrement;"`

	// entity
	AhpHistoryEntity

	abstraction.EntityJustCreated

	// context
	Context *abstraction.Context `json:"-" gorm:"-"`
}

// TableName ...
func (AhpHistoryEntityModel) TableName() string {
	return "ahp_history"
}

type AhpHistoryCountDataModel struct {
	Count int `json:"count"`
}

func (m *AhpHistoryEntityModel) BeforeCreate(tx *gorm.DB) (err error) {
	// m.CreatedAt = *general.Now()
	return
}
