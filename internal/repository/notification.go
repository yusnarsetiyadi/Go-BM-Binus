package repository

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"
	"bm_binus/pkg/util/general"
	"fmt"

	"gorm.io/gorm"
)

type Notification interface {
	Create(ctx *abstraction.Context, data *model.NotificationEntityModel) *gorm.DB
	FindByUserId(ctx *abstraction.Context, userId *int) (data []*model.NotificationEntityModel, err error)
	CountByUserId(ctx *abstraction.Context, userId *int) (countTotal *int, countRead *int, countUnread *int, err error)
	FindById(ctx *abstraction.Context, id int) (*model.NotificationEntityModel, error)
	Update(ctx *abstraction.Context, data *model.NotificationEntityModel) *gorm.DB
	FindByUserIdArr(ctx *abstraction.Context, userId *int) (data []*model.NotificationEntityModel, err error)
}

type notification struct {
	abstraction.Repository
}

func NewNotification(db *gorm.DB) *notification {
	return &notification{
		Repository: abstraction.Repository{
			Db: db,
		},
	}
}

func (r *notification) Create(ctx *abstraction.Context, data *model.NotificationEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Create(data)
}

func (r *notification) FindByUserId(ctx *abstraction.Context, userId *int) (data []*model.NotificationEntityModel, err error) {
	where, whereParam := general.ProcessWhereParam(ctx, "notification", fmt.Sprintf("user_id = %d", *userId))
	order := general.ProcessOrder(ctx)
	err = r.CheckTrx(ctx).
		Where(where, whereParam).
		Order(order).
		Find(&data).
		Error
	return
}

func (r *notification) CountByUserId(ctx *abstraction.Context, userId *int) (countTotal *int, countRead *int, countUnread *int, err error) {
	var count model.NotificationCountDataModel
	whereTotal, whereParamTotal := general.ProcessWhereParam(ctx, "notification", fmt.Sprintf("user_id = %d", *userId))
	err = r.CheckTrx(ctx).
		Table("notification").
		Select(`
			COUNT(*) AS count_total, 
			COUNT(CASE WHEN is_read = TRUE THEN 1 END) AS count_read,
			COUNT(CASE WHEN is_read = FALSE THEN 1 END) AS count_unread
		`).
		Where(whereTotal, whereParamTotal).
		Find(&count).
		Error
	countTotal = &count.CountTotal
	countRead = &count.CountRead
	countUnread = &count.CountUnread
	return
}

func (r *notification) FindById(ctx *abstraction.Context, id int) (*model.NotificationEntityModel, error) {
	conn := r.CheckTrx(ctx)

	var data model.NotificationEntityModel
	err := conn.
		Where("id = ?", id).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *notification) Update(ctx *abstraction.Context, data *model.NotificationEntityModel) *gorm.DB {
	return r.CheckTrx(ctx).Model(data).Where("id = ?", data.ID).Updates(data)
}

func (r *notification) FindByUserIdArr(ctx *abstraction.Context, userId *int) (data []*model.NotificationEntityModel, err error) {
	err = r.CheckTrx(ctx).
		Where("user_id = ? AND is_read = ?", userId, false).
		Find(&data).
		Error
	return
}
