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

// HeatingEquipment — услуга по предметной области:
// отопительное оборудование.
type HeatingEquipment struct {
	ID uint `gorm:"primaryKey"`

	Name string `gorm:"column:equipment_name;type:varchar(120);not null"`

	Description *string `gorm:"column:short_description;type:varchar(500)"`

	Status StatusType `gorm:"column:status;type:varchar(20);not null;index"`

	ImageURL string `gorm:"column:image_url;type:varchar(500);not null"`
	VideoURL string `gorm:"column:video_url;type:varchar(500);not null"`

	// Два поля по предметной области
	Power *float64 `gorm:"column:power_kw"`
	Cost  *float64 `gorm:"column:cost_rub"`

	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	FormedAt  *time.Time `gorm:"column:formed_at"`

	CreatorID uint `gorm:"column:creator_id;not null;index"`

	Creator User `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	Likes []HeatingEquipmentLike `gorm:"foreignKey:HeatingEquipmentID"`
}

func (HeatingEquipment) TableName() string {
	return "heating_equipment"
}

func (h HeatingEquipment) LikesCount() int {
	return len(h.Likes)
}

type HeatingEquipmentLike struct {
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"column:user_id;not null;uniqueIndex:idx_user_heating_equipment"`

	HeatingEquipmentID uint `gorm:"column:heating_equipment_id;not null;uniqueIndex:idx_user_heating_equipment"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	HeatingEquipment HeatingEquipment `gorm:"foreignKey:HeatingEquipmentID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func (HeatingEquipmentLike) TableName() string {
	return "heating_equipment_likes"
}