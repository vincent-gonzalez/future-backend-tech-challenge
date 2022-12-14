package middleware

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// HeaderMiddleware sets the content type of responses to JSON
func HeaderMiddleware(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		h(w, r, ps)
	}
}
