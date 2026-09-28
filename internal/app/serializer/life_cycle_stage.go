package serializer

import (
	"tco-backend/internal/app/ds"
	"tco-backend/internal/app/storage"
)

type LifeCycleStageListItem struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	ImageURL     string `json:"image_url"`
	PricePerUnit int    `json:"price_per_unit"`
	LikesCount   int64  `json:"likes_count"`
	IsMine       int    `json:"is_mine"`
}

type LifeCycleStageFeedItem struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
	VideoURL      string `json:"video_url"`
	PricePerUnit  int    `json:"price_per_unit"`
	StagePosition int    `json:"stage_position"`
	LikesCount    int64  `json:"likes_count"`
	IsLiked       int    `json:"is_liked"`
}

type LifeCycleStage struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
	VideoURL      string `json:"video_url"`
	PricePerUnit  int    `json:"price_per_unit"`
	StagePosition int    `json:"stage_position"`
}

type LikeLifeCycleStageRequest struct {
	Like *int `json:"like" binding:"required,oneof=0 1"`
}

type PublishLifeCycleStageRequest struct {
	Description   string `json:"description" binding:"required"`
	PricePerUnit  *int   `json:"price_per_unit" binding:"required,min=0"`
	StagePosition *int   `json:"stage_position" binding:"required,min=1"`
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func NewLifeCycleStageList(stages []ds.LifeCycleStageWithStats, engineerID uint) []LifeCycleStageListItem {
	items := make([]LifeCycleStageListItem, 0, len(stages))
	for _, s := range stages {
		items = append(items, LifeCycleStageListItem{
			ID:           s.ID,
			Title:        s.Title,
			ImageURL:     storage.ObjectURL(s.ImageKey),
			PricePerUnit: s.PricePerUnit,
			LikesCount:   s.LikesCount,
			IsMine:       boolToInt(s.CreatorID == engineerID),
		})
	}
	return items
}

func NewLifeCycleStageFeedItem(s ds.LifeCycleStageWithStats) LifeCycleStageFeedItem {
	return LifeCycleStageFeedItem{
		ID:            s.ID,
		Title:         s.Title,
		Description:   s.Description,
		ImageURL:      storage.ObjectURL(s.ImageKey),
		VideoURL:      storage.ObjectURL(s.VideoKey),
		PricePerUnit:  s.PricePerUnit,
		StagePosition: s.StagePosition,
		LikesCount:    s.LikesCount,
		IsLiked:       boolToInt(s.IsLiked),
	}
}

func NewLifeCycleStage(s ds.LifeCycleStage) LifeCycleStage {
	return LifeCycleStage{
		ID:            s.ID,
		Title:         s.Title,
		Description:   s.Description,
		ImageURL:      storage.ObjectURL(s.ImageKey),
		VideoURL:      storage.ObjectURL(s.VideoKey),
		PricePerUnit:  s.PricePerUnit,
		StagePosition: s.StagePosition,
	}
}
