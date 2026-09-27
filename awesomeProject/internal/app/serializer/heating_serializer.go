package serializer

import (
	"awesomeProject/internal/app/model"
	"strings"
)

type HeatingEquipment struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Power       float64 `json:"power"`
	Cost        float64 `json:"cost"`
	ImageURL    string  `json:"image_url"`
	VideoURL    string  `json:"video_url"`
	IsCreator   int     `json:"is_creator"`
	Liked       int     `json:"liked"`
	LikesCount  int     `json:"likes_count"`
}

func HeatingEquipmentFromModel(item model.HeatingEquipment, currentUserID uint, mediaBaseURL, bucket string) HeatingEquipment {
	resp := HeatingEquipment{
		ID:         item.ID,
		Name:       item.Name,
		IsCreator:  boolToInt(item.CreatorID == currentUserID),
		Liked:      boolToInt(item.IsLikedBy(currentUserID)),
		LikesCount: item.LikesCount(),
		ImageURL:   mediaURL(mediaBaseURL, bucket, item.ImageName),
		VideoURL:   mediaURL(mediaBaseURL, bucket, item.VideoName),
	}
	if item.Description != nil {
		resp.Description = *item.Description
	}
	if item.Power != nil {
		resp.Power = *item.Power
	}
	if item.Cost != nil {
		resp.Cost = *item.Cost
	}
	return resp
}

func HeatingEquipmentList(items []model.HeatingEquipment, currentUserID uint, mediaBaseURL, bucket string) []HeatingEquipment {
	result := make([]HeatingEquipment, 0, len(items))
	for _, item := range items {
		result = append(result, HeatingEquipmentFromModel(item, currentUserID, mediaBaseURL, bucket))
	}
	return result
}

func mediaURL(baseURL, bucket, name string) string {
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
		return name
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.Trim(bucket, "/") + "/" + strings.TrimLeft(name, "/")
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
