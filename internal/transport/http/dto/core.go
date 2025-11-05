package dto

//go:generate easyjson -all core.go

type ErrorResponse struct {
	Code      int    `json:"code"`
	ErrorText string `json:"error"`
}
