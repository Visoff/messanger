package handlers

import (
	"net/http"

	"github.com/Visoff/messanger/pkgs/httperrors"
	"github.com/google/uuid"
)

type Handler func(http.ResponseWriter, *http.Request) error

func ToErrorHandler(h http.Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}
}

type Middleware func(Handler) Handler

func (f Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f(w, r)
	if err != nil {
		httperrors.WriteError(w, err)
	}
}

type RouteMux struct {
	base string
	mux  *http.ServeMux
}

func NewRouteMux(mux *http.ServeMux, base string) *RouteMux {
	if base[0] != '/' {
		base = "/" + base
	}
	if base[len(base)-1] != '/' {
		base = base + "/"
	}
	return &RouteMux{
		base: base,
		mux:  mux,
	}
}

func (r *RouteMux) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(r.base+pattern, http.StripPrefix(r.base, handler))
}

func (r *RouteMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func GetParamID(r *http.Request, key string) (uuid.UUID, error) {
	id := r.PathValue(key)
	if id == "" {
		return uuid.Nil, httperrors.NewHTTPNotFoundError("ID not found")
	}
	res, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, httperrors.NewHTTPNotFoundError("ID not found")
	}
	return res, nil
}
