package contact

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/gomail"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"bm_binus/pkg/ws"
	"bytes"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.UserCreateRequest) (map[string]interface{}, error)
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	FindById(ctx *abstraction.Context, payload *dto.UserFindByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.UserUpdateRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.UserDeleteByIDRequest) (map[string]interface{}, error)
	ChangePassword(ctx *abstraction.Context, payload *dto.UserChangePasswordRequest) (map[string]interface{}, error)
	ResetPassword(ctx *abstraction.Context, payload *dto.UserResetPasswordRequest) (map[string]interface{}, error)
	GetUserInfo(ctx *abstraction.Context) (map[string]interface{}, error)
	Export(ctx *abstraction.Context) (string, *bytes.Buffer, error)
}

type service struct {
	UserRepository repository.User

	DB      *gorm.DB
	DbRedis *redis.Client
}

func NewService(f *factory.Factory) Service {
	return &service{
		UserRepository: f.UserRepository,

		DB:      f.Db,
		DbRedis: f.DbRedis,
	}
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.UserCreateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_ADMIN {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		userEmail, err := s.UserRepository.FindByEmail(ctx, payload.Email)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userEmail != nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "email already exist")
		}

		passwordString := general.GeneratePassword(8, 1, 1, 1, 1)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordString), bcrypt.DefaultCost)
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		modelUser := &model.UserEntityModel{
			Context: ctx,
			UserEntity: model.UserEntity{
				Name:      payload.Name,
				Email:     payload.Email,
				Password:  string(hashedPassword),
				RoleId:    payload.RoleId,
				IsDelete:  false,
				IsLocked:  false,
				LoginFrom: "",
			},
		}
		if err = s.UserRepository.Create(ctx, modelUser).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		if err = gomail.SendMail(payload.Email, "Welcome to SelarasHomeId (Login Information)", general.ParseTemplateEmailToHtml("./assets/html/email/notif_login_info.html", struct {
			NAME     string
			EMAIL    string
			PASSWORD string
			LINK     string
		}{
			NAME:     payload.Name,
			EMAIL:    payload.Email,
			PASSWORD: passwordString,
			LINK:     constant.BASE_URL,
		})); err != nil {
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

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	var res []map[string]interface{} = nil
	data, err := s.UserRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.UserRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	for _, v := range data {
		res = append(res, map[string]interface{}{
			"id":         v.ID,
			"name":       v.Name,
			"email":      v.Email,
			"is_delete":  v.IsDelete,
			"is_locked":  v.IsLocked,
			"login_from": v.LoginFrom,
			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*v.UpdatedAt),
			"role": map[string]interface{}{
				"id":   v.Role.ID,
				"name": v.Role.Name,
			},
		})
	}
	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}

func (s *service) FindById(ctx *abstraction.Context, payload *dto.UserFindByIDRequest) (map[string]interface{}, error) {
	var res map[string]interface{} = nil
	data, err := s.UserRepository.FindById(ctx, payload.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if data != nil {
		res = map[string]interface{}{
			"id":         data.ID,
			"name":       data.Name,
			"email":      data.Email,
			"is_delete":  data.IsDelete,
			"is_locked":  data.IsLocked,
			"login_from": data.LoginFrom,
			"created_at": general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
			"role": map[string]interface{}{
				"id":   data.Role.ID,
				"name": data.Role.Name,
			},
		}

	}
	return map[string]interface{}{
		"data": res,
	}, nil
}

func (s *service) Update(ctx *abstraction.Context, payload *dto.UserUpdateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = payload.ID
		if payload.Name != nil {
			newUserData.Name = *payload.Name
		}
		if payload.Email != nil {
			userEmail, err := s.UserRepository.FindByEmail(ctx, *payload.Email)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if userEmail != nil && userEmail.Email != *payload.Email {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "email already exist")
			}
			newUserData.Email = *payload.Email
		}
		if payload.RoleId != nil {
			newUserData.RoleId = *payload.RoleId
		}
		if payload.IsLocked != nil {
			newUserData.IsLocked = *payload.IsLocked
			if err = s.UserRepository.UpdateLocked(ctx, newUserData).Error; err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
		}

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		if payload.IsLocked != nil {
			if *payload.IsLocked {
				if err = gomail.SendMail(userData.Email, "Account Locked for SelarasHomeId", general.ParseTemplateEmailToHtml("./assets/html/email/notif_locked_user.html", struct {
					NAME  string
					EMAIL string
				}{
					NAME:  userData.Name,
					EMAIL: userData.Email,
				})); err != nil {
					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
				}

				userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
				for _, v := range userLoginFrom {
					general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
				}
			} else {
				if err = gomail.SendMail(userData.Email, "Account Unlocked for SelarasHomeId", general.ParseTemplateEmailToHtml("./assets/html/email/notif_unlocked_user.html", struct {
					NAME  string
					EMAIL string
				}{
					NAME:  userData.Name,
					EMAIL: userData.Email,
				})); err != nil {
					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
				}

				userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
				for _, v := range userLoginFrom {
					general.RemoveUUIDFromRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success update!",
	}, nil
}

func (s *service) Delete(ctx *abstraction.Context, payload *dto.UserDeleteByIDRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_ADMIN {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = userData.ID
		newUserData.IsDelete = true

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}

func (s *service) ChangePassword(ctx *abstraction.Context, payload *dto.UserChangePasswordRequest) (map[string]interface{}, error) {
	var sendNotifTo []int = nil
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.ID != payload.ID {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this user is not permitted")
		}

		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		if err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(payload.OldPassword)); err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, err, "old password is wrong")
		}

		if payload.OldPassword == payload.NewPassword {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "the new password cannot be the same as the old password")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = userData.ID
		newUserData.Password = string(hashedPassword)

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
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
		"message": "success change password!",
	}, nil
}

