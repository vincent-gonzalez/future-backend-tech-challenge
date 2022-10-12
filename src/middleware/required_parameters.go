package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/types"
	"net/http"
)

func RequiredParameters(h httprouter.Handle, params ...string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		query := r.URL.Query()
		for _, parameter := range params {
			if len(query.Get(parameter)) < 1 {
				response := &types.ErrorResponse{
					Status:  "error",
					Message: fmt.Sprintf("%s is required", parameter),
				}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		h(w, r, ps)
	}
}
