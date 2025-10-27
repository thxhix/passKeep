package dto

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"time"
)

//go:generate easyjson -all keychain.go

type GetKeysRecord struct {
	KeyUUID   uuid.UUID        `json:"uuid"`
	KeyType   keychain.KeyType `json:"type"`
	Title     string           `json:"title"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type GetKeysResponse struct {
	Keys []*GetKeysRecord `json:"keys"`
}

type GetKeyResponse struct {
	KeyUUID   uuid.UUID        `json:"uuid"`
	KeyType   keychain.KeyType `json:"type"`
	Title     string           `json:"title"`
	Data      json.RawMessage  `json:"data"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type AddSuccessResponse struct {
	UUID string `json:"key_uuid"`
}

type AddCredentialsDTO struct {
	Title    string `json:"title"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Site     string `json:"site,omitempty"`
	Note     string `json:"note,omitempty"`
}

type CredentialsResponseDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Site     string `json:"site,omitempty"`
	Note     string `json:"note,omitempty"`
}

type AddCardDTO struct {
	Title   string `json:"title"`
	Number  string `json:"number"`
	ExpDate string `json:"exp_date"`
	CVV     string `json:"cvv"`
	Holder  string `json:"holder"`
	Bank    string `json:"bank,omitempty"`
	Note    string `json:"note,omitempty"`
}

type CardResponseDTO struct {
	Number  string `json:"number"`
	ExpDate string `json:"exp_date"`
	CVV     string `json:"cvv"`
	Holder  string `json:"holder"`
	Bank    string `json:"bank,omitempty"`
	Note    string `json:"note,omitempty"`
}

type AddTextDTO struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Note  string `json:"note,omitempty"`
}

type TextResponseDTO struct {
	Text string `json:"text"`
	Note string `json:"note,omitempty"`
}

type AddFileDTO struct {
	Title string `json:"title"`
	File  []byte `json:"-"`
	Note  string `json:"note,omitempty"`
}

type FileResponseDTO struct {
	File []byte `json:"-"`
	Note string `json:"note,omitempty"`
}
