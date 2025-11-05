package repository

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"

	"gorm.io/gorm"
)

type Dashboard interface {
	GetByStatus(ctx *abstraction.Context, user_id *int) (data []*model.RequestCountByStatus, err error)
	GetByEventType(ctx *abstraction.Context, user_id *int) (data []*model.RequestCountByEventType, err error)
}

type dashboard struct {
	abstraction.Repository
}

func NewDashboard(db *gorm.DB) *dashboard {
	return &dashboard{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *dashboard) GetByStatus(ctx *abstraction.Context, user_id *int) (data []*model.RequestCountByStatus, err error) {
	conn := r.CheckTrx(ctx)
	query := conn.Table("request AS r").
		Where("r.is_delete = ?", false)

	if user_id != nil {
		query = query.Where("r.user_id = ?", *user_id)
	}

	err = query.Select("r.status_id, COUNT(r.id) AS total").
		Group("r.status_id").
		Scan(&data).Error

	return
}

func (r *dashboard) GetByEventType(ctx *abstraction.Context, user_id *int) (data []*model.RequestCountByEventType, err error) {
	conn := r.CheckTrx(ctx)
	query := conn.Table("request AS r").
		Where("r.is_delete = ?", false)

	if user_id != nil {
		query = query.Where("r.user_id = ?", *user_id)
	}

	err = query.Select("r.event_type_id, COUNT(r.id) AS total").
		Group("r.event_type_id").
		Scan(&data).Error

	return
}
