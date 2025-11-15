package repository

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"
	"bm_binus/pkg/util/general"
	"fmt"

	"gorm.io/gorm"
)

type Comment interface {
	FindByRequestId(ctx *abstraction.Context, request_id int, no_paging bool) (data []*model.CommentEntityModel, err error)
	CountByRequestId(ctx *abstraction.Context, request_id int) (data *int, err error)
	Create(ctx *abstraction.Context, data *model.CommentEntityModel) *gorm.DB
	FindById(ctx *abstraction.Context, id int) (*model.CommentEntityModel, error)
	Update(ctx *abstraction.Context, data *model.CommentEntityModel) *gorm.DB
}

type comment struct {
	abstraction.Repository
}

func NewComment(db *gorm.DB) *comment {
	return &comment{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *comment) FindByRequestId(ctx *abstraction.Context, request_id int, no_paging bool) (data []*model.CommentEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "comment", "is_delete = @false"+fmt.Sprintf(" AND request_id = %d", request_id))
	limit, offset := general.ProcessLimitOffset(ctx, no_paging)
	order := "created_at ASC"
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Limit(limit).
		Offset(offset).
		Preload("CreateBy").
		Preload("CreateBy.Role").
		Preload("UpdateBy").
		Preload("UpdateBy.Role").
		Find(&data).
		Error
	return
}

func (r *comment) CountByRequestId(ctx *abstraction.Context, request_id int) (data *int, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "comment", "is_delete = @false"+fmt.Sprintf(" AND request_id = %d", request_id))
	var count model.CommentCountDataModel
	err = r.CheckTrx(ctx).
		Table("comment").
		Select("COUNT(*) AS count").
		Where(where, whereParam).
		Find(&count).
		Error
	data = &count.Count
	return
}

func (r *comment) Create(ctx *abstraction.Context, data *model.CommentEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *comment) FindById(ctx *abstraction.Context, id int) (*model.CommentEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.CommentEntityModel
	err := conn.
		Where("id = ? AND is_delete = ?", id, false).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *comment) Update(ctx *abstraction.Context, data *model.CommentEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}
