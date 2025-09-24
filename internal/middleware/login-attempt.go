package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"bm_binus/internal/abstraction"
	"bm_binus/internal/model"
	"bm_binus/pkg/database"
	"bm_binus/pkg/gomail"
	"bm_binus/pkg/util/general"
	"bm_binus/pkg/util/response"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type LoginAttemptStore interface {
	// This method checks if a user with the given identifier is allowed to attempt login.
	// It returns a boolean indicating if the user is allowed,
	// a float64 representing the number of seconds to wait before retrying if login is not allowed, and
	// an error if there are too many login attempts.
	Allow(identifier string, email string) (bool, float64, error)

	// This method increments the login attempt count for a given identifier and updates the user's last seen time.
	// It takes an echo.Context object representing the HTTP request context and returns an error object if there was an error during the process.
	IncreaseAttempt(c echo.Context, identifier string, email string) error
}

/*
The LoginAttemptConfig struct defines the configuration for a middleware that limits the number of login attempts for a user.

Skipper: A function that determines if the middleware should skip a request or not.

IdentifierExtractor: A function that extracts the identifier for a request, typically the user's IP address.

Store: An interface that defines the methods required to store and retrieve login attempt information.

ErrorHandler: A function that handles errors that occur during the login attempt process.

DenyHandler: A function that handles requests that exceed the maximum number of login attempts.

LockedHandler: A function that handles requests that are locked out due to too many login attempts that make account is locked.

In summary, the LoginAttemptConfig struct provides the necessary configuration for a middleware that limits the number of login attempts for a user.
*/
type LoginAttemptConfig struct {
	// Skipper defines a function to skip middleware. Returning true skips processing the middleware.
	Skipper echoMiddleware.Skipper

	// Extractor is used to extract data from echo.Context
	IdentifierExtractor echoMiddleware.Extractor

	// Store is used to store and retrieve login attempt information
	Store LoginAttemptStore

	// ErrorHandler is used to handle errors that occur during the login attempt process
	ErrorHandler func(c echo.Context, err error) error

	// DenyHandler is used to handle requests that exceed the maximum number of login attempts
	DenyHandler func(c echo.Context, identifier string, err error) error

	// LockedHandler is used to handle requests that are locked out due to too many login attempts that make account is locked
	LockedHandler func(c echo.Context, identifier string, err error) error
}

var DefaultLoginAttemptConfig = LoginAttemptConfig{
	Skipper: func(c echo.Context) bool {
		return !strings.Contains(c.Request().RequestURI, "login")
	},
	IdentifierExtractor: func(c echo.Context) (string, error) {
		return c.RealIP(), nil
	},
	ErrorHandler: func(c echo.Context, err error) error {
		return echoMiddleware.ErrExtractorError
	},
	DenyHandler: func(c echo.Context, identifier string, err error) error {
		return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), err.Error())
	},
	LockedHandler: func(c echo.Context, identifier string, err error) error {
		return response.ErrorBuilder(http.StatusUnauthorized, errors.New("unauthorized"), err.Error())
	},
}

// LoginAttempt returns an echo.MiddlewareFunc that applies rate limiting to incoming requests.
//
// It takes a LoginAttemptStore as a parameter and sets default values for any missing fields. It then creates a new LoginAttemptConfig
// with the provided store and calls LoginAttemptWithConfig with the config.
//
// Parameters:
// - store: The LoginAttemptStore that will be used to store and retrieve login attempt information.
//
// Returns:
// - echo.MiddlewareFunc: The middleware function that applies rate limiting to incoming requests.
func LoginAttempt(store LoginAttemptStore) echo.MiddlewareFunc {
	c := DefaultLoginAttemptConfig
	c.Store = store
	return LoginAttemptWithConfig(c)
}

