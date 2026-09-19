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
	Name     string
	ImageURL string
	VideoURL string
	Power    float64
	Cost     float64
}

type PublishDraftData struct {
	Description string
}

type HeatingEquipmentRepository interface {
	GetByID(id uint) (*model.HeatingEquipment, error)
	GetNextPublished(currentID uint) (*model.HeatingEquipment, error)
	GetFirstPublished() (*model.HeatingEquipment, error)
	GetDraft(creatorID uint) (*model.HeatingEquipment, error)
	GetAllPublished(minPower float64) ([]model.HeatingEquipment, error)
	CreateDraft(creatorID uint, data DraftData) (*model.HeatingEquipment, error)
	PublishDraft(creatorID, equipmentID uint, data PublishDraftData) error
	DeleteEquipmentSQL(equipmentID uint) error
}

type postgresHeatingEquipmentRepository struct {
	db *gorm.DB
}

func NewPostgresHeatingEquipmentRepository(dsn string) (HeatingEquipmentRepository, error) {
	if dsn == "" {
		return nil, errors.New("пустая строка подключения к PostgreSQL: проверьте .env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &postgresHeatingEquipmentRepository{db: db}, nil
}

func (r *postgresHeatingEquipmentRepository) GetByID(id uint) (*model.HeatingEquipment, error) {
	var item model.HeatingEquipment
	err := r.db.Preload("Likes").
		Where("id = ? AND status = ?", id, model.StatusPublished).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postgresHeatingEquipmentRepository) GetFirstPublished() (*model.HeatingEquipment, error) {
	var item model.HeatingEquipment
	err := r.db.Preload("Likes").
		Where("status = ?", model.StatusPublished).
		Order("id ASC").First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postgresHeatingEquipmentRepository) GetNextPublished(currentID uint) (*model.HeatingEquipment, error) {
	var item model.HeatingEquipment
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

func (r *postgresHeatingEquipmentRepository) GetDraft(creatorID uint) (*model.HeatingEquipment, error) {
	var item model.HeatingEquipment
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

func (r *postgresHeatingEquipmentRepository) GetAllPublished(minPower float64) ([]model.HeatingEquipment, error) {
	var items []model.HeatingEquipment
	q := r.db.Preload("Likes").Where("status = ?", model.StatusPublished)
	if minPower > 0 {
		q = q.Where("power_kw >= ?", minPower)
	}
	if err := q.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *postgresHeatingEquipmentRepository) CreateDraft(creatorID uint, data DraftData) (*model.HeatingEquipment, error) {
	existing, err := r.GetDraft(creatorID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	item := model.HeatingEquipment{
		Name:      data.Name,
		ImageURL:  data.ImageURL,
		VideoURL:  data.VideoURL,
		Power:     &data.Power,
		Cost:      &data.Cost,
		Status:    model.StatusDraft,
		CreatedAt: time.Now(),
		CreatorID: creatorID,
	}

	if err := r.db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postgresHeatingEquipmentRepository) PublishDraft(
	creatorID uint,
	equipmentID uint,
	data PublishDraftData,
) error {

	formedAt := time.Now()

	result := r.db.Model(&model.HeatingEquipment{}).
		Where(
			"id = ? AND creator_id = ? AND status = ?",
			equipmentID,
			creatorID,
			model.StatusDraft,
		).
		Updates(map[string]any{
			"short_description": data.Description,
			"status":            model.StatusPublished,
			"formed_at":         formedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("черновик не найден")
	}

	return nil
}

// DeleteEquipmentSQL — логическое удаление чистым SQL, без ORM.
func (r *postgresHeatingEquipmentRepository) DeleteEquipmentSQL(equipmentID uint) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	result, err := sqlDB.Exec(
		"UPDATE heating_equipment SET status = $1 WHERE id = $2 AND status <> $1",
		model.StatusDeleted, equipmentID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("отопительное оборудование с id %d не найдено или уже удалено", equipmentID)
	}
	return nil
}
