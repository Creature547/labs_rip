package repository

import (
	"awesomeProject/internal/app/model"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DraftData struct {
	Title              string
	ImageURL           string
	VideoURL           string
	Description        string
	Power              float64
	Cost               float64
	Area               float64
	CeilingHeight      float64
	FloorType          string
	WallMaterial       string
	DesiredTemperature float64
	ElectricityTariff  float64
}

type PublishDraftData struct {
	Description        string
	Power              float64
	Cost               float64
	Area               float64
	CeilingHeight      float64
	FloorType          string
	WallMaterial       string
	DesiredTemperature float64
	ElectricityTariff  float64
}

type HeatingRepository interface {
	GetByID(id uint) (*model.HeatingCalculation, error)
	GetNextPublished(currentID uint) (*model.HeatingCalculation, error)
	GetFirstPublished() (*model.HeatingCalculation, error)
	GetDraft(creatorID uint) (*model.HeatingCalculation, error)
	GetAllPublished(minPower float64) ([]model.HeatingCalculation, error)
	CreateDraft(creatorID uint, data DraftData) (*model.HeatingCalculation, error)
	PublishDraft(creatorID, calculationID uint, data PublishDraftData) error
	DeleteCalculationSQL(calculationID uint) error
}

type postgresHeatingRepository struct {
	db *gorm.DB
}

func NewPostgresHeatingRepository(dsn string) (HeatingRepository, error) {
	if dsn == "" {
		return nil, errors.New("пустая строка подключения к PostgreSQL: проверьте .env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &postgresHeatingRepository{db: db}, nil
}

func (r *postgresHeatingRepository) GetByID(id uint) (*model.HeatingCalculation, error) {
	var item model.HeatingCalculation
	err := r.db.Preload("Likes").
		Where("id = ? AND status = ?", id, model.StatusPublished).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postgresHeatingRepository) GetFirstPublished() (*model.HeatingCalculation, error) {
	var item model.HeatingCalculation
	err := r.db.Preload("Likes").
		Where("status = ?", model.StatusPublished).
		Order("id ASC").First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postgresHeatingRepository) GetNextPublished(currentID uint) (*model.HeatingCalculation, error) {
	var item model.HeatingCalculation
	err := r.db.Preload("Likes").
		Where("status = ? AND id > ?", model.StatusPublished, currentID).
		Order("id ASC").First(&item).Error
	if err == nil {
		return &item, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return r.GetFirstPublished()
}

func (r *postgresHeatingRepository) GetDraft(creatorID uint) (*model.HeatingCalculation, error) {
	var item model.HeatingCalculation
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, model.StatusDraft).
		Order("id DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postgresHeatingRepository) GetAllPublished(minPower float64) ([]model.HeatingCalculation, error) {
	var items []model.HeatingCalculation
	q := r.db.Preload("Likes").Where("status = ?", model.StatusPublished)
	if minPower > 0 {
		q = q.Where("power_kw >= ?", minPower)
	}
	if err := q.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *postgresHeatingRepository) CreateDraft(creatorID uint, data DraftData) (*model.HeatingCalculation, error) {
	existing, err := r.GetDraft(creatorID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	item := model.HeatingCalculation{
		Title:              data.Title,
		ImageURL:           data.ImageURL,
		VideoURL:           data.VideoURL,
		Description:        data.Description,
		Power:              data.Power,
		Cost:               data.Cost,
		Area:               data.Area,
		CeilingHeight:      data.CeilingHeight,
		FloorType:          data.FloorType,
		WallMaterial:       data.WallMaterial,
		DesiredTemperature: data.DesiredTemperature,
		ElectricityTariff:  data.ElectricityTariff,
		Status:             model.StatusDraft,
		CreatedAt:          time.Now(),
		CreatorID:          creatorID,
	}

	if err := r.db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postgresHeatingRepository) PublishDraft(creatorID, calculationID uint, data PublishDraftData) error {
	formedAt := time.Now()
	result := r.db.Model(&model.HeatingCalculation{}).
		Where("id = ? AND creator_id = ? AND status = ?", calculationID, creatorID, model.StatusDraft).
		Updates(map[string]any{
			"description":                data.Description,
			"power_kw":                   data.Power,
			"cost_rub":                   data.Cost,
			"area_m2":                    data.Area,
			"ceiling_height_m":           data.CeilingHeight,
			"floor_type":                 data.FloorType,
			"wall_material":              data.WallMaterial,
			"desired_temperature_c":      data.DesiredTemperature,
			"electricity_tariff_rub_kwh": data.ElectricityTariff,
			"status":                     model.StatusPublished,
			"formed_at":                  formedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("черновик для публикации не найден")
	}
	return nil
}


func (r *postgresHeatingRepository) DeleteCalculationSQL(calculationID uint) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	result, err := sqlDB.Exec(
		"UPDATE heating_calculations SET status = $1 WHERE id = $2 AND status <> $1",
		model.StatusDeleted, calculationID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("расчет отопления с id %d не найден или уже удален", calculationID)
	}
	return nil
}
