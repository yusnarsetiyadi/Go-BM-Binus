package middleware

import (
	"bm_binus/internal/abstraction"
	"bm_binus/internal/config"
	"bm_binus/pkg/constant"
	"bm_binus/pkg/database"
	"bm_binus/pkg/util/aescrypt"
	"bm_binus/pkg/util/encoding"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"slices"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

func Authentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			id         int
			role_id    int
			email      string
			uuid_login string
			jwtKey     = config.Get().JWT.SecretKey
		)
		authToken := c.Request().Header.Get("Authorization")
		if authToken == "" {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if !strings.Contains(authToken, "Bearer") {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		tokenString := strings.Replace(authToken, "Bearer ", "", -1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method :%v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		})
		if token == nil || !token.Valid || err != nil {
			if errJWT, ok := err.(*jwt.ValidationError); ok {
				if errJWT.Errors == jwt.ValidationErrorExpired {
					return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), errJWT.Error()).SendError(c)
				} else {
					return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
				}
			} else {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return response.ErrorBuilder(http.StatusUnauthorized, err, "error when claim token").SendError(c)
		}

		destructID := claims["id"]
		if destructID == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
			if destructID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructID), jwtKey); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
			if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
		}

		destructRoleID := claims["role_id"]
		if destructRoleID == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
			if destructRoleID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructRoleID), jwtKey); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
			if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
		}

		destructEmail := claims["email"]
		if destructEmail == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if email, err = encoding.Decode(fmt.Sprintf("%v", destructEmail)); err != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}

		destructUuidLogin := claims["uuid_login"]
		if destructUuidLogin == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if uuid_login, err = encoding.Decode(fmt.Sprintf("%v", destructUuidLogin)); err != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}

		dbRedis := database.InitRedis()
		userMustLogout := general.GetRedisUUIDArray(dbRedis, constant.REDIS_KEY_AUTO_LOGOUT)
		if slices.Contains(userMustLogout, uuid_login) {
			return response.ErrorBuilder(http.StatusUnprocessableEntity, errors.New("unprocessable"), "expired_token").SendError(c)
		}

		cc := c.(*abstraction.Context)
		cc.Auth = &abstraction.AuthContext{
			ID:        id,
			RoleID:    role_id,
			Email:     email,
			UuidLogin: uuid_login,
		}

		return next(cc)
	}
}

func Logout(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			id         int
			role_id    int
			email      string
			uuid_login string
			jwtKey     = config.Get().JWT.SecretKey
		)
		authToken := c.Request().Header.Get("Authorization")
		if authToken == "" {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if !strings.Contains(authToken, "Bearer") {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		tokenString := strings.Replace(authToken, "Bearer ", "", -1)
		token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method :%v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		}, jwt.WithoutClaimsValidation())

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return response.ErrorBuilder(http.StatusUnauthorized, err, "error when claim token").SendError(c)
		}

		destructID := claims["id"]
		if destructID == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
			if destructID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructID), jwtKey); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
			if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
		}

		destructRoleID := claims["role_id"]
		if destructRoleID == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
			if destructRoleID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructRoleID), jwtKey); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
			if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
		}

		destructEmail := claims["email"]
		if destructEmail == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if email, err = encoding.Decode(fmt.Sprintf("%v", destructEmail)); err != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}

		destructUuidLogin := claims["uuid_login"]
		if destructUuidLogin == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if uuid_login, err = encoding.Decode(fmt.Sprintf("%v", destructUuidLogin)); err != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}

		cc := c.(*abstraction.Context)
		cc.Auth = &abstraction.AuthContext{
			ID:        id,
			RoleID:    role_id,
			Email:     email,
			UuidLogin: uuid_login,
		}

		return next(cc)
	}
}

func RefreshToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			id         int
			role_id    int
			email      string
			uuid_login string
			jwtKey     = config.Get().JWT.SecretKey
		)
		authToken := c.Request().Header.Get("Authorization")
		if authToken == "" {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if !strings.Contains(authToken, "Bearer") {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		tokenString := strings.Replace(authToken, "Bearer ", "", -1)
		token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method :%v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		}, jwt.WithoutClaimsValidation())

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return response.ErrorBuilder(http.StatusUnauthorized, err, "error when claim token").SendError(c)
		}

		destructID := claims["id"]
		if destructID == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
			if destructID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructID), jwtKey); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
			if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
		}

		destructRoleID := claims["role_id"]
		if destructRoleID == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
			if destructRoleID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructRoleID), jwtKey); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
			if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
				return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
			}
		}

		destructEmail := claims["email"]
		if destructEmail == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if email, err = encoding.Decode(fmt.Sprintf("%v", destructEmail)); err != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}

		destructUuidLogin := claims["uuid_login"]
		if destructUuidLogin == nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if uuid_login, err = encoding.Decode(fmt.Sprintf("%v", destructUuidLogin)); err != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}

		dbRedis := database.InitRedis()
		userMustLogout := general.GetRedisUUIDArray(dbRedis, constant.REDIS_KEY_AUTO_LOGOUT)
		if slices.Contains(userMustLogout, uuid_login) {
			return response.ErrorBuilder(http.StatusUnprocessableEntity, errors.New("unprocessable"), "expired_token").SendError(c)
		}

		keysRefreshToken := fmt.Sprintf(constant.REDIS_KEY_REFRESH_TOKEN, uuid_login)
		value := dbRedis.Incr(context.Background(), keysRefreshToken)
		if value.Err() != nil {
			return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token").SendError(c)
		}
		if value.Val() > constant.REDIS_MAX_REFRESH_TOKEN {
			dbRedis.Del(context.Background(), keysRefreshToken)
			return response.ErrorBuilder(http.StatusUnprocessableEntity, errors.New("unprocessable"), "expired_token").SendError(c)
		}

		cc := c.(*abstraction.Context)
		cc.Auth = &abstraction.AuthContext{
			ID:        id,
			RoleID:    role_id,
			Email:     email,
			UuidLogin: uuid_login,
		}

		return next(cc)
	}
}

func JustValidateToken(tokenString string) (*abstraction.Context, *response.MetaError) {
	var (
		id         int
		role_id    int
		email      string
		uuid_login string
		jwtKey     = config.Get().JWT.SecretKey
	)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, response.ErrorBuilder(http.StatusUnauthorized, fmt.Errorf("unexpected signing method :%v", token.Header["alg"]), "error unexpected signing method")
		}
		return []byte(jwtKey), nil
	})

	if token == nil || !token.Valid || err != nil {
		if errJWT, ok := err.(*jwt.ValidationError); ok {
			if errJWT.Errors == jwt.ValidationErrorExpired {
				return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), errJWT.Error())
			} else {
				return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
			}
		} else {
			return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, response.ErrorBuilder(http.StatusUnauthorized, err, "error when claim token")
	}

	destructID := claims["id"]
	if destructID == nil {
		return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
	}
	if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
		if destructID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructID), jwtKey); err != nil {
			return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
		}
		if id, err = strconv.Atoi(fmt.Sprintf("%v", destructID)); err != nil {
			return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
		}
	}

	destructRoleID := claims["role_id"]
	if destructRoleID == nil {
		return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
	}
	if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
		if destructRoleID, err = aescrypt.DecryptAES(fmt.Sprintf("%v", destructRoleID), jwtKey); err != nil {
			return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
		}
		if role_id, err = strconv.Atoi(fmt.Sprintf("%v", destructRoleID)); err != nil {
			return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
		}
	}

	destructEmail := claims["email"]
	if destructEmail == nil {
		return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
	}
	if email, err = encoding.Decode(fmt.Sprintf("%v", destructEmail)); err != nil {
		return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
	}

	destructUuidLogin := claims["uuid_login"]
	if destructUuidLogin == nil {
		return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
	}
	if uuid_login, err = encoding.Decode(fmt.Sprintf("%v", destructUuidLogin)); err != nil {
		return nil, response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), "invalid_token")
	}

	cc := new(abstraction.Context)
	cc.Auth = &abstraction.AuthContext{
		ID:        id,
		RoleID:    role_id,
		Email:     email,
		UuidLogin: uuid_login,
	}

	return cc, nil
}
