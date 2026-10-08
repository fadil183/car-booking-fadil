package config

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {

	postgressDSN := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_PSQL_HOST"),
		os.Getenv("DB_PSQL_USR"),
		os.Getenv("DB_PSQL_PW"),
		os.Getenv("DB_PSQL_NAME"),
		os.Getenv("DB_PSQL_PORT"),
	)
	db, err := gorm.Open(postgres.Open(postgressDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	return db.Debug()
}
