package event_type

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type Service interface {
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	Create(ctx *abstraction.Context, payload *dto.EventTypeCreateRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.EventTypeDeleteByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.EventTypeUpdateRequest) (map[string]interface{}, error)
}

type service struct {
	EventTypeRepository repository.EventType

	DB *gorm.DB
}

func NewService(f *factory.Factory) Service {
	return &service{
		EventTypeRepository: f.EventTypeRepository,

		DB: f.Db,
	}
}

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	var (
		res           []map[string]interface{} = nil
		priorityCount                          = make(map[int]int)
		hasDuplicate                           = false
	)
	data, err := s.EventTypeRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.EventTypeRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	for _, v := range data {
		res = append(res, map[string]interface{}{
			"id":       v.ID,
			"name":     v.Name,
			"priority": v.Priority,
		})
		priorityCount[v.Priority]++
		if priorityCount[v.Priority] > 1 {
			hasDuplicate = true
		}
	}

	resReturn := map[string]interface{}{
		"count": count,
		"data":  res,
	}
	switch {
	case hasDuplicate:
		resReturn["info"] = "Terdapat nilai prioritas yang duplikat, segera perbaiki!"
	default:
		resReturn["info"] = "Nilai prioritas sudah sesuai."
	}

	return resReturn, nil
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.EventTypeCreateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_BM {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		modelEventType := &model.EventTypeEntityModel{
			Context: ctx,
			EventTypeEntity: model.EventTypeEntity{
				Name:     payload.Name,
				Priority: payload.Priority,
				IsDelete: false,
			},
		}
		if err := s.EventTypeRepository.Create(ctx, modelEventType).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success create!",
	}, nil
}

func (s *service) Delete(ctx *abstraction.Context, payload *dto.EventTypeDeleteByIDRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_BM {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		eventTypeData, err := s.EventTypeRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if eventTypeData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "event type not found")
		}

		newEventTypeData := new(model.EventTypeEntityModel)
		newEventTypeData.Context = ctx
		newEventTypeData.ID = eventTypeData.ID
		newEventTypeData.IsDelete = true

		if err = s.EventTypeRepository.Update(ctx, newEventTypeData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}

func (s *service) Update(ctx *abstraction.Context, payload *dto.EventTypeUpdateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		eventTypeData, err := s.EventTypeRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if eventTypeData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "event type not found")
		}

		newEventTypeData := new(model.EventTypeEntityModel)
		newEventTypeData.Context = ctx
		newEventTypeData.ID = payload.ID
		if payload.Name != nil {
			newEventTypeData.Name = *payload.Name
		}
		if payload.Priority != nil {
			newEventTypeData.Priority = *payload.Priority
		}

		if err = s.EventTypeRepository.Update(ctx, newEventTypeData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success update!",
	}, nil
}
