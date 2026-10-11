package model

import "time"

type StatusType string

const (
	StatusDraft     StatusType = "черновик"
	StatusPublished StatusType = "опубликован"
	StatusDeleted   StatusType = "удален"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"-"`
	Login        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"-"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(100);not null" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at;not null" json:"-"`
}

func (User) TableName() string { return "users" }

// HeatingEquipment is the service entity for the subject area.
// System fields are intentionally excluded from JSON; the API uses serializers.
type HeatingEquipment struct {
	ID uint `gorm:"primaryKey" json:"-"`

	Name        string   `gorm:"column:equipment_name;type:varchar(120);not null" json:"-"`
	Description *string  `gorm:"column:short_description;type:varchar(500)" json:"-"`
	Power       *float64 `gorm:"column:power_kw" json:"-"`
	Cost        *float64 `gorm:"column:cost_rub" json:"-"`

	ImageName string `gorm:"column:image_name;type:varchar(255);not null" json:"-"`
	VideoName string `gorm:"column:video_name;type:varchar(255);not null" json:"-"`

	Status StatusType `gorm:"column:status;type:varchar(20);not null;index" json:"-"`

	CreatorID   uint  `gorm:"column:creator_id;not null;index" json:"-"`
	ModeratorID *uint `gorm:"column:moderator_id;index" json:"-"`

	CreatedAt  time.Time  `gorm:"column:created_at;not null" json:"-"`
	FormedAt   *time.Time `gorm:"column:formed_at" json:"-"`
	FinishedAt *time.Time `gorm:"column:finished_at" json:"-"`

	Creator   User  `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Moderator *User `gorm:"foreignKey:ModeratorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`

	Likes []HeatingEquipmentLike `gorm:"foreignKey:HeatingEquipmentID" json:"-"`
}

func (HeatingEquipment) TableName() string { return "heating_equipment" }

func (h HeatingEquipment) LikesCount() int { return len(h.Likes) }

func (h HeatingEquipment) IsLikedBy(userID uint) bool {
	for _, like := range h.Likes {
		if like.UserID == userID {
			return true
		}
	}
	return false
}

type HeatingEquipmentLike struct {
	ID uint `gorm:"primaryKey" json:"-"`

	UserID             uint `gorm:"column:user_id;not null;uniqueIndex:idx_user_heating_equipment" json:"-"`
	HeatingEquipmentID uint `gorm:"column:heating_equipment_id;not null;uniqueIndex:idx_user_heating_equipment" json:"-"`

	CreatedAt time.Time `gorm:"column:created_at;not null" json:"-"`

	User             User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	HeatingEquipment HeatingEquipment `gorm:"foreignKey:HeatingEquipmentID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}

func (HeatingEquipmentLike) TableName() string { return "heating_equipment_likes" }
