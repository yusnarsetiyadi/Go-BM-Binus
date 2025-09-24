package response

type MetaSuccess struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

type MetaError struct {
	Success      bool        `json:"success"`
	Code         int         `json:"code"`
	Data         interface{} `json:"data"`
	errorMessage error
}
