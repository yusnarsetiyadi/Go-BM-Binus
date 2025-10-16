package dto

type CommentCreateRequest struct {
	RequestId int    `json:"request_id" form:"request_id" validate:"required"`
	Comment   string `json:"comment" form:"comment" validate:"required"`
}

type CommentFindByRequestIDRequest struct {
	RequestId int `param:"request_id" validate:"required"`
}

type CommentDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type CommentUpdateRequest struct {
	ID      int     `param:"id" validate:"required"`
	Comment *string `json:"comment" form:"comment"`
}
