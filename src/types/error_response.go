package types

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func CreateErrorResponse(errorMessage string) *ErrorResponse {
	return &ErrorResponse{
		Status:  "error",
		Message: errorMessage,
	}
}
