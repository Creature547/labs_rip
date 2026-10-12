package repository

import (
	"awesomeProject/internal/app/model"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrDraftExists      = errors.New("draft already exists")
	ErrInvalidLikeValue = errors.New("like must be 0 or 1")
	ErrLoginExists      = errors.New("login already exists")
)

type Settings struct {
	PostgresDSN    string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
}

type EquipmentFilter struct {
	Name     string
	MinPower *float64
	MaxCost  *float64
}

type DraftData struct {
	Name        string
	Description string
	Power       float64
	Cost        float64
}

type HeatingEquipmentRepository interface {
	ListPublished(filter EquipmentFilter) ([]model.HeatingEquipment, error)
	GetDraft(creatorID uint) (*model.HeatingEquipment, error)
	CreateDraft(creatorID uint, data DraftData, image, video *multipart.FileHeader) (*model.HeatingEquipment, error)
	PublishDraft(creatorID, equipmentID uint) (*model.HeatingEquipment, error)
	GetFirstPublished() (*model.HeatingEquipment, error)
	GetPublishedByID(id uint) (*model.HeatingEquipment, error)
	GetNextPublished(currentID uint) (*model.HeatingEquipment, error)
	SetLike(userID, equipmentID uint, value int) (*model.HeatingEquipment, error)
	SoftDelete(creatorID, equipmentID uint) error
	MediaBucket() string
}

type UserRepository interface {
	RegisterUser(login, passwordHash string) (*model.User, error)
	GetUserByLogin(login string) (*model.User, error)
}

type PostgresRepository struct {
	db     *gorm.DB
	minio  *minio.Client
	bucket string
}

