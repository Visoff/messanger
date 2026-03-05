package handlers

import (
	"net/http"

	"github.com/Visoff/messanger/pkgs/httperrors"
)

type Handler func(http.ResponseWriter, *http.Request) error

func (f Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f(w, r)
	if err != nil {
		httperrors.WriteError(w, err)
	}
}
