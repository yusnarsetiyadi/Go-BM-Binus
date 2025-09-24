package token

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/config"
	"bm_binus/pkg/util/aescrypt"
	"bm_binus/pkg/util/encoding"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

type TokenClaims struct {
	ID        string `json:"id"`
	RoleID    string `json:"role_id"`
	Email     string `json:"email"`
	UuidLogin string `json:"uuid_login"`
	Exp       int64  `json:"exp"`

	jwt.RegisteredClaims
}

func (c TokenClaims) AuthContext() (*abstraction.AuthContext, error) {
	var (
		id         int
		role_id    int
		email      string
		uuid_login string
		err        error

		encryptionKey = config.Get().JWT.SecretKey
	)

	destructID := c.ID
	if destructID == "" {
		return nil, errors.New("invalid_token")
	}
	if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
		if destructID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructID), encryptionKey); err != nil {
			return nil, errors.New("invalid_token")
		}
		if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
			return nil, errors.New("invalid_token")
		}
	}

	destructRoleID := c.RoleID
	if destructRoleID == "" {
		return nil, errors.New("invalid_token")
	}
	if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
		if destructRoleID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructRoleID), encryptionKey); err != nil {
			return nil, errors.New("invalid_token")
		}
		if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
			return nil, errors.New("invalid_token")
		}
	}

	destructEmail := c.Email
	if destructEmail == "" {
		return nil, errors.New("invalid_token")
	}
	if email, err = encoding.Decode(fmt.Sprintf("%v", destructEmail)); err != nil {
		return nil, errors.New("invalid_token")
	}

	destructUuidLogin := c.UuidLogin
	if destructUuidLogin == "" {
		return nil, errors.New("invalid_token")
	}
	if uuid_login, err = encoding.Decode(fmt.Sprintf("%v", destructUuidLogin)); err != nil {
		return nil, errors.New("invalid_token")
	}

	return &abstraction.AuthContext{
		ID:        id,
		RoleID:    role_id,
		Email:     email,
		UuidLogin: uuid_login,
	}, nil
}