func New(settings Settings) (*PostgresRepository, error) {
	if strings.TrimSpace(settings.PostgresDSN) == "" {
		return nil, errors.New("empty postgres dsn")
	}
	if settings.MinioEndpoint == "" || settings.MinioAccessKey == "" || settings.MinioSecretKey == "" || settings.MinioBucket == "" {
		return nil, errors.New("incomplete minio settings")
	}

	db, err := gorm.Open(postgres.Open(settings.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	minioClient, err := minio.New(settings.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.MinioAccessKey, settings.MinioSecretKey, ""),
		Secure: settings.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	repo := &PostgresRepository{db: db, minio: minioClient, bucket: settings.MinioBucket}
	if err := repo.ensureBucket(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *PostgresRepository) MediaBucket() string { return r.bucket }

func (r *PostgresRepository) ensureBucket(ctx context.Context) error {
	exists, err := r.minio.BucketExists(ctx, r.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := r.minio.MakeBucket(ctx, r.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	// Public read is useful for the SPA: DB stores only object names,
	// while the serializer builds URLs for the client.
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, r.bucket)
	return r.minio.SetBucketPolicy(ctx, r.bucket, policy)
}

func (r *PostgresRepository) ListPublished(filter EquipmentFilter) ([]model.HeatingEquipment, error) {
	items := make([]model.HeatingEquipment, 0)
	q := r.db.Preload("Likes").Where("status = ?", model.StatusPublished)
	if filter.Name != "" {
		q = q.Where("LOWER(equipment_name) LIKE ?", "%"+strings.ToLower(filter.Name)+"%")
	}
	if filter.MinPower != nil {
		q = q.Where("power_kw >= ?", *filter.MinPower)
	}
	if filter.MaxCost != nil {
		q = q.Where("cost_rub <= ?", *filter.MaxCost)
	}
	if err := q.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) GetDraft(creatorID uint) (*model.HeatingEquipment, error) {
	var item model.HeatingEquipment
	err := r.db.Preload("Likes").Where("creator_id = ? AND status = ?", creatorID, model.StatusDraft).
		Order("id DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *PostgresRepository) CreateDraft(creatorID uint, data DraftData, image, video *multipart.FileHeader) (*model.HeatingEquipment, error) {
	var count int64
	if err := r.db.Model(&model.HeatingEquipment{}).
		Where("creator_id = ? AND status = ?", creatorID, model.StatusDraft).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrDraftExists
	}

	imageName, err := generatedObjectName("image", image.Filename)
	if err != nil {
		return nil, err
	}
	videoName, err := generatedObjectName("video", video.Filename)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err := r.putFile(ctx, imageName, image); err != nil {
		return nil, err
	}
	if err := r.putFile(ctx, videoName, video); err != nil {
		_ = r.minio.RemoveObject(ctx, r.bucket, imageName, minio.RemoveObjectOptions{})
		return nil, err
	}

	description := data.Description
	item := model.HeatingEquipment{
		Name:        data.Name,
		Description: &description,
		Power:       &data.Power,
		Cost:        &data.Cost,
		ImageName:   imageName,
		VideoName:   videoName,
		Status:      model.StatusDraft,
		CreatorID:   creatorID,
		CreatedAt:   time.Now(),
	}

	if err := r.db.Create(&item).Error; err != nil {
		_ = r.minio.RemoveObject(ctx, r.bucket, imageName, minio.RemoveObjectOptions{})
		_ = r.minio.RemoveObject(ctx, r.bucket, videoName, minio.RemoveObjectOptions{})
		return nil, err
	}
	item.Likes = []model.HeatingEquipmentLike{}
	return &item, nil
}

func (r *PostgresRepository) PublishDraft(creatorID, equipmentID uint) (*model.HeatingEquipment, error) {
	formedAt := time.Now()
	result := r.db.Model(&model.HeatingEquipment{}).
		Where("id = ? AND creator_id = ? AND status = ?", equipmentID, creatorID, model.StatusDraft).
		Updates(map[string]any{"status": model.StatusPublished, "formed_at": formedAt})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetPublishedByID(equipmentID)
}

func (r *PostgresRepository) GetFirstPublished() (*model.HeatingEquipment, error) {
	var item model.HeatingEquipment
	err := r.db.Preload("Likes").Where("status = ?", model.StatusPublished).Order("id ASC").First(&item).Error
	return normalizeNotFound(&item, err)
}

func (r *PostgresRepository) GetPublishedByID(id uint) (*model.HeatingEquipment, error) {
	var item model.HeatingEquipment
	err := r.db.Preload("Likes").Where("id = ? AND status = ?", id, model.StatusPublished).First(&item).Error
	return normalizeNotFound(&item, err)
}

func (r *PostgresRepository) GetNextPublished(currentID uint) (*model.HeatingEquipment, error) {
	var item model.HeatingEquipment

	err := r.db.
		Preload("Likes").
		Where("status = ? AND id > ?", model.StatusPublished, currentID).
		Order("id ASC").
		First(&item).Error

	if err == nil {
		return &item, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return r.GetFirstPublished()
}

func (r *PostgresRepository) SetLike(userID, equipmentID uint, value int) (*model.HeatingEquipment, error) {
	if value != 0 && value != 1 {
		return nil, ErrInvalidLikeValue
	}
	if _, err := r.GetPublishedByID(equipmentID); err != nil {
		return nil, err
	}

	if value == 1 {
		like := model.HeatingEquipmentLike{UserID: userID, HeatingEquipmentID: equipmentID, CreatedAt: time.Now()}
		if err := r.db.Where("user_id = ? AND heating_equipment_id = ?", userID, equipmentID).
			FirstOrCreate(&like).Error; err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Where("user_id = ? AND heating_equipment_id = ?", userID, equipmentID).
			Delete(&model.HeatingEquipmentLike{}).Error; err != nil {
			return nil, err
		}
	}
	return r.GetPublishedByID(equipmentID)
}

func (r *PostgresRepository) SoftDelete(creatorID, equipmentID uint) error {
	now := time.Now()
	result := r.db.Model(&model.HeatingEquipment{}).
		Where("id = ? AND creator_id = ? AND status <> ?", equipmentID, creatorID, model.StatusDeleted).
		Updates(map[string]any{"status": model.StatusDeleted, "finished_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) RegisterUser(login, passwordHash string) (*model.User, error) {
	user := model.User{Login: login, PasswordHash: passwordHash, CreatedAt: time.Now()}
	if err := r.db.Create(&user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrLoginExists
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) putFile(ctx context.Context, objectName string, header *multipart.FileHeader) error {
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = r.minio.PutObject(ctx, r.bucket, objectName, file, header.Size, minio.PutObjectOptions{ContentType: header.Header.Get("Content-Type")})
	return err
}

func generatedObjectName(prefix, originalName string) (string, error) {
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(originalName))
	if len(ext) > 10 || strings.ContainsAny(ext, `/\\`) {
		ext = ""
	}
	return prefix + "_" + hex.EncodeToString(random) + ext, nil
}

func normalizeNotFound(item *model.HeatingEquipment, err error) (*model.HeatingEquipment, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresRepository) GetUserByLogin(login string) (*model.User, error) {
	var user model.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
