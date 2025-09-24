package dto

type UserCreateRequest struct {
	Name   string `json:"name" form:"name" validate:"required"`
	Email  string `json:"email" form:"email" validate:"required"`
	RoleId int    `json:"role_id" form:"role_id" validate:"required"`
}

type UserFindByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type UserUpdateRequest struct {
	ID       int     `param:"id" validate:"required"`
	Name     *string `json:"name" form:"name"`
	Email    *string `json:"email" form:"email"`
	RoleId   *int    `json:"role_id" form:"role_id"`
	IsLocked *bool   `json:"is_locked" form:"is_locked"`
}

type UserDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type UserChangePasswordRequest struct {
	ID          int    `param:"id" validate:"required"`
	OldPassword string `json:"old_password" form:"old_password" validate:"required"`
	NewPassword string `json:"new_password" form:"new_password" validate:"required"`
}

type UserResetPasswordRequest struct {
	ID int `param:"id" validate:"required"`
}
