package migration

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func DoMigration(direction string, dbConnStr string) {
	log := zap.S()
	m, err := migrate.New("file://db/migrations", dbConnStr)
	if err != nil {
		log.Fatalf("Error getting migration files: %v", err)
	}

	if direction == "up" {
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Infof("%v", err)
			} else {
				log.Fatalf("Failed to migrate up: %v", err)
			}
		}
	}
	if direction == "down" {
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				log.Infof("%v", err)
			} else {
				log.Fatalf("Failed to migrate down: %v", err)
			}
		}
	}
}
