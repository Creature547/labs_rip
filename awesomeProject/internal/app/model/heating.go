package model

import "time"

type StatusType string

const (
	StatusDraft     StatusType = "черновик"
	StatusPublished StatusType = "опубликован"
	StatusDeleted   StatusType = "удален"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Login    string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(100);not null"`
}

type HeatingCalculation struct {
	ID          uint       `gorm:"primaryKey"`
	Title       string     `gorm:"type:varchar(120);not null"`
	Description string     `gorm:"type:varchar(500)"`
	Status      StatusType `gorm:"type:varchar(20);not null;index"`
	ImageURL    string     `gorm:"type:varchar(500);not null"`
	VideoURL    string     `gorm:"type:varchar(500);not null"`

	// Параметры устройства и помещения из макета лабораторной.
	Power              float64 `gorm:"column:power_kw;not null;default:0"`
	Cost               float64 `gorm:"column:cost_rub;not null;default:0"`
	Area               float64 `gorm:"column:area_m2;not null;default:0"`
	CeilingHeight      float64 `gorm:"column:ceiling_height_m;not null;default:2.7"`
	FloorType          string  `gorm:"column:floor_type;type:varchar(30);not null;default:'1 этаж'"`
	WallMaterial       string  `gorm:"column:wall_material;type:varchar(80);not null;default:'Газобетонный блок'"`
	DesiredTemperature float64 `gorm:"column:desired_temperature_c;not null;default:22"`
	ElectricityTariff  float64 `gorm:"column:electricity_tariff_rub_kwh;not null;default:5.5"`

	CreatedAt time.Time `gorm:"not null"`
	FormedAt  *time.Time
	CreatorID uint `gorm:"not null;index"`

	Creator User          `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Likes   []HeatingLike `gorm:"foreignKey:HeatingCalculationID"`
}

func (HeatingCalculation) TableName() string { return "heating_calculations" }

func (h HeatingCalculation) LikesCount() int { return len(h.Likes) }

type HeatingLike struct {
	ID                   uint `gorm:"primaryKey"`
	UserID               uint `gorm:"not null;uniqueIndex:idx_user_heating"`
	HeatingCalculationID uint `gorm:"not null;uniqueIndex:idx_user_heating"`

	User               User               `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	HeatingCalculation HeatingCalculation `gorm:"foreignKey:HeatingCalculationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func (HeatingLike) TableName() string { return "heating_likes" }
