package repository

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"
	"bm_binus/pkg/util/general"

	"gorm.io/gorm"
)

type AhpHistory interface {
	Create(ctx *abstraction.Context, data *model.AhpHistoryEntityModel) *gorm.DB
	FindById(ctx *abstraction.Context, id int) (*model.AhpHistoryEntityModel, error)
	Find(ctx *abstraction.Context, no_paging bool) (data []*model.AhpHistoryEntityModel, err error)
	Count(ctx *abstraction.Context) (data *int, err error)
	Update(ctx *abstraction.Context, data *model.AhpHistoryEntityModel) *gorm.DB
}

type ahp_history struct {
	abstraction.Repository
}

func NewAhpHistory(db *gorm.DB) *ahp_history {
	return &ahp_history{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *ahp_history) Create(ctx *abstraction.Context, data *model.AhpHistoryEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *ahp_history) FindById(ctx *abstraction.Context, id int) (*model.AhpHistoryEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.AhpHistoryEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *ahp_history) Find(ctx *abstraction.Context, no_paging bool) (data []*model.AhpHistoryEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "ahp_history", "is_delete = @false")
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

func (r *ahp_history) Count(ctx *abstraction.Context) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "ahp_history", "is_delete = @false")
	var count model.AhpHistoryCountDataModel
	err = r.CheckTrx(ctx).
		Table("ahp_history").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *ahp_history) Update(ctx *abstraction.Context, data *model.AhpHistoryEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}