func (s *service) ResetPassword(ctx *abstraction.Context, payload *dto.UserResetPasswordRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_ADMIN {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		userLogin, err := s.UserRepository.FindById(ctx, ctx.Auth.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		passwordString := general.GeneratePassword(8, 1, 1, 1, 1)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordString), bcrypt.DefaultCost)
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = userData.ID
		newUserData.Password = string(hashedPassword)

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		if err = gomail.SendMail(userData.Email, "Reset Password for SelarasHomeId", general.ParseTemplateEmailToHtml("./assets/html/email/notif_reset_password.html", struct {
			NAME      string
			RESETNAME string
			EMAIL     string
			PASSWORD  string
			LINK      string
		}{
			NAME:      userData.Name,
			RESETNAME: userLogin.Name,
			EMAIL:     userData.Email,
			PASSWORD:  passwordString,
			LINK:      constant.BASE_URL,
		})); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success reset password!",
	}, nil
}

func (s *service) GetUserInfo(ctx *abstraction.Context) (map[string]interface{}, error) {
	var res map[string]interface{} = nil
	data, err := s.UserRepository.FindById(ctx, ctx.Auth.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if data != nil {
		res = map[string]interface{}{
			"id":         data.ID,
			"name":       data.Name,
			"email":      data.Email,
			"is_delete":  data.IsDelete,
			"is_locked":  data.IsLocked,
			"login_from": data.LoginFrom,
			"created_at": general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
			"role": map[string]interface{}{
				"id":   data.Role.ID,
				"name": data.Role.Name,
			},
		}

	}
	return map[string]interface{}{
		"data": res,
	}, nil
}

func (s *service) Export(ctx *abstraction.Context) (string, *bytes.Buffer, error) {
	data, err := s.UserRepository.Find(ctx, true)
	if err != nil && err.Error() != "record not found" {
		return "", nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	f := excelize.NewFile()
	sheet := "Master Data - User"
	index, err := f.NewSheet(general.TruncateSheetName(sheet))
	if err != nil {
		return "", nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(index)
	f.SetCellValue(sheet, "A1", "No")
	f.SetCellValue(sheet, "B1", "Nama")
	f.SetCellValue(sheet, "C1", "Email")
	f.SetCellValue(sheet, "D1", "Role")
	f.SetCellValue(sheet, "E1", "Status Akun")
	f.SetCellValue(sheet, "F1", "Login Terbaru")
	f.SetCellValue(sheet, "G1", "Tanggal Dibuat")
	for i, v := range data {
		colA := fmt.Sprintf("A%d", i+2)
		colB := fmt.Sprintf("B%d", i+2)
		colC := fmt.Sprintf("C%d", i+2)
		colD := fmt.Sprintf("D%d", i+2)
		colE := fmt.Sprintf("E%d", i+2)
		colF := fmt.Sprintf("F%d", i+2)
		colG := fmt.Sprintf("G%d", i+2)
		no := i + 1
		f.SetCellValue(sheet, colA, no)
		f.SetCellValue(sheet, colB, v.Name)
		f.SetCellValue(sheet, colC, v.Email)
		f.SetCellValue(sheet, colD, v.Role.Name)
		if v.IsLocked {
			f.SetCellValue(sheet, colE, "Terkunci")
		} else {
			f.SetCellValue(sheet, colE, "Tidak Terkunci")
		}
		if v.LoginFrom != "" {
			f.SetCellValue(sheet, colF, general.CapitalizeEachWord(v.LoginFrom))
		} else {
			f.SetCellValue(sheet, colF, "-")
		}
		f.SetCellValue(sheet, colG, v.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	filename := fmt.Sprintf("Master Data - User (%s).xlsx", general.NowLocal().Format("2006-01-02"))
	return filename, &buf, nil
}
