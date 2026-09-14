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
	GetAllPublished(minPower float64) []model.HeatingCalculation
}

type inMemoryHeatingRepository struct {
	items []model.HeatingCalculation
}

func NewInMemoryHeatingRepository() HeatingRepository {
	return &inMemoryHeatingRepository{
		items: []model.HeatingCalculation{
			{
				ID:          1,
				Title:       "Цилиндрическая печь",
				Description: "Современная цилиндрическая печь-камин сочетает стильный панорамный обзор огня и высокую отдачу тепла. Она идеально подходит для эффективного обогрева загородных домов и коттеджей среднего размера. Компактный корпус позволяет легко вписать её в любой современный интерьер.",
				ImageURL:    "http://127.0.0.1:9000/heating/%D0%A6%D0%B8%D0%BB%D0%B8%D0%BD%D0%B4%D1%80%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B0%D1%8F%20%D0%BF%D0%B5%D1%87%D1%8C.png",
				VideoURL:    "http://127.0.0.1:9000/heating/%D0%A6%D0%B8%D0%BB%D0%B8%D0%BD%D0%B4%D1%80%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B0%D1%8F%20%D0%BF%D0%B5%D1%87%D1%8C.mp4",
				Status:      model.StatusPublished,
				Power:       7.0,
				Cost:        85000,
				Likes:       []model.UserLike{{UserID: 101}, {UserID: 102}},
			},
			{
				ID:          2,
				Title:       "Бойлер (косвенного нагрева)",
				Description: "Высокопроизводительный бойлер косвенного нагрева обеспечивает стабильную подачу горячей воды для всей семьи. Прибор подключается к основному котлу отопления и поддерживает заданную температуру с минимальными энергозатратами. Надежная теплоизоляция долго сохраняет воду горячей.",
				ImageURL:    "http://127.0.0.1:9000/heating/%D0%91%D0%BE%D0%B9%D0%BB%D0%B5%D1%80.png",
				VideoURL:    "http://127.0.0.1:9000/heating/%D0%B1%D0%BE%D0%B9%D0%BB%D0%B5%D1%80%D1%8B.mp4",
				Status:      model.StatusPublished,
				Power:       24.0,
				Cost:        65000,
				Likes:       []model.UserLike{{UserID: 101}, {UserID: 103}, {UserID: 104}},
			},
			{
				ID:          3,
				Title:       "Электрическая печь (Электрокамин)",
				Description: "Электрический камин с реалистичным эффектом живого пламени создаст уютную атмосферу в доме без дыма и копоти. Прибор оснащен быстрым обогревом и удобной регулировкой интенсивности горения. Отлично подходит для безопасного использования в квартирах и жилых комнатах.",
				ImageURL:    "http://127.0.0.1:9000/heating/%D0%AD%D0%BB%D0%B5%D0%BA%D1%82%D1%80%D0%BE%D0%BA%D0%B0%D0%BC%D0%B8%D0%BD.png",
				VideoURL:    "http://127.0.0.1:9000/heating/%D1%8D%D0%BB%D0%B5%D0%BA%D1%82%D1%80%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B0%D1%8F%20%D0%BF%D0%B5%D1%87%D1%8C%20%D1%81%20%D0%B8%D0%B7%D0%BE%D0%B1%D1%80%D0%B0%D0%B6%D0%B5%D0%BD%D0%B8%D0%B5%D0%BC%20%D0%BA%D0%B0%D0%BC%D0%B8%D0%BD%D0%B0.mp4",
				Status:      model.StatusPublished,
				Power:       3.0,
				Cost:        35000,
				Likes:       []model.UserLike{{UserID: 102}},
			},
			{
				ID:          4,
				Title:       "Теплогенератор",
				Description: "Мощный промышленный теплогенератор предназначен для быстрого и эффективного обогрева больших площадей, складов и производственных цехов. Высокий КПД и прочный металлический корпус гарантируют надежную работу в жестких условиях. Работает с системами принудительной вентиляции и подачи воздуха.",
				ImageURL:    "http://127.0.0.1:9000/heating/%D0%A2%D0%B5%D0%BF%D0%BB%D0%BE%D0%B3%D0%B5%D0%BD%D0%B5%D1%80%D0%B0%D1%82%D0%BE%D1%80.png",
				VideoURL:    "http://127.0.0.1:9000/heating/%D0%9F%D1%80%D0%BE%D0%BC%D1%8B%D1%88%D0%BB%D0%B5%D0%BD%D0%BD%D1%8B%D0%B9%20%D1%82%D0%B5%D0%BF%D0%BB%D0%BE%D0%B3%D0%B5%D0%BD%D0%B5%D1%80%D0%B0%D1%82%D0%BE%D1%80.mp4",
				Status:      model.StatusPublished,
				Power:       200.0,
				Cost:        200000,
				Likes:       []model.UserLike{{UserID: 101}, {UserID: 106}, {UserID: 107}, {UserID: 108}, {UserID: 109}, {UserID: 110}, {UserID: 111}, {UserID: 112}, {UserID: 113}, {UserID: 114}},
			},
			{
				ID:          5,
				Title:       "Новое устройство (Черновик)",
				Description: "Предварительное описание для нового отопительного прибора.",
				ImageURL:    "",
				VideoURL:    "",
				Status:      model.StatusDraft,
				Power:       5.0,
				Cost:        40000,
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
	return nil, errors.New("устройство не найдено")
}

func (r *inMemoryHeatingRepository) GetFirstPublished() (*model.HeatingCalculation, error) {
	for _, item := range r.items {
		if item.Status == model.StatusPublished {
			return &item, nil
		}
	}
	return nil, errors.New("нет опубликованных устройств")
}

func (r *inMemoryHeatingRepository) GetNextPublished(currentID int) (*model.HeatingCalculation, error) {
	var published []model.HeatingCalculation
	for _, item := range r.items {
		if item.Status == model.StatusPublished {
			published = append(published, item)
		}
	}

	if len(published) == 0 {
		return nil, errors.New("нет устройств")
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

func (r *inMemoryHeatingRepository) GetAllPublished(minPower float64) []model.HeatingCalculation {
	var result []model.HeatingCalculation
	for _, item := range r.items {
		if item.Status == model.StatusPublished {
			if minPower == 0 || item.Power >= minPower {
				result = append(result, item)
			}
		}
	}
	return result
}