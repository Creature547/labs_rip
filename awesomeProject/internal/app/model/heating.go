package model


type StatusType string

const (
	StatusDraft     StatusType = "draft"
	StatusPublished StatusType = "published"
	StatusDeleted   StatusType = "deleted"
)

// UserLike представляет лайк от пользователя
type UserLike struct {
	UserID int `json:"user_id"`
}

// HeatingCalculation - анемичная модель услуги по 1НФ
type HeatingCalculation struct {
	ID           int        `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ImageURL     string     `json:"image_url"`
	VideoURL     string     `json:"video_url"`
	Status       StatusType `json:"status"`

	// 6 атомарных полей предметной области
	HouseArea    int        `json:"house_area"`     // Площадь дома (м²)
	FuelType     string     `json:"fuel_type"`      // Тип отопления
	AvgTemp      float64    `json:"avg_temp"`       // Средняя температура (°C)
	TariffRate   float64    `json:"tariff_rate"`    // Тариф (руб/кВт·ч или руб/м³)
	SeasonDays   int        `json:"season_days"`    // Дней в сезоне
	SeasonCost   int        `json:"season_cost"`    // Итоговая стоимость (руб)

	Likes        []UserLike `json:"likes"`
}

// LikesCount считает количество лайков по коллекции (для контроллера)
func (h HeatingCalculation) LikesCount() int {
	return len(h.Likes)
}
