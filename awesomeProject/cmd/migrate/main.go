package main

import (
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/model"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.HeatingCalculation{},
		&model.HeatingLike{},
	); err != nil {
		log.Fatalf("cant migrate db: %v", err)
	}

	// У каждого пользователя может быть не более одного отопительного расчета-черновика.
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_one_heating_draft_per_creator
		ON heating_calculations (creator_id)
		WHERE status = 'черновик'
	`).Error; err != nil {
		log.Fatalf("cant create draft constraint: %v", err)
	}
}
