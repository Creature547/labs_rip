package serializer

import (
	"awesomeProject/internal/app/model"
	"encoding/json"
	"testing"
)

func TestHeatingEquipmentSerializerAlwaysHasSameFields(t *testing.T) {
	item := model.HeatingEquipment{ID: 7, Name: "Boiler", CreatorID: 1, ImageName: "image.png", VideoName: "video.mp4"}
	response := HeatingEquipmentFromModel(item, 1, "http://localhost:9000", "heating-media")

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	expected := []string{"id", "name", "description", "power", "cost", "image_url", "video_url", "is_creator", "liked", "likes_count"}
	for _, key := range expected {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing JSON field %q", key)
		}
	}
	if len(decoded) != len(expected) {
		t.Fatalf("unexpected JSON fields: %#v", decoded)
	}
	if decoded["description"] != "" {
		t.Fatalf("empty description must still be present, got %#v", decoded["description"])
	}
}

func TestHeatingEquipmentSerializerFlags(t *testing.T) {
	item := model.HeatingEquipment{
		ID:        1,
		Name:      "Heater",
		CreatorID: 1,
		Likes:     []model.HeatingEquipmentLike{{UserID: 1}},
	}
	response := HeatingEquipmentFromModel(item, 1, "http://localhost:9000", "heating-media")
	if response.IsCreator != 1 || response.Liked != 1 || response.LikesCount != 1 {
		t.Fatalf("unexpected flags: %+v", response)
	}
}
