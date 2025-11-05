package dashboard

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"net/http"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service interface {
	GetDashboard(ctx *abstraction.Context, payload *dto.GetDashboardRequest) (map[string]interface{}, error)
}

type service struct {
	UserRepository      repository.User
	RequestRepository   repository.Request
	DashboardRepository repository.Dashboard
	StatusRepository    repository.Status
	EventTypeRepository repository.EventType

	DB      *gorm.DB
	DbRedis *redis.Client
}

func NewService(f *factory.Factory) Service {
	return &service{
		UserRepository:      f.UserRepository,
		RequestRepository:   f.RequestRepository,
		DashboardRepository: f.DashboardRepository,
		StatusRepository:    f.StatusRepository,
		EventTypeRepository: f.EventTypeRepository,

		DB:      f.Db,
		DbRedis: f.DbRedis,
	}
}

func (s *service) GetDashboard(ctx *abstraction.Context, payload *dto.GetDashboardRequest) (map[string]interface{}, error) {
	countAllUsers, err := s.UserRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	countUsePriority, err := general.GetUsePriorityCount(s.DbRedis)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	countAllRequest, err := s.RequestRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	var userId *int = nil
	if ctx.Auth.RoleID == constant.ROLE_ID_STAF {
		userId = &ctx.Auth.ID
	}
	getDashboardByStatus, err := s.DashboardRepository.GetByStatus(ctx, userId)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	getDashboardByEventType, err := s.DashboardRepository.GetByEventType(ctx, userId)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	dataStatus, err := s.StatusRepository.Find(ctx, true)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	dataEventType, err := s.EventTypeRepository.Find(ctx, true)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	// =============================
	// STATUS HANDLER
	// =============================
	statusCountMap := make(map[int]int)
	for _, item := range getDashboardByStatus {
		statusCountMap[item.StatusID] = item.Total
	}

	var resDashboardByStatus []map[string]interface{}
	for _, s := range dataStatus {
		total := 0
		if val, ok := statusCountMap[s.ID]; ok {
			total = val
		}
		resDashboardByStatus = append(resDashboardByStatus, map[string]interface{}{
			"status": s.Name,
			"total":  total,
		})
	}

	// =============================
	// EVENT TYPE HANDLER
	// =============================
	eventTypeCountMap := make(map[int]int)
	for _, item := range getDashboardByEventType {
		eventTypeCountMap[item.EventTypeID] = item.Total
	}

	var resDashboardByEventType []map[string]interface{}
	for _, e := range dataEventType {
		total := 0
		if val, ok := eventTypeCountMap[e.ID]; ok {
			total = val
		}
		resDashboardByEventType = append(resDashboardByEventType, map[string]interface{}{
			"event_type": e.Name,
			"total":      total,
		})
	}

	res := make(map[string]interface{})

	switch payload.RoleId {
	case constant.ROLE_ID_STAF:
		res["chart_by_status"] = resDashboardByStatus
		res["chart_by_event_type"] = resDashboardByEventType

	case constant.ROLE_ID_BM:
		res["count_user"] = countAllUsers
		res["count_use_priority"] = countUsePriority
		res["count_request"] = countAllRequest
		res["chart_by_status"] = resDashboardByStatus
		res["chart_by_event_type"] = resDashboardByEventType

	case constant.ROLE_ID_ADMIN:
		res["chart_by_status"] = resDashboardByStatus
		res["chart_by_event_type"] = resDashboardByEventType
	}

	return map[string]interface{}{
		"data": res,
	}, nil
}
