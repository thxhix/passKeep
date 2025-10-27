package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = "passkeeper"
	keyringUser    = "session"
)

type Tokens struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

func SaveTokens(tokens Tokens) error {
	b, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	return keyring.Set(keyringService, keyringUser, string(b))
}

func LoadTokens() (Tokens, error) {
	var t Tokens
	s, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return t, nil
		}
		fmt.Println(err)
		return t, err
	}
	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return t, err
	}
	return t, nil
}

func DeleteTokens() error {
	return keyring.Delete(keyringService, keyringUser)
}
