package repository

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"
	"bm_binus/pkg/util/general"

	"gorm.io/gorm"
)

type EventType interface {
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.EventTypeEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
	Create(ctx *abstraction.Context, data *model.EventTypeEntityModel) *gorm.DB
	FindById(ctx *abstraction.Context, id int) (*model.EventTypeEntityModel, error)
	Update(ctx *abstraction.Context, data *model.EventTypeEntityModel) *gorm.DB
}

type event_type struct {
	abstraction.Repository
}

func NewEventType(db *gorm.DB) *event_type {
	return &event_type{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *event_type) Find(ctx *abstraction.Context, no_paging bool) (data []*model.EventTypeEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "event_type", "is_delete = @false")
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := general.ProcessOrder(ctx)
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&data).
		Error
	return
}

func (r *event_type) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "event_type", "is_delete = @false")
	var count model.EventTypeCountDataModel
	err = r.CheckTrx(ctx).
		Table("event_type").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *event_type) Create(ctx *abstraction.Context, data *model.EventTypeEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *event_type) FindById(ctx *abstraction.Context, id int) (*model.EventTypeEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.EventTypeEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *event_type) Update(ctx *abstraction.Context, data *model.EventTypeEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}
