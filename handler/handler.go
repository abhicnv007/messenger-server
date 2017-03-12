package handler

import "net/http"
import "github.com/abhicnv007/whistle/response"

//Handler is an object that is passed to the serve mux to handle routes
type Handler struct {

	//AuthHandler Handles authentication and/or authorisation of the request
	AuthHandler func(w http.ResponseWriter, r *http.Request) error

	//MainHandler Handles the actual work of the request
	MainHandler func(w http.ResponseWriter, r *http.Request)

	//	Values map[interface{}]interface{}
}

//SetGlobalAuthFunc sets the suth function for all handlers
// NOTE: Has to ba called before creating any new handlers
func SetGlobalAuthFunc(fn func(w http.ResponseWriter, r *http.Request) error) {
	rootAuthHandler = fn
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	if h.AuthHandler != nil {
		if err = h.AuthHandler(w, r); err != nil {
			response.New(w).WithCode(http.StatusUnauthorized).
				Error(err.Error())
			return
		}
	}

	if h.MainHandler != nil {
		h.MainHandler(w, r)
	}
}

//NoAuth ensures the handler does not uthenticate the request
func (h *Handler) NoAuth() *Handler {
	h.AuthHandler = nil
	return h
}

var rootAuthHandler func(w http.ResponseWriter, r *http.Request) error

//New Returns a new handler
func New(fn func(w http.ResponseWriter, r *http.Request)) *Handler {
	return &Handler{
		AuthHandler: rootAuthHandler,
		MainHandler: fn,
	}
}
