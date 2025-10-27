package client_http

import (
	"errors"
	"fmt"
)

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: status=%d body=%s", e.StatusCode, e.Body)
}

var (
	ErrTooManyRedirects = errors.New("too many redirects")

	ErrPathIsEmpty  = errors.New("can't do request with empty path")
	ErrEmptyBaseURL = errors.New("empty base url provided")
	ErrEmptyMethod  = errors.New("empty http method")

	ErrClientRequest       = errors.New("client request error")
	ErrClientRequestFailed = errors.New("client request failed")

	ErrServerResponse          = errors.New("server response provided error")
	ErrServerResponseUnmarshal = errors.New("server response unmarshal error")
	ErrServerResponseTooLarge  = errors.New("server response too large, max - 10 Mb")

	ErrCantMarshalBody = errors.New("body marshalling error")
	ErrCantReadBody    = errors.New("body read error")
)
