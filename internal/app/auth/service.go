package auth

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/config"
	"bm_binus/internal/dto"
	"bm_binus/internal/factory"
	"bm_binus/internal/model"
	modelToken "bm_binus/internal/model/token"
	"bm_binus/internal/repository"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/gomail"
	"bm_binus/pkg/util/aescrypt"
	"bm_binus/pkg/util/encoding"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"bm_binus/pkg/util/trxmanager"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	Login(ctx *abstraction.Context, payload *dto.AuthLoginRequest) (map[string]interface{}, error)
	Logout(ctx *abstraction.Context) (map[string]interface{}, error)
	RefreshToken(ctx *abstraction.Context) (map[string]interface{}, error)
	SendEmailForgotPassword(ctx *abstraction.Context, payload *dto.AuthSendEmailForgotPasswordRequest) (map[string]interface{}, error)
	ValidationResetPassword(ctx *abstraction.Context, payload *dto.AuthValidationResetPasswordRequest) (string, error)
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

func (s *service) encryptTokenClaims(v int) (encryptedString string, err error) {
	encryptedString, err = aescrypt.EncryptAES(fmt.Sprint(v), config.Get().JWT.SecretKey)
	return
}

func (s *service) Login(ctx *abstraction.Context, payload *dto.AuthLoginRequest) (map[string]interface{}, error) {
	var (
		err   error
		data  = new(model.UserEntityModel)
		token string
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		data, err = s.UserRepository.FindByEmail(ctx, payload.Email)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if data == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "email or password is incorrect")
		}

		if err = bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(payload.Password)); err != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "email or password is incorrect")
		}

		var encryptedUserID string
		if encryptedUserID, err = s.encryptTokenClaims(data.ID); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		var encryptedUserRoleID string
		if encryptedUserRoleID, err = s.encryptTokenClaims(data.RoleId); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		encodedEmail := encoding.Encode(data.Email)
		uuidUserLogin := uuid.NewString()
		encodedUuidLogin := encoding.Encode(uuidUserLogin)

		tokenClaims := &modelToken.TokenClaims{
			ID:        encryptedUserID,
			RoleID:    encryptedUserRoleID,
			Email:     encodedEmail,
			UuidLogin: encodedUuidLogin,
			Exp:       time.Now().Add(time.Duration(24 * time.Hour)).Unix(),
		}
		authToken := modelToken.NewAuthToken(tokenClaims)
		token, err = authToken.Token()
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		general.AppendUUIDToRedisArray(s.DbRedis, general.GenerateRedisKeyUserLogin(data.ID), uuidUserLogin)

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"token": token,
		"data": map[string]interface{}{
			"id":         data.ID,
			"name":       data.Name,
			"email":      data.Email,
			"created_at": general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
			"role": map[string]interface{}{
				"id":   data.Role.ID,
				"name": data.Role.Name,
			},
		},
	}, nil
}

func (s *service) Logout(ctx *abstraction.Context) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {

		general.RemoveUUIDFromRedisArray(s.DbRedis, general.GenerateRedisKeyUserLogin(ctx.Auth.ID), ctx.Auth.UuidLogin)
		general.RemoveUUIDFromRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, ctx.Auth.UuidLogin)

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "success logout!",
	}, nil
}

func (s *service) RefreshToken(ctx *abstraction.Context) (map[string]interface{}, error) {
	var token string
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		data, err := s.UserRepository.FindById(ctx, ctx.Auth.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		var encryptedUserID string
		if encryptedUserID, err = s.encryptTokenClaims(data.ID); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		var encryptedUserRoleID string
		if encryptedUserRoleID, err = s.encryptTokenClaims(data.RoleId); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		encodedEmail := encoding.Encode(data.Email)
		encodedUuidLogin := encoding.Encode(ctx.Auth.UuidLogin)

		tokenClaims := &modelToken.TokenClaims{
			ID:        encryptedUserID,
			RoleID:    encryptedUserRoleID,
			Email:     encodedEmail,
			UuidLogin: encodedUuidLogin,
			Exp:       time.Now().Add(time.Duration(24 * time.Hour)).Unix(),
		}
		authToken := modelToken.NewAuthToken(tokenClaims)
		token, err = authToken.Token()
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"token": token,
	}, nil
}

func (s *service) SendEmailForgotPassword(ctx *abstraction.Context, payload *dto.AuthSendEmailForgotPasswordRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		data, err := s.UserRepository.FindByEmail(ctx, payload.Email)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if data == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "email not found")
		}

		eksternalToken := new(modelToken.AuthEksternalToken)
		eksternalToken.UserId = data.ID
		token, err := eksternalToken.GenerateTokenEksternal()
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		s.DbRedis.Set(context.Background(), *token, *token, 0)

		if err = gomail.SendMail(data.Email, "Forgot Password for Building Management Binus", general.ParseTemplateEmailToHtml("./assets/html/email/notif_forgot_password.html", struct {
			NAME  string
			EMAIL string
			LINK  string
		}{
			NAME:  data.Name,
			EMAIL: data.Email,
			LINK:  constant.BASE_URL + "/auth/validation/reset-password/" + *token,
		})); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "success send email forgot password!",
	}, nil
}

func (s *service) ValidationResetPassword(ctx *abstraction.Context, payload *dto.AuthValidationResetPasswordRequest) (string, error) {
	userData := new(model.UserEntityModel)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		_, err := s.DbRedis.Get(context.Background(), payload.Token).Result()
		if err == redis.Nil {
			return errors.New("your token is invalid")
		} else {
			s.DbRedis.Del(context.Background(), payload.Token)
		}

		data, err := modelToken.ValidateTokenEksternal(payload.Token)
		if err != nil {
			return errors.New("your token is invalid")
		}

		userData, err = s.UserRepository.FindById(ctx, data.UserId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		passwordString := general.GeneratePassword(8, 1, 1, 1, 1)
		password := []byte(passwordString)
		hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
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

		if err = gomail.SendMail(userData.Email, "Reset Password for Building Management Binus", general.ParseTemplateEmailToHtml("./assets/html/email/notif_reset_password.html", struct {
			NAME      string
			RESETNAME string
			EMAIL     string
			PASSWORD  string
			LINK      string
		}{
			NAME:      userData.Name,
			RESETNAME: "System",
			EMAIL:     userData.Email,
			PASSWORD:  passwordString,
			LINK:      constant.BASE_URL_UI,
		})); err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
		}

		return nil
	}); err != nil {
		return "", err
	}

	return userData.Email, nil
}
