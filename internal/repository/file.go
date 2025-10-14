package repository

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"
	"bm_binus/pkg/util/general"
	"fmt"

	"gorm.io/gorm"
)

type File interface {
	FindByRequestId(ctx *abstraction.Context, request_id int, no_paging bool) (data []*model.FileEntityModel, err error)
	CountByRequestId(ctx *abstraction.Context, request_id int) (data *int, err error)
	Create(ctx *abstraction.Context, data *model.FileEntityModel) *gorm.DB
	FindById(ctx *abstraction.Context, id int) (*model.FileEntityModel, error)
	Update(ctx *abstraction.Context, data *model.FileEntityModel) *gorm.DB
}

type file struct {
	abstraction.Repository
}

func NewFile(db *gorm.DB) *file {
	return &file{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *file) FindByRequestId(ctx *abstraction.Context, request_id int, no_paging bool) (data []*model.FileEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "file", "is_delete = @false"+fmt.Sprintf(" AND request_id = %d", request_id))
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := "created_at DESC"
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&data).
		Error
	return
}

func (r *file) CountByRequestId(ctx *abstraction.Context, request_id int) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "file", "is_delete = @false"+fmt.Sprintf(" AND request_id = %d", request_id))
	var count model.FileCountDataModel
	err = r.CheckTrx(ctx).
		Table("file").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *file) Create(ctx *abstraction.Context, data *model.FileEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *file) FindById(ctx *abstraction.Context, id int) (*model.FileEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.FileEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *file) Update(ctx *abstraction.Context, data *model.FileEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}
