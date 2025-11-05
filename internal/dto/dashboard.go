package dto

type GetDashboardRequest struct {
	RoleId int `param:"role_id" validate:"required"`
}
