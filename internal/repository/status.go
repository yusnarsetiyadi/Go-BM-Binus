package repository

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"
	"bm_binus/pkg/util/general"

	"gorm.io/gorm"
)

type Status interface {
	FindById(ctx *abstraction.Context, id int) (*model.StatusEntityModel, error)
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.StatusEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
}

type status struct {
	abstraction.Repository
}

func NewStatus(db *gorm.DB) *status {
	return &status{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *status) FindById(ctx *abstraction.Context, id int) (*model.StatusEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.StatusEntityModel
	err := conn.
		Where("id = ?", id, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *status) Find(ctx *abstraction.Context, no_paging bool) (data []*model.StatusEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "status", "")
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

func (r *status) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "status", "")
	var count model.StatusCountDataModel
	err = r.CheckTrx(ctx).
		Table("status").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}
