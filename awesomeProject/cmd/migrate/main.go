package main

import (
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/model"
	"log"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Compatibility with the lab-2 schema.
	if db.Migrator().HasTable("heating_calculations") && !db.Migrator().HasTable("heating_equipment") {
		mustExec(db, `ALTER TABLE heating_calculations RENAME TO heating_equipment`)
	}
	if db.Migrator().HasTable("heating_likes") && !db.Migrator().HasTable("heating_equipment_likes") {
		mustExec(db, `ALTER TABLE heating_likes RENAME TO heating_equipment_likes`)
	}
	if db.Migrator().HasTable("heating_equipment") {
		renameColumnIfNeeded(db, "heating_equipment", "title", "equipment_name")
		renameColumnIfNeeded(db, "heating_equipment", "name", "equipment_name")
		renameColumnIfNeeded(db, "heating_equipment", "image_url", "image_name")
		renameColumnIfNeeded(db, "heating_equipment", "video_url", "video_name")
	}
	if db.Migrator().HasTable("users") {
		renameColumnIfNeeded(db, "users", "password", "password_hash")
	}
	if db.Migrator().HasTable("heating_equipment_likes") {
		renameColumnIfNeeded(db, "heating_equipment_likes", "heating_calculation_id", "heating_equipment_id")
	}

	// Compatibility: old lab-2 tables already contain rows,
	// so new NOT NULL timestamps must be filled before AutoMigrate.
	if db.Migrator().HasTable("users") {
		ensureTimestampColumn(db, "users", "created_at")
	}

	if db.Migrator().HasTable("heating_equipment_likes") {
		ensureTimestampColumn(db, "heating_equipment_likes", "created_at")
	}

	if err := db.AutoMigrate(&model.User{}, &model.HeatingEquipment{}, &model.HeatingEquipmentLike{}); err != nil {
		log.Fatalf("cant migrate db: %v", err)
	}

	// Remove obsolete subject fields from older versions.
	for _, column := range []string{
		"area_m2", "ceiling_height_m", "floor_type", "wall_material",
		"desired_temperature_c", "electricity_tariff_rub_kwh",
	} {
		if db.Migrator().HasColumn("heating_equipment", column) {
			if err := db.Migrator().DropColumn("heating_equipment", column); err != nil {
				log.Fatalf("cant drop old column %s: %v", column, err)
			}
		}
	}

	mustExec(db, `DROP INDEX IF EXISTS idx_one_heating_draft_per_creator`)
	mustExec(db, `DROP INDEX IF EXISTS idx_one_heating_equipment_draft_per_creator`)
	mustExec(db, `CREATE UNIQUE INDEX idx_one_heating_equipment_draft_per_creator ON heating_equipment (creator_id) WHERE status = 'черновик'`)

	seedCurrentUser(db)
	mustExec(db, `SELECT setval(pg_get_serial_sequence('users','id'), GREATEST(COALESCE((SELECT MAX(id) FROM users), 1), 1), true)`)
}

func seedCurrentUser(db *gorm.DB) {
	var count int64
	if err := db.Model(&model.User{}).Where("id = 1").Count(&count).Error; err != nil {
		log.Fatalf("cant check current user: %v", err)
	}
	if count > 0 {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("creator123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("cant hash seed password: %v", err)
	}
	user := model.User{ID: 1, Login: "creator", PasswordHash: string(hash), CreatedAt: time.Now()}
	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("cant seed current user: %v", err)
	}
}

func renameColumnIfNeeded(db *gorm.DB, table, oldName, newName string) {
	if db.Migrator().HasColumn(table, oldName) && !db.Migrator().HasColumn(table, newName) {
		mustExec(db, `ALTER TABLE `+table+` RENAME COLUMN `+oldName+` TO `+newName)
	}
}

func mustExec(db *gorm.DB, sql string) {
	if err := db.Exec(sql).Error; err != nil {
		log.Fatalf("migration sql failed: %v", err)
	}
}
func ensureTimestampColumn(db *gorm.DB, table, column string) {
	if !db.Migrator().HasColumn(table, column) {
		mustExec(db, `ALTER TABLE `+table+` ADD COLUMN `+column+` timestamptz`)
	}

	mustExec(db, `UPDATE `+table+` SET `+column+` = NOW() WHERE `+column+` IS NULL`)
	mustExec(db, `ALTER TABLE `+table+` ALTER COLUMN `+column+` SET NOT NULL`)
}