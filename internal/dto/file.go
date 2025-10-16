package dto

import "mime/multipart"

type FileCreateRequest struct {
	RequestId int `json:"request_id" form:"request_id" validate:"required"`
	Files     []*multipart.FileHeader
}

type FileFindByRequestIDRequest struct {
	RequestId int `param:"request_id" validate:"required"`
}

type FileDeleteByIDRequest struct {
	ID int `param:"id" validate:"required"`
}

type FileUpdateRequest struct {
	ID   int     `param:"id" validate:"required"`
	Name *string `json:"name" form:"name"`
}
