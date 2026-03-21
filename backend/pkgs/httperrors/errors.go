package httperrors

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message"`

	Details map[string]string `json:"details,omitempty"`
}

type HTTPError interface {
	Error() string
	StatusCode() int
	Status() string
}

func WriteError(w http.ResponseWriter, err error) {
	if httperr, ok := err.(HTTPError); ok {
		w.WriteHeader(httperr.StatusCode())
		w.Write(fmt.Appendf(nil, `{"error":"%s","status":%d,"message":"%s"}`, httperr.Status(), httperr.StatusCode(), httperr.Error()))
	} else {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

type HTTPBadRequestError struct {
	msg string
}

func NewHTTPBadRequestError(msg string) *HTTPBadRequestError {
	return &HTTPBadRequestError{msg: msg}
}

func (e *HTTPBadRequestError) Error() string {
	return e.msg
}

func (e *HTTPBadRequestError) StatusCode() int {
	return 400
}

func (e *HTTPBadRequestError) Status() string {
	return "Bad request"
}

type HTTPBodyParsingError struct {
	err error
}

func NewHTTPBodyParsingError(err error) *HTTPBodyParsingError {
	return &HTTPBodyParsingError{err: err}
}

func (e *HTTPBodyParsingError) Error() string {
	return e.err.Error()
}

func (e *HTTPBodyParsingError) StatusCode() int {
	return 400
}

func (e *HTTPBodyParsingError) Status() string {
	return "Body parsing error"
}

type HTTPValidationError struct {
	err map[string]string
}

func NewHTTPValidationError(err map[string]string) error {
	if len(err) == 0 {
		return nil
	}
	return &HTTPValidationError{err: err}
}

func (e *HTTPValidationError) Error() string {
	b, _ := json.Marshal(e.err)
	return string(b)
}

func (e *HTTPValidationError) StatusCode() int {
	return 400
}

func (e *HTTPValidationError) Status() string {
	return "Validation error"
}

type HTTPJSONParsingError struct {
	err error
}

func NewHTTPJSONParsingError(err error) *HTTPJSONParsingError {
	return &HTTPJSONParsingError{err: err}
}

func (e *HTTPJSONParsingError) Error() string {
	return e.err.Error()
}

func (e *HTTPJSONParsingError) StatusCode() int {
	return 400
}

func (e *HTTPJSONParsingError) Status() string {
	return "JSON parsing error"
}

type HTTPFormParsingError struct {
	err error
}

func NewHTTPFormParsingError(err error) *HTTPFormParsingError {
	return &HTTPFormParsingError{err: err}
}

func (e *HTTPFormParsingError) Error() string {
	return e.err.Error()
}

func (e *HTTPFormParsingError) StatusCode() int {
	return 400
}

func (e *HTTPFormParsingError) Status() string {
	return "Form parsing error"
}

type HTTPUnsupportedMediaTypeError struct{}

func NewHTTPUnsupportedMediaTypeError() *HTTPUnsupportedMediaTypeError {
	return &HTTPUnsupportedMediaTypeError{}
}

func (e *HTTPUnsupportedMediaTypeError) Error() string {
	return "Unsupported media type"
}

func (e *HTTPUnsupportedMediaTypeError) StatusCode() int {
	return 415
}

func (e *HTTPUnsupportedMediaTypeError) Status() string {
	return "Unsupported media type"
}

type HTTPNotFoundError struct {
	msg string
}

func NewHTTPNotFoundError(msg string) *HTTPNotFoundError {
	return &HTTPNotFoundError{msg: msg}
}

func (e *HTTPNotFoundError) Error() string {
	return e.msg
}

func (e *HTTPNotFoundError) StatusCode() int {
	return 404
}

func (e *HTTPNotFoundError) Status() string {
	return "Not found"
}

type HTTPUnauthorizedError struct {
	msg string
}

func NewHTTPUnauthorizedError(msg string) *HTTPUnauthorizedError {
	return &HTTPUnauthorizedError{msg: msg}
}

func (e *HTTPUnauthorizedError) Error() string {
	return e.msg
}

func (e *HTTPUnauthorizedError) StatusCode() int {
	return 401
}

func (e *HTTPUnauthorizedError) Status() string {
	return "Unauthorized"
}

type HTTPConflictError struct {
	msg string
}

func NewHTTPConflictError(msg string) *HTTPConflictError {
	return &HTTPConflictError{msg: msg}
}

func (e *HTTPConflictError) Error() string {
	return e.msg
}

func (e *HTTPConflictError) StatusCode() int {
	return 409
}

func (e *HTTPConflictError) Status() string {
	return "Conflict"
}
