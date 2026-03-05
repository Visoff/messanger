package dtos

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/Visoff/messanger/pkgs/httperrors"
)

type ValidationError struct {
	errors []string
}

func (e *ValidationError) Error() string {
	b, _ := json.Marshal(e.errors)
	return string(b)
}

type DTO interface {
	Validate() error
}

func Validate(dto DTO) error {
	if err := dto.Validate(); err != nil {
		return err
	}
	return nil
}

func ParseJson(r *http.Request, dto DTO) error {
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		return httperrors.NewHTTPJSONParsingError(err)
	}
	return dto.Validate()
}

func ParseForm(r *http.Request, dto DTO) error {
	if err := r.ParseForm(); err != nil {
		return httperrors.NewHTTPFormParsingError(err)
	}

	dto_type := reflect.TypeOf(dto)
	if dto_type.Kind() == reflect.Pointer {
		dto_type = dto_type.Elem()
	}

	for i := 0; i < dto_type.NumField(); i++ {
		field := dto_type.Field(i)
		if val := r.FormValue(field.Tag.Get("json")); val != "" {
			reflect.ValueOf(dto).Elem().Field(i).SetString(val)
		}
	}

	return dto.Validate()
}

func ParseFromBody(r *http.Request, dto DTO) error {
	if r.Header.Get("Content-Type") == "application/json" {
		return ParseJson(r, dto)
	}

	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		return ParseForm(r, dto)
	}

	return httperrors.NewHTTPUnsupportedMediaTypeError()
}