// LoginAttemptWithConfig returns an echo.MiddlewareFunc that applies rate limiting to incoming requests.
//
// It takes a LoginAttemptConfig as a parameter and sets default values for any missing fields. It then checks if the
// provided Store configuration is nil and panics if it is. The returned middleware function wraps the next handler
// and applies the rate limiting logic.
//
// The rate limiting logic checks if the request should be skipped based on the provided Skipper function. If the request
// should be skipped, the next handler is called. Otherwise, the IdentifierExtractor function is called to extract the
// identifier for the request. The Store's Allow method is then called to check if the request is allowed based on the
// identifier. If the request is not allowed or an error occurs, the DenyHandler function is called with the
// identifier and error. If the request is allowed, the next handler is called. After the next handler is called, the
// Store's IncreaseAttempt method is called to increase the attempt count for the identifier. Finally, any error returned
// by the next handler or the Store's IncreaseAttempt method is returned.
//
// Parameters:
// - config: The LoginAttemptConfig containing the configuration for the rate limiting middleware.
//
// Returns:
// - echo.MiddlewareFunc: The middleware function that applies rate limiting to incoming requests.
func LoginAttemptWithConfig(config LoginAttemptConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultLoginAttemptConfig.Skipper
	}
	if config.IdentifierExtractor == nil {
		config.IdentifierExtractor = DefaultLoginAttemptConfig.IdentifierExtractor
	}
	if config.DenyHandler == nil {
		config.DenyHandler = DefaultLoginAttemptConfig.DenyHandler
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultLoginAttemptConfig.ErrorHandler
	}
	if config.LockedHandler == nil {
		config.LockedHandler = DefaultLoginAttemptConfig.LockedHandler
	}
	if config.Store == nil {
		panic("Store configuration must be provided")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}
			var (
				identifier        string
				allow             bool
				retryAfterSeconds float64
				reqBody           []byte
				req               = c.Request()
				parsedReq         = make(map[string]interface{})
			)
			if identifier, err = config.IdentifierExtractor(c); err != nil {
				response.ErrorResponse(config.ErrorHandler(c, err)).SendError(c)
				return response.ErrorResponse(config.ErrorHandler(c, err)).SendError(c)
			}

			// Request
			if req.Body != nil { // Read
				reqBody, _ = io.ReadAll(req.Body)
			}
			req.Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset
			c.SetRequest(req)

			_ = json.Unmarshal(reqBody, &parsedReq)

			// check if request is allowed
			if allow, retryAfterSeconds, err = config.Store.Allow(identifier, fmt.Sprint(parsedReq["email"])); err != nil || !allow {
				if retryAfterSeconds != 0 {
					return response.ErrorResponse(config.DenyHandler(c, identifier, err)).SendError(c)
				}
				return response.ErrorResponse(config.LockedHandler(c, identifier, err)).SendError(c)
			}

			// get actual HTTP response from given endpoint and send it back to c (echo.Context)
			if err = next(c); err != nil {
				return response.ErrorResponse(err).SendError(c)
			}

			// increase attempt count
			if err = config.Store.IncreaseAttempt(c, identifier, fmt.Sprint(parsedReq["email"])); err != nil {
				return response.ErrorResponse(config.DenyHandler(c, identifier, err)).SendError(c)
			}
			return
		}
	}
}

type (
	// LoginAttemptMemoryStore is a simple implementation of the LoginAttemptStore interface.
	// It stores login attempts in memory and cleans up stale users after a certain period of time.
	LoginAttemptMemoryStore struct {
		users map[string]*User // map of user identifiers to their login attempts

		mutex sync.RWMutex // mutex for synchronizing access to the users map

		maxAttempts int // maximum number of login attempts allowed

		cleanedUpIn time.Duration // duration for which the users map is cleaned up
		lastCleanUp time.Time     // last time the users map was cleaned up

		isError func(c echo.Context) bool // function to check if a given context indicates an error during login
		timeNow func() time.Time          // function to get the current time
	}

	// User represents a single user in the LoginAttemptMemoryStore.
	// It contains the number of login attempts, the last time the user was seen, and the time when the user was locked out.
	User struct {
		Attempts     int           // number of login attempts
		LastSeen     time.Time     // last time the user was seen
		LockedAt     time.Time     // time when the user was locked out
		LockDuration time.Duration // duration for which the user is locked out
		Locked       bool          // whether the user is locked out
	}
)

// LoginAttemptMemoryStoreConfig represents the configuration for the LoginAttemptMemoryStore.
// It contains the maximum number of login attempts allowed, the duration for which a user is locked out after too many failed login attempts,
// and a function to check if a given context indicates an error during login.

type LoginAttemptMemoryStoreConfig struct {
	MaxAttempts int                       // Defines the maximum number of login attempts allowed.
	CleanedUpIn time.Duration             // Defines the duration for which the users map is cleaned up.
	IsError     func(c echo.Context) bool // Checks if a given context indicates an error during login.
}

var DefaultLoginAttemptMemoryStoreConfig = LoginAttemptMemoryStoreConfig{
	CleanedUpIn: 1 * time.Minute,
	IsError: func(c echo.Context) bool {
		return c.Response().Status == http.StatusUnauthorized
	},
}

// NewLoginAttemptMemoryStore creates a new instance of LoginAttemptMemoryStore with the given maxAttempts.
//
// Parameters:
// - maxAttempts: an integer representing the maximum login attempts allowed.
// Returns:
// - a pointer to a LoginAttemptMemoryStore object.
func NewLoginAttemptMemoryStore(maxAttempts int) *LoginAttemptMemoryStore {
	return NewLoginAttemptMemoryStoreWithConfig(LoginAttemptMemoryStoreConfig{
		MaxAttempts: maxAttempts,
	})
}

