package model

type StatusType string

const (
	StatusDraft     StatusType = "draft"
	StatusPublished StatusType = "published"
	StatusDeleted   StatusType = "deleted"
)

type UserLike struct {
	UserID int `json:"user_id"`
}

type HeatingCalculation struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	VideoURL    string     `json:"video_url"`
	Status      StatusType `json:"status"`

	// 2 ключевых параметра устройства
	Power       float64    `json:"power"` // Мощность (кВт)
	Cost        float64    `json:"cost"`  // Стоимость (₽)

	Likes       []UserLike `json:"likes"`
}

func (h HeatingCalculation) LikesCount() int {
	return len(h.Likes)
}