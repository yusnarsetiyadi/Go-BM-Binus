package dto

import "mime/multipart"

type RequestCreateRequest struct {
	EventName        string `json:"event_name" form:"event_name" validate:"required"`
	EventLocation    string `json:"event_location" form:"event_location" validate:"required"`
	EventDateStart   string `json:"event_date_start" form:"event_date_start" validate:"required"`
	EventDateEnd     string `json:"event_date_end" form:"event_date_end" validate:"required"`
	Description      string `json:"description" form:"description" validate:"required"`
	EventTypeId      int    `json:"event_type_id" form:"event_type_id" validate:"required"`
	CountParticipant int    `json:"count_participant" form:"count_participant" validate:"required"`
	Files            []*multipart.FileHeader
}

type RequestFindRequest struct {
	UseAhp          *string `query:"use_ahp"`
	EventComplexity *string `query:"event_complexity"`
}
