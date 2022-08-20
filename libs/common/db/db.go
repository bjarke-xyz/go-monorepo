package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ConnectionStringer interface {
	ConnectionString() string
}

func Connect(connStringer ConnectionStringer) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", connStringer.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return db, nil
}
