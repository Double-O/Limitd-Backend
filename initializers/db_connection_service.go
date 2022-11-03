package initializers

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDbConnection() (*gorm.DB, error) {
	dsn := os.Getenv("POSTGRES_DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Error().
			Str("file", "repository.db_connection_service").
			Str("function", "NewDbConnection").
			Msg(fmt.Sprintf("Db connection failed with dsn : %s", dsn))
		return nil, err
	}
	return db, nil
}
