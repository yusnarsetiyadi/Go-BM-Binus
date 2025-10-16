package comment

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"bm_binus/pkg/ws"
	"errors"
	"net/http"
	"slices"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.CommentCreateRequest) (map[string]interface{}, error)
	// FindByRequestId(ctx *abstraction.Context, payload *dto.CommentFindByRequestIDRequest) (map[string]interface{}, error)
	// Delete(ctx *abstraction.Context, payload *dto.CommentDeleteByIDRequest) (map[string]interface{}, error)
	// Update(ctx *abstraction.Context, payload *dto.CommentUpdateRequest) (map[string]interface{}, error)
}

type service struct {
	CommentRepository      repository.Comment
	RequestRepository      repository.Request
	NotificationRepository repository.Notification
	UserRepository         repository.User

	DB *gorm.DB
}

func NewService(f *factory.Factory) Service {
	return &service{
		CommentRepository:      f.CommentRepository,
		RequestRepository:      f.RequestRepository,
		NotificationRepository: f.NotificationRepository,
		UserRepository:         f.UserRepository,

		DB: f.Db,
	}
}

func SendNotif(s *service, ctx *abstraction.Context, title string, message string, userId int, requestId int) error {
	modelNotification := &model.NotificationEntityModel{
		Context: ctx,
		NotificationEntity: model.NotificationEntity{
			Title:     title,
			Message:   message,
			IsRead:    false,
			UserId:    userId,
			RequestId: requestId,
		},
	}
	if err := s.NotificationRepository.Create(ctx, modelNotification).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.CommentCreateRequest) (map[string]interface{}, error) {
	var (
		sendNotifTo      []int
		statusesForAdmin = []int{
			constant.STATUS_ID_PROSES,
			constant.STATUS_ID_FINALISASI,
			constant.STATUS_ID_SELESAI,
		}
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		requestData, err := s.RequestRepository.FindById(ctx, payload.RequestId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if requestData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request not found")
		}

		modelComment := &model.CommentEntityModel{
			Context: ctx,
			CommentEntity: model.CommentEntity{
				RequestId: payload.RequestId,
				Comment:   payload.Comment,
				IsDelete:  false,
			},
		}
		if err := s.CommentRepository.Create(ctx, modelComment).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		newRequestData := new(model.RequestEntityModel)
		newRequestData.Context = ctx
		newRequestData.ID = payload.RequestId
		newRequestData.UpdatedAt = general.NowLocal()
		if err = s.RequestRepository.Update(ctx, newRequestData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userBM, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_BM, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		userAdmin, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_ADMIN, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		addIDs := func(users []*model.UserEntityModel) {
			for _, v := range users {
				sendNotifTo = append(sendNotifTo, v.ID)
			}
		}
		switch ctx.Auth.RoleID {
		case constant.ROLE_ID_STAF:
			addIDs(userBM)
			if slices.Contains(statusesForAdmin, requestData.StatusId) {
				addIDs(userAdmin)
			}
		case constant.ROLE_ID_BM:
			sendNotifTo = append(sendNotifTo, requestData.UserId)
			if slices.Contains(statusesForAdmin, requestData.StatusId) {
				addIDs(userAdmin)
			}
		case constant.ROLE_ID_ADMIN:
			sendNotifTo = append(sendNotifTo, requestData.UserId)
			addIDs(userBM)
		}

		for _, v := range sendNotifTo {
			err = SendNotif(s, ctx, "Notifikasi Komentar Baru!", payload.Comment, v, requestData.ID)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, v := range general.RemoveDuplicateArrayInt(sendNotifTo) {
		if err := ws.PublishNotificationWithoutTransaction(v, s.DB, ctx); err != nil {
			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
	}

	return map[string]interface{}{
		"message": "success create!",
	}, nil
}

// func (s *service) FindByRequestId(ctx *abstraction.Context, payload *dto.CommentFindByRequestIDRequest) (map[string]interface{}, error) {
// 	var res []map[string]interface{} = nil

// 	requestData, err := s.RequestRepository.FindById(ctx, payload.RequestId)
// 	if err != nil && err.Error() != "record not found" {
// 		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 	}
// 	if requestData == nil {
// 		return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request not found")
// 	}

// 	data, err := s.CommentRepository.FindByRequestId(ctx, payload.RequestId, false)
// 	if err != nil && err.Error() != "record not found" {
// 		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 	}
// 	count, err := s.CommentRepository.CountByRequestId(ctx, payload.RequestId)
// 	if err != nil && err.Error() != "record not found" {
// 		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 	}

// 	for _, v := range data {
// 		res = append(res, map[string]interface{}{
// 			"id":         v.ID,
// 			"request_id": v.RequestId,
// 			"comment":    v.Comment,
// 			"is_delete":  v.IsDelete,
// 			"is_history": v.IsHistory,
// 			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
// 			"updated_at": general.FormatWithZWithoutChangingTime(*v.UpdatedAt),
// 			"created_by": map[string]interface{}{
// 				"id":    v.CreateBy.ID,
// 				"name":  v.CreateBy.Name,
// 				"email": v.CreateBy.Email,
// 			},
// 			"updated_by": map[string]interface{}{
// 				"id":    v.UpdateBy.ID,
// 				"name":  v.UpdateBy.Name,
// 				"email": v.UpdateBy.Email,
// 			},
// 		})
// 	}

// 	return map[string]interface{}{
// 		"count": count,
// 		"data":  res,
// 	}, nil
// }

// func (s *service) Delete(ctx *abstraction.Context, payload *dto.CommentDeleteByIDRequest) (map[string]interface{}, error) {
// 	var sendNotifTo []int = nil
// 	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
// 		requestCommentData, err := s.CommentRepository.FindById(ctx, payload.ID)
// 		if err != nil && err.Error() != "record not found" {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}
// 		if requestCommentData == nil {
// 			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request comment not found")
// 		}

// 		requestData, err := s.RequestRepository.FindById(ctx, requestCommentData.RequestId)
// 		if err != nil && err.Error() != "record not found" {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		userLogin, err := s.UserRepository.FindById(ctx, ctx.Auth.ID)
// 		if err != nil && err.Error() != "record not found" {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		userAdmin, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_ADMIN, true)
// 		if err != nil && err.Error() != "record not found" {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		var assignedMember []*model.UserEntityModel
// 		if requestData.AssignToUser != nil {
// 			assignToUserArr := general.StringToArrayInt(requestData.AssignToUser)
// 			for _, v := range assignToUserArr {
// 				dataUser, err := s.UserRepository.FindById(ctx, v)
// 				if err != nil && err.Error() != "record not found" {
// 					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 				}
// 				if dataUser != nil {
// 					assignedMember = append(assignedMember, dataUser)
// 				}
// 			}
// 		}

// 		newCommentData := new(model.CommentEntityModel)
// 		newCommentData.Context = ctx
// 		newCommentData.ID = payload.ID
// 		newCommentData.IsDelete = true
// 		if err = s.CommentRepository.Update(ctx, newCommentData).Error; err != nil {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		newRequestData := new(model.RequestEntityModel)
// 		newRequestData.Context = ctx
// 		newRequestData.ID = requestCommentData.RequestId
// 		newRequestData.UpdatedAt = general.NowLocal()
// 		if err = s.RequestRepository.Update(ctx, newRequestData).Error; err != nil {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		for _, v := range userAdmin {
// 			if v.ID != userLogin.ID {
// 				modelNotifikasi := new(model.NotifikasiEntityModel)
// 				modelNotifikasi.Context = ctx
// 				modelNotifikasi.Title = fmt.Sprintf("%s menghapus komentar pada tugas (%s)", userLogin.Name, requestData.Title)
// 				modelNotifikasi.Message = fmt.Sprintf("Komentar yang dihapus: %s", requestCommentData.Comment)
// 				modelNotifikasi.IsRead = false
// 				modelNotifikasi.UserId = v.ID
// 				modelNotifikasi.RequestId = requestData.ID
// 				if err := s.NotifikasiRepository.Create(ctx, modelNotifikasi).Error; err != nil {
// 					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 				}
// 				sendNotifTo = append(sendNotifTo, v.ID)
// 			}
// 		}

// 		if requestData.CreateBy.RoleId != constant.ROLE_ID_ADMIN && requestData.CreateBy.ID != userLogin.ID {
// 			modelNotifikasi := new(model.NotifikasiEntityModel)
// 			modelNotifikasi.Context = ctx
// 			modelNotifikasi.Title = fmt.Sprintf("%s menghapus komentar pada tugas yang anda buat", userLogin.Name)
// 			modelNotifikasi.Message = fmt.Sprintf("Komentar yang dihapus: %s", requestCommentData.Comment)
// 			modelNotifikasi.IsRead = false
// 			modelNotifikasi.UserId = requestData.CreateBy.ID
// 			modelNotifikasi.RequestId = requestData.ID
// 			if err := s.NotifikasiRepository.Create(ctx, modelNotifikasi).Error; err != nil {
// 				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 			}
// 			sendNotifTo = append(sendNotifTo, requestData.CreateBy.ID)
// 		}

// 		for _, v := range assignedMember {
// 			if v.Role.ID != constant.ROLE_ID_ADMIN && v.ID != requestData.CreateBy.ID && v.ID != userLogin.ID {
// 				modelNotifikasi := new(model.NotifikasiEntityModel)
// 				modelNotifikasi.Context = ctx
// 				modelNotifikasi.Title = fmt.Sprintf("%s menghapus komentar pada tugas (%s)", userLogin.Name, requestData.Title)
// 				modelNotifikasi.Message = fmt.Sprintf("Komentar yang dihapus: %s", requestCommentData.Comment)
// 				modelNotifikasi.IsRead = false
// 				modelNotifikasi.UserId = v.ID
// 				modelNotifikasi.RequestId = requestData.ID
// 				if err := s.NotifikasiRepository.Create(ctx, modelNotifikasi).Error; err != nil {
// 					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 				}
// 				sendNotifTo = append(sendNotifTo, v.ID)
// 			}
// 		}

// 		return nil
// 	}); err != nil {
// 		return nil, err
// 	}

// 	for _, v := range general.RemoveDuplicateArrayInt(sendNotifTo) {
// 		if err := ws.PublishNotificationWithoutTransaction(v, s.DB, ctx); err != nil {
// 			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}
// 	}

// 	return map[string]interface{}{
// 		"message": "success delete!",
// 	}, nil
// }

// func (s *service) Update(ctx *abstraction.Context, payload *dto.CommentUpdateRequest) (map[string]interface{}, error) {
// 	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
// 		requestCommentData, err := s.CommentRepository.FindById(ctx, payload.ID)
// 		if err != nil && err.Error() != "record not found" {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}
// 		if requestCommentData == nil {
// 			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request comment not found")
// 		}

// 		newCommentData := new(model.CommentEntityModel)
// 		newCommentData.Context = ctx
// 		newCommentData.ID = payload.ID
// 		if payload.Comment != nil {
// 			newCommentData.Comment = *payload.Comment
// 		}
// 		if err = s.CommentRepository.Update(ctx, newCommentData).Error; err != nil {
// 			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
// 		}

// 		newRequestData := new(model.RequestEntityModel)
// 		newRequestData.Context = ctx
// 		newRequestData.ID = requestCommentData.RequestId
// 		newRequestData.UpdatedAt = general.NowLocal()
// 		if err = s.RequestRepository.Update(ctx, newRequestData).Error; err != nil {
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
