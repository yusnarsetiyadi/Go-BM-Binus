package dto

type EventTypeCreateRequest struct {
	Name     string `json:"name" form:"name" validate:"required"`
	Priority int    `json:"priority" form:"priority" validate:"required"`
}

type EventTypeDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type EventTypeUpdateRequest struct {
	ID       int     `param:"id" validate:"required"`
	Name     *string `json:"name" form:"name"`
	Priority *int    `json:"priority" form:"priority"`
}
