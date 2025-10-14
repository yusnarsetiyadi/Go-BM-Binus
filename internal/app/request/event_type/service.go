package event_type

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/factory"
	"bm_binus/internal/repository"
	"bm_binus/pkg/util/response"
	"net/http"

	"gorm.io/gorm"
)

type Service interface {
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	// Create(ctx *abstraction.Context, payload *dto.EventTypeCreateRequest) (map[string]interface{}, error)
	// Delete(ctx *abstraction.Context, payload *dto.EventTypeDeleteByIDRequest) (map[string]interface{}, error)
	// Update(ctx *abstraction.Context, payload *dto.EventTypeUpdateRequest) (map[string]interface{}, error)
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
	data, err := s.EventTypeRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.EventTypeRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	var res []map[string]interface{} = nil
	for _, v := range data {
		res = append(res, map[string]interface{}{
			"id":       v.ID,
			"name":     v.Name,
			"priority": v.Priority,
		})
	}
	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}

// func (s *service) Create(ctx *abstraction.Context, payload *dto.EventTypeCreateRequest) (map[string]interface{}, error) {
// 	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
// 		if ctx.Auth.RoleID != constant.ROLE_ID_BM {
// 			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
// 		}

// 		modelEventType := &model.EventTypeEntityModel{
// 			Context: ctx,
// 			EventTypeEntity: model.EventTypeEntity{
// 				Name:     payload.Name,
// 				Priority: payload.Priority,
// 				IsDelete: false,
// 			},
// 		}
// 		if err := s.EventTypeRepository.Create(ctx, modelEventType).Error; err != nil {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		return nil
// 	}); err != nil {
// 		return nil, err
// 	}
// 	return map[string]interface{}{
// 		"message": "success create!",
// 	}, nil
// }

// func (s *service) Delete(ctx *abstraction.Context, payload *dto.EventTypeDeleteByIDRequest) (map[string]interface{}, error) {
// 	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
// 		taskTypeData, err := s.EventTypeRepository.FindById(ctx, payload.ID)
// 		if err != nil && err.Error() != "record not found" {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}
// 		if taskTypeData == nil {
// 			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task type not found")
// 		}

// 		newEventTypeData := new(model.EventTypeEntityModel)
// 		newEventTypeData.Context = ctx
// 		newEventTypeData.ID = taskTypeData.ID
// 		newEventTypeData.IsDelete = true

// 		if err = s.EventTypeRepository.Update(ctx, newEventTypeData).Error; err != nil {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		return nil
// 	}); err != nil {
// 		return nil, err
// 	}
// 	return map[string]interface{}{
// 		"message": "success delete!",
// 	}, nil
// }

// func (s *service) Update(ctx *abstraction.Context, payload *dto.EventTypeUpdateRequest) (map[string]interface{}, error) {
// 	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
// 		taskTypeData, err := s.EventTypeRepository.FindById(ctx, payload.ID)
// 		if err != nil && err.Error() != "record not found" {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}
// 		if taskTypeData == nil {
// 			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task type not found")
// 		}

// 		newEventTypeData := new(model.EventTypeEntityModel)
// 		newEventTypeData.Context = ctx
// 		newEventTypeData.ID = payload.ID
// 		if payload.Name != nil {
// 			newEventTypeData.Name = *payload.Name
// 		}
// 		if payload.EventId != nil {
// 			taskData, err := s.EventRepository.FindById(ctx, *payload.EventId)
// 			if err != nil && err.Error() != "record not found" {
// 				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 			}
// 			if taskData == nil {
// 				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task not found")
// 			}
// 			newEventTypeData.EventId = *payload.EventId
// 		}

// 		if err = s.EventTypeRepository.Update(ctx, newEventTypeData).Error; err != nil {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		return nil
// 	}); err != nil {
// 		return nil, err
// 	}
// 	return map[string]interface{}{
// 		"message": "success update!",
// 	}, nil
// }
