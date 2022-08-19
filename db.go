package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func GetDb(config *Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", config.GetDbConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return db, nil
}
