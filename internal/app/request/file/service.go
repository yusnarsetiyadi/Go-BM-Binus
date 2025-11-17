package file

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
	"slices"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/drive/v3"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.FileCreateRequest) (map[string]interface{}, error)
	FindByRequestId(ctx *abstraction.Context, payload *dto.FileFindByRequestIDRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.FileDeleteByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.FileUpdateRequest) (map[string]interface{}, error)
}

type service struct {
	FileRepository         repository.File
	RequestRepository      repository.Request
	NotificationRepository repository.Notification
	UserRepository         repository.User

	DB     *gorm.DB
	sDrive *drive.Service
	fDrive *drive.File
}

func NewService(f *factory.Factory) Service {
	return &service{
		FileRepository:         f.FileRepository,
		RequestRepository:      f.RequestRepository,
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

func (s *service) Create(ctx *abstraction.Context, payload *dto.FileCreateRequest) (map[string]interface{}, error) {
	var (
		allFileUploaded  []string
		allFileName      []string
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
			allFileName = append(allFileName, newFile.Name)

			modelFile := &model.FileEntityModel{
				Context: ctx,
				FileEntity: model.FileEntity{
					RequestId: requestData.ID,
					File:      newFile.Id,
					FileName:  newFile.Name,
					IsDelete:  false,
				},
			}
			if err := s.FileRepository.Create(ctx, modelFile).Error; err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
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
			err = SendNotif(s, ctx, "Berkas Baru!", general.FormatNamesFromArray(allFileName), v, requestData.ID)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
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

func (s *service) FindByRequestId(ctx *abstraction.Context, payload *dto.FileFindByRequestIDRequest) (map[string]interface{}, error) {
	var res []map[string]interface{} = nil

	requestData, err := s.RequestRepository.FindById(ctx, payload.RequestId)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if requestData == nil {
		return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "request not found")
	}

	data, err := s.FileRepository.FindByRequestId(ctx, payload.RequestId, true)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.FileRepository.CountByRequestId(ctx, payload.RequestId)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	for _, v := range data {
		file, err := gdrive.GetFile(s.sDrive, v.File)
		if err != nil {
			return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "file not found")
		}

		res = append(res, map[string]interface{}{
			"id":         v.ID,
			"request_id": v.RequestId,
			"file": map[string]interface{}{
				"view_saved": general.ConvertLinkToFileSaved(file.WebContentLink, file.Name, file.FileExtension),
				"view":       "https://lh3.googleusercontent.com/d/" + v.File,
				"content":    file.WebContentLink,
				"ext":        file.FileExtension,
				"name":       file.Name,
			},
			"file_name":  v.FileName,
			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*v.UpdatedAt),
		})
	}

	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}

func (s *service) Delete(ctx *abstraction.Context, payload *dto.FileDeleteByIDRequest) (map[string]interface{}, error) {
	var (
		fileNameDelete   string
		sendNotifTo      []int
		statusesForAdmin = []int{
			constant.STATUS_ID_PROSES,
			constant.STATUS_ID_FINALISASI,
			constant.STATUS_ID_SELESAI,
		}
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		fileData, err := s.FileRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if fileData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "file not found")
		}

		requestData, err := s.RequestRepository.FindById(ctx, fileData.RequestId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		fileNameDelete = fileData.FileName
		newFileData := new(model.FileEntityModel)
		newFileData.Context = ctx
		newFileData.ID = fileData.ID
		newFileData.IsDelete = true
		if err = s.FileRepository.Update(ctx, newFileData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		newRequestData := new(model.RequestEntityModel)
		newRequestData.Context = ctx
		newRequestData.ID = fileData.RequestId
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
			err = SendNotif(s, ctx, "Berkas dihapus!", fileNameDelete, v, requestData.ID)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}

func (s *service) Update(ctx *abstraction.Context, payload *dto.FileUpdateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		fileData, err := s.FileRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if fileData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "file not found")
		}

		newFileData := new(model.FileEntityModel)
		newFileData.Context = ctx
		newFileData.ID = payload.ID
		if payload.Name != nil {
			_, err := gdrive.RenameFile(s.sDrive, fileData.File, *payload.Name)
			if err != nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "failed rename file: "+err.Error())
			}
			newFileData.FileName = *payload.Name
		}
		if err = s.FileRepository.Update(ctx, newFileData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		newRequestData := new(model.RequestEntityModel)
		newRequestData.Context = ctx
		newRequestData.ID = fileData.RequestId
		newRequestData.UpdatedAt = general.NowLocal()
		if err = s.RequestRepository.Update(ctx, newRequestData).Error; err != nil {
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
