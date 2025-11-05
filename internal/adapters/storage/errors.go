package storage

import "errors"

var (
	ErrNoPostgresConnection = errors.New("no postgresql connection provided")
)
