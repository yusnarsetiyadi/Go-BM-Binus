package request

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/gdrive"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"bm_binus/pkg/ws"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/drive/v3"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.RequestCreateRequest) (map[string]interface{}, error)
}

type service struct {
	RequestRepository      repository.Request
	EventTypeRepository    repository.EventType
	FileRepository         repository.File
	NotificationRepository repository.Notification
	UserRepository         repository.User

	DB     *gorm.DB
	sDrive *drive.Service
	fDrive *drive.File
}

func NewService(f *factory.Factory) Service {
	return &service{
		RequestRepository:      f.RequestRepository,
		EventTypeRepository:    f.EventTypeRepository,
		FileRepository:         f.FileRepository,
		NotificationRepository: f.NotificationRepository,
		UserRepository:         f.UserRepository,

		DB:     f.Db,
		sDrive: f.GDrive.Service,
		fDrive: f.GDrive.FolderBM,
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

func (s *service) Create(ctx *abstraction.Context, payload *dto.RequestCreateRequest) (map[string]interface{}, error) {
	var (
		allFileUploaded []string
		sendNotifTo     []int
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_STAF {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		parsedEventDateStart, err := general.Parse("2006-01-02 15:04:05", payload.EventDateStart)
		if err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "err parse event date start:"+err.Error())
		}

		parsedEventDateEnd, err := general.Parse("2006-01-02 15:04:05", payload.EventDateEnd)
		if err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "err parse event date end:"+err.Error())
		}

		eventTypeData, err := s.EventTypeRepository.FindById(ctx, payload.EventTypeId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if eventTypeData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "event type not found")
		}

		modelRequest := &model.RequestEntityModel{
			Context: ctx,
			RequestEntity: model.RequestEntity{
				UserId:           ctx.Auth.ID,
				EventName:        payload.EventName,
				EventLocation:    payload.EventLocation,
				EventDateStart:   parsedEventDateStart,
				EventDateEnd:     parsedEventDateEnd,
				Description:      payload.Description,
				EventTypeId:      payload.EventTypeId,
				CountParticipant: payload.CountParticipant,
				StatusId:         constant.STATUS_ID_PENGAJUAN,
				IsDelete:         false,
			},
		}
		if err = s.RequestRepository.Create(ctx, modelRequest).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		for _, file := range payload.Files {
			f, err := file.Open()
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			defer f.Close()

			isFileAvailable, fullFileName := general.ValidateFileUpload(file.Filename)
			if !isFileAvailable {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), fmt.Sprintf("file format for %s is not approved", file.Filename))
			}

			newFile, err := gdrive.CreateFile(s.sDrive, fullFileName, "application/octet-stream", f, s.fDrive.Id)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			allFileUploaded = append(allFileUploaded, newFile.Id)

			modelFile := &model.FileEntityModel{
				Context: ctx,
				FileEntity: model.FileEntity{
					RequestId: modelRequest.ID,
					File:      newFile.Id,
					FileName:  newFile.Name,
					IsDelete:  false,
				},
			}
			if err := s.FileRepository.Create(ctx, modelFile).Error; err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		userBM, err := s.UserRepository.FindByRoleIdArr(ctx, constant.ROLE_ID_BM, true)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		for _, v := range userBM {
			err := SendNotif(s, ctx, "Notifikasi Event Baru!", modelRequest.EventName, v.ID, modelRequest.ID)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			sendNotifTo = append(sendNotifTo, v.ID)
		}

		return nil
	}); err != nil {
		for _, v := range allFileUploaded {
			errDel := gdrive.DeleteFile(s.sDrive, v)
			if errDel != nil {
				logrus.Error("error delete file for error trxmanager:", errDel.Error())
			}
		}
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
