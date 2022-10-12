package types

type DataResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

func CreateDataResponse(status string, responseData interface{}) *DataResponse {
	return &DataResponse{
		Status: status,
		Data:   responseData,
	}
}
