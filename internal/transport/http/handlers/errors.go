package handlers

import "errors"

var (
	ErrCantWriteResponseBody = errors.New("cant write response body")

	ErrBadQuery            = errors.New("bad query params provided")
	ErrBadRequest          = errors.New("bad request, check your request body")
	ErrUnauthorizedError   = errors.New("wrong provided auth token")
	ErrInternalServerError = errors.New("internal server error")
	ErrNotFound            = errors.New("resource not found")

	ErrPayloadFileLimit    = errors.New("payload file is too large, max 10Mb")
	ErrPayloadFileNotFound = errors.New("payload file not found")

	ErrInternalPublicError = errors.New("Something went wrong..")
)
