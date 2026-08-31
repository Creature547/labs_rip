package repository

import (
	"awesomeProject/internal/app/model"
	"errors"
)

type HeatingRepository interface {
	GetByID(id int) (*model.HeatingCalculation, error)
	GetNextPublished(currentID int) (*model.HeatingCalculation, error)
	GetFirstPublished() (*model.HeatingCalculation, error)
	GetDraft() (*model.HeatingCalculation, error)
	GetAllPublished(minArea int) []model.HeatingCalculation
}

type inMemoryHeatingRepository struct {
	items []model.HeatingCalculation
}

func NewInMemoryHeatingRepository() HeatingRepository {
	return &inMemoryHeatingRepository{
		items: []model.HeatingCalculation{
			{
				ID:          1,
				Title:       "Коттедж в Подмосковье",
				Description: "Расчет системы отопления для двухэтажного дома с электрическим котлом и водяным теплым полом на первом этаже.",
				ImageURL:    "http://localhost:9000/heating-bucket/cottage.jpg",
				VideoURL:    "http://localhost:9000/heating-bucket/cottage.mp4",
				Status:      model.StatusPublished,
				HouseArea:   250,
				FuelType:    "Электричество",
				AvgTemp:     -5.5,
				TariffRate:  7.5,
				SeasonDays:  210,
				SeasonCost:  85000,
				Likes:       []model.UserLike{{UserID: 101}, {UserID: 102}},
			},
			{
				ID:          2,
				Title:       "Загородный дом (Газ)",
				Description: "Оптимальный расчет при подключении магистрального газа. Высокий КПД и минимальные эксплуатационные расходы.",
				ImageURL:    "http://localhost:9000/heating-bucket/gas_house.jpg",
				VideoURL:    "http://localhost:9000/heating-bucket/gas_house.mp4",
				Status:      model.StatusPublished,
				HouseArea:   180,
				FuelType:    "Магистральный газ",
				AvgTemp:     -4.0,
				TariffRate:  6.8,
				SeasonDays:  180,
				SeasonCost:  32000,
				Likes:       []model.UserLike{{UserID: 101}, {UserID: 103}, {UserID: 104}, {UserID: 105}},
			},
			{
				ID:          3,
				Title:       "Дачный дом (Пеллеты)",
				Description: "Автономное отопление твердотопливным пеллетным котлом для регионов без газификации.",
				ImageURL:    "http://localhost:9000/heating-bucket/dacha.jpg",
				VideoURL:    "http://localhost:9000/heating-bucket/dacha.mp4",
				Status:      model.StatusPublished,
				HouseArea:   100,
				FuelType:    "Пеллеты",
				AvgTemp:     -6.0,
				TariffRate:  9.0,
				SeasonDays:  150,
				SeasonCost:  45000,
				Likes:       []model.UserLike{{UserID: 102}},
			},
			{
				ID:          4,
				Title:       "Таунхаус (Тепловой насос)",
				Description: "Современная геотермальная система отопления с минимальным энергопотреблением.",
				ImageURL:    "http://localhost:9000/heating-bucket/townhouse.jpg",
				VideoURL:    "http://localhost:9000/heating-bucket/townhouse.mp4",
				Status:      model.StatusPublished,
				HouseArea:   140,
				FuelType:    "Тепловой насос",
				AvgTemp:     -3.0,
				TariffRate:  5.2,
				SeasonDays:  200,
				SeasonCost:  28000,
				Likes:       []model.UserLike{{UserID: 101}, {UserID: 106}, {UserID: 107}},
			},
			{
				ID:          5,
				Title:       "Новый черновик расчета",
				Description: "Предварительные параметры для строящегося объекта.",
				ImageURL:    "http://localhost:9000/heating-bucket/draft.jpg",
				VideoURL:    "http://localhost:9000/heating-bucket/draft.mp4",
				Status:      model.StatusDraft, // Единственный черновик
				HouseArea:   120,
				FuelType:    "Дрова / Уголь",
				AvgTemp:     -5.0,
				TariffRate:  4.5,
				SeasonDays:  180,
				SeasonCost:  30000,
				Likes:       []model.UserLike{},
			},
			{
				ID:          6,
				Title:       "Удаленный объект",
				Description: "Архивная запись.",
				ImageURL:    "",
				VideoURL:    "",
				Status:      model.StatusDeleted, // Не отображается в интерфейсе
				HouseArea:   300,
				FuelType:    "Дизель",
				AvgTemp:     -10.0,
				TariffRate:  12.0,
				SeasonDays:  210,
				SeasonCost:  150000,
				Likes:       []model.UserLike{},
			},
		},
	}
}

func (r *inMemoryHeatingRepository) GetByID(id int) (*model.HeatingCalculation, error) {
	for _, item := range r.items {
		if item.ID == id && item.Status == model.StatusPublished {
			return &item, nil
		}
	}
	return nil, errors.New("расчет не найден")
}

func (r *inMemoryHeatingRepository) GetFirstPublished() (*model.HeatingCalculation, error) {
	for _, item := range r.items {
		if item.Status == model.StatusPublished {
			return &item, nil
		}
	}
	return nil, errors.New("нет опубликованных расчетов")
}

func (r *inMemoryHeatingRepository) GetNextPublished(currentID int) (*model.HeatingCalculation, error) {
	var published []model.HeatingCalculation
	for _, item := range r.items {
		if item.Status == model.StatusPublished {
			published = append(published, item)
		}
	}

	if len(published) == 0 {
		return nil, errors.New("нет расчетов")
	}

	for i, item := range published {
		if item.ID == currentID {
			nextIndex := (i + 1) % len(published)
			return &published[nextIndex], nil
		}
	}

	return &published[0], nil
}

func (r *inMemoryHeatingRepository) GetDraft() (*model.HeatingCalculation, error) {
	for _, item := range r.items {
		if item.Status == model.StatusDraft {
			return &item, nil
		}
	}
	return nil, errors.New("черновик не найден")
}

func (r *inMemoryHeatingRepository) GetAllPublished(minArea int) []model.HeatingCalculation {
	var result []model.HeatingCalculation
	for _, item := range r.items {
		if item.Status == model.StatusPublished {
			if minArea == 0 || item.HouseArea >= minArea {
				result = append(result, item)
			}
		}
	}
	return result
}
