package dto

type NotificationSetReadRequest struct {
	ID int `param:"id" validate:"required"`
}