// NewLoginAttemptMemoryStoreWithConfig creates a new instance of LoginAttemptMemoryStore
// with the given configuration.
//
// Parameters:
// - config: a LoginAttemptMemoryStoreConfig object containing the configuration
//
// Returns:
// - store: a pointer to a LoginAttemptMemoryStore object
func NewLoginAttemptMemoryStoreWithConfig(config LoginAttemptMemoryStoreConfig) (store *LoginAttemptMemoryStore) {
	store = new(LoginAttemptMemoryStore)
	store.maxAttempts = config.MaxAttempts
	store.cleanedUpIn = config.CleanedUpIn
	if config.CleanedUpIn == 0 {
		store.cleanedUpIn = DefaultLoginAttemptMemoryStoreConfig.CleanedUpIn
	}
	if config.IsError == nil {
		store.isError = DefaultLoginAttemptMemoryStoreConfig.IsError
	}
	store.users = make(map[string]*User)
	store.timeNow = time.Now
	store.lastCleanUp = store.timeNow()
	return
}

// Allow checks if a user with the given identifier is allowed to attempt login.
//
// Parameters:
// - identifier: a string representing the identifier of the user.
// - email: a string representing the email of the user.
// Returns:
// - bool: true if the user is allowed to login, false otherwise.
// - float64: the number of seconds to wait before retrying if login is not allowed.
// - error: an error if there are too many login attempts, nil otherwise.
func (store *LoginAttemptMemoryStore) Allow(identifier string, email string) (bool, float64, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	key := fmt.Sprintf("%v:%v", identifier, email)
	user, exists := store.users[key]
	if !exists {
		user = new(User)
		store.users[key] = user
	}

	now := store.timeNow()
	user.LastSeen = now
	if now.Sub(store.lastCleanUp) > store.cleanedUpIn {
		store.cleanUpStaleUsers()
	}

	if user.Locked {
		return false, 0, fmt.Errorf("account is locked. please contact admin to unlock your account")
	}

	retryAfterSeconds := user.LockedAt.Sub(now).Truncate(time.Second).Seconds()
	if now.Before(user.LockedAt) {
		return false, retryAfterSeconds, fmt.Errorf("too many login attempts, retry after %v seconds", retryAfterSeconds)
	}

	return true, 0, nil
}

// IncreaseAttempt increments the login attempt count for a given identifier and updates the user's last seen time.
//
// Parameters:
// - c: an echo.Context object representing the HTTP request context.
// - identifier: a string representing the identifier of the user.
// - email: a string representing the email of the user.
//
// Returns:
// - error: an error object if there was an error during the process, otherwise nil.
func (store *LoginAttemptMemoryStore) IncreaseAttempt(c echo.Context, identifier string, email string) (err error) {
	conn, _ := database.Connection("MYSQL")
	store.mutex.Lock()
	defer store.mutex.Unlock()

	key := fmt.Sprintf("%v:%v", identifier, email)
	user, exists := store.users[key]
	if !exists {
		user = new(User)
		store.users[key] = user
	}

	var userEntityModel *model.UserEntityModel
	if err = conn.Model(&model.UserEntityModel{}).Where("email = ?", email).Find(&userEntityModel).Error; err != nil {
		return
	}
	userEntityModel.Context = &abstraction.Context{
		Auth: &abstraction.AuthContext{
			ID: userEntityModel.ID,
		},
	}

	now := store.timeNow()
	user.LastSeen = now
	if now.Sub(store.lastCleanUp) > store.cleanedUpIn {
		store.cleanUpStaleUsers()
	}

	if store.isError(c) {
		user.Attempts++
		if user.Attempts >= store.maxAttempts {
			switch user.LockDuration {
			case 1 * time.Minute:
				user.LockDuration = 15 * time.Minute
			case 15 * time.Minute:
				err = conn.Model(userEntityModel).Where("email = ?", email).Update("is_locked", true).Error
				err = gomail.SendMail(email, "Account Locked for SelarasHomeId", general.ParseTemplateEmailToHtml("./assets/html/email/notif_locked_user.html", struct {
					NAME  string
					EMAIL string
				}{
					NAME:  userEntityModel.Name,
					EMAIL: userEntityModel.Email,
				}))
				user.Locked = true
			default:
				user.LockDuration = 1 * time.Minute
			}
			user.LockedAt = now.Add(user.LockDuration)
			user.Attempts = 0
		}
	} else {
		user.Attempts = 0
		user.LockDuration = 1 * time.Minute
	}
	return
}

// cleanUpStaleUsers removes users from the store that have not been active for a certain period of time.
//
// No parameters.
// No return values.
func (store *LoginAttemptMemoryStore) cleanUpStaleUsers() {
	for identifier, user := range store.users {
		if store.timeNow().Sub(user.LastSeen) > store.cleanedUpIn {
			delete(store.users, identifier)
		}
	}
	store.lastCleanUp = store.timeNow()
}
