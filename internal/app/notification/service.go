package notification

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	"bm_binus/internal/repository"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type Service interface {
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	SetRead(ctx *abstraction.Context, payload *dto.NotificationSetReadRequest) (map[string]interface{}, error)
	SetReadAll(ctx *abstraction.Context) (map[string]interface{}, error)
}

type service struct {
	NotificationRepository repository.Notification

	DB *gorm.DB
}

func NewService(f *factory.Factory) Service {
	return &service{
		NotificationRepository: f.NotificationRepository,

		DB: f.Db,
	}
}

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	var res []map[string]interface{} = nil
	data, err := s.NotificationRepository.FindByUserId(ctx, &ctx.Auth.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	countTotal, countRead, countUnread, err := s.NotificationRepository.CountByUserId(ctx, &ctx.Auth.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	for _, v := range data {
		res = append(res, map[string]interface{}{
			"id":         v.ID,
			"title":      v.Title,
			"message":    v.Message,
			"is_read":    v.IsRead,
			"user_id":    v.UserId,
			"request_id": v.RequestId,
			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
		})
	}
	return map[string]interface{}{
		"count_total":  countTotal,
		"count_read":   countRead,
		"count_unread": countUnread,
		"data":         res,
	}, nil
}

func (s *service) SetRead(ctx *abstraction.Context, payload *dto.NotificationSetReadRequest) (map[string]interface{}, error) {
	var (
		notificationData = new(model.NotificationEntityModel)
		err              error
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		notificationData, err = s.NotificationRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if notificationData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "notification not found")
		}

		newNotificationData := new(model.NotificationEntityModel)
		newNotificationData.Context = ctx
		newNotificationData.ID = notificationData.ID
		newNotificationData.IsRead = true

		if err = s.NotificationRepository.Update(ctx, newNotificationData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message":    "success set read!",
		"user_id":    notificationData.UserId,
		"request_id": notificationData.RequestId,
	}, nil
}

func (s *service) SetReadAll(ctx *abstraction.Context) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		notificationData, err := s.NotificationRepository.FindByUserIdArr(ctx, &ctx.Auth.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		for _, v := range notificationData {
			newNotificationData := new(model.NotificationEntityModel)
			newNotificationData.Context = ctx
			newNotificationData.ID = v.ID
			newNotificationData.IsRead = true
			if err = s.NotificationRepository.Update(ctx, newNotificationData).Error; err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success set read all!",
	}, nil
}
