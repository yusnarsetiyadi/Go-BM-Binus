package dto

import (
	"bm_binus/internal/config"
	modeltoken "bm_binus/internal/model/token"
	"fmt"

	"github.com/golang-jwt/jwt"
)

type AuthLoginRequest struct {
	Email    string `json:"email" form:"email" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

func (r RefreshTokenRequest) TokenClaims() (*modeltoken.TokenClaims, error) {
	token, err := jwt.Parse(r.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method :%v", token.Header["alg"])
		}
		return []byte(config.Get().JWT.SecretKey), nil
	})
	if token == nil || !token.Valid || err != nil {
		if jwtErrValidation, ok := err.(*jwt.ValidationError); ok {
			c := token.Claims.(jwt.MapClaims)
			return &modeltoken.TokenClaims{
				ID:        c["id"].(string),
				RoleID:    c["role_id"].(string),
				Email:     c["email"].(string),
				UuidLogin: c["uuid_login"].(string),
				Exp:       int64(c["exp"].(float64)),
			}, jwtErrValidation
		}
		return nil, jwt.NewValidationError("invalid_token", jwt.ValidationErrorMalformed)
	}
	c := token.Claims.(jwt.MapClaims)
	return &modeltoken.TokenClaims{
		ID:        c["id"].(string),
		RoleID:    c["role_id"].(string),
		Email:     c["email"].(string),
		UuidLogin: c["uuid_login"].(string),
		Exp:       int64(c["exp"].(float64)),
	}, nil
}

type AuthSendEmailForgotPasswordRequest struct {
	Email string `json:"email" form:"email" validate:"required"`
}

type AuthValidationResetPasswordRequest struct {
	Token string `param:"token" validate:"required"`
}
