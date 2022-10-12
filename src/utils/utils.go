package utils

import (
	"encoding/json"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/types"
	"net/http"
	"time"
)

const Datetime_format_without_t = "2006-01-02 15:04:05-07:00"
const Datetime_format = "2006-01-02T15:04:05-07:00"

func ConvertDateTimeToRFC3339(dateTime string, currentFormat string) string {
	temp, _ := time.Parse(currentFormat, dateTime)
	return temp.Format(time.RFC3339)
}

func writeResponse(w http.ResponseWriter, responseBody interface{}, responseCode int) {
	w.WriteHeader(responseCode)
	json.NewEncoder(w).Encode(responseBody)
}

func WriteDataResponse(w http.ResponseWriter, status string, data interface{}, responseCode int) {
	dataResponse := types.CreateDataResponse(status, data)
	writeResponse(w, dataResponse, responseCode)
}

func WriteErrorResponse(w http.ResponseWriter, errorMessage string, errorCode int) {
	errorResponse := types.CreateErrorResponse(errorMessage)
	writeResponse(w, errorResponse, errorCode)
}
