package repository

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"
	"bm_binus/pkg/util/general"

	"gorm.io/gorm"
)

type Request interface {
	Create(ctx *abstraction.Context, data *model.RequestEntityModel) *gorm.DB
	FindById(ctx *abstraction.Context, id int) (*model.RequestEntityModel, error)
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.RequestEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
	Update(ctx *abstraction.Context, data *model.RequestEntityModel) *gorm.DB
}

type request struct {
	abstraction.Repository
}

func NewRequest(db *gorm.DB) *request {
	return &request{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *request) Create(ctx *abstraction.Context, data *model.RequestEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *request) FindById(ctx *abstraction.Context, id int) (*model.RequestEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.RequestEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *request) Find(ctx *abstraction.Context, no_paging bool) (data []*model.RequestEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "request", "is_delete = @false")
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

func (r *request) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "request", "is_delete = @false")
	var count model.RequestCountDataModel
	err = r.CheckTrx(ctx).
		Table("request").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *request) Update(ctx *abstraction.Context, data *model.RequestEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}
