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

	// Сохраняем данные старой версии лабораторной при переходе
	// heating_calculations -> heating_equipment.
	if db.Migrator().HasTable("heating_calculations") && !db.Migrator().HasTable("heating_equipment") {
		if err := db.Exec(`ALTER TABLE heating_calculations RENAME TO heating_equipment`).Error; err != nil {
			log.Fatalf("cant rename heating_calculations: %v", err)
		}
	}

	if db.Migrator().HasTable("heating_likes") && !db.Migrator().HasTable("heating_equipment_likes") {
		if err := db.Exec(`ALTER TABLE heating_likes RENAME TO heating_equipment_likes`).Error; err != nil {
			log.Fatalf("cant rename heating_likes: %v", err)
		}
	}

	if db.Migrator().HasTable("heating_equipment") {
		if db.Migrator().HasColumn("heating_equipment", "title") && !db.Migrator().HasColumn("heating_equipment", "name") {
			if err := db.Exec(`ALTER TABLE heating_equipment RENAME COLUMN title TO name`).Error; err != nil {
				log.Fatalf("cant rename title column: %v", err)
			}
		}
	}

	if db.Migrator().HasTable("heating_equipment_likes") {
		if db.Migrator().HasColumn("heating_equipment_likes", "heating_calculation_id") && !db.Migrator().HasColumn("heating_equipment_likes", "heating_equipment_id") {
			if err := db.Exec(`ALTER TABLE heating_equipment_likes RENAME COLUMN heating_calculation_id TO heating_equipment_id`).Error; err != nil {
				log.Fatalf("cant rename like foreign key: %v", err)
			}
		}
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.HeatingEquipment{},
		&model.HeatingEquipmentLike{},
	); err != nil {
		log.Fatalf("cant migrate db: %v", err)
	}

	// Удаляем лишние поля старого "расчета дома": в услуге должны остаться
	// только два предметных поля — мощность и стоимость.
	for _, column := range []string{
		"area_m2",
		"ceiling_height_m",
		"floor_type",
		"wall_material",
		"desired_temperature_c",
		"electricity_tariff_rub_kwh",
	} {
		if db.Migrator().HasColumn("heating_equipment", column) {
			if err := db.Migrator().DropColumn("heating_equipment", column); err != nil {
				log.Fatalf("cant drop old column %s: %v", column, err)
			}
		}
	}

	// Один пользователь — не более одного черновика.
	if err := db.Exec(`DROP INDEX IF EXISTS idx_one_heating_draft_per_creator`).Error; err != nil {
		log.Fatalf("cant drop old draft index: %v", err)
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_one_heating_equipment_draft_per_creator
		ON heating_equipment (creator_id)
		WHERE status = 'черновик'
	`).Error; err != nil {
		log.Fatalf("cant create draft constraint: %v", err)
	}
}
