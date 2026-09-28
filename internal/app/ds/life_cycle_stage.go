package ds

import "time"

type LifeCycleStageStatus string

const (
	StatusDraft     LifeCycleStageStatus = "черновик"
	StatusPublished LifeCycleStageStatus = "опубликован"
	StatusDeleted   LifeCycleStageStatus = "удален"
)

type LifeCycleStage struct {
	ID            uint                 `gorm:"primaryKey"`
	Title         string               `gorm:"column:title;type:varchar(150);not null"`
	Description   string               `gorm:"column:description;type:text"`
	Status        LifeCycleStageStatus `gorm:"column:status;type:varchar(20);not null"`
	ImageKey      string               `gorm:"column:image_key;type:varchar(255)"`
	VideoKey      string               `gorm:"column:video_key;type:varchar(255)"`
	PricePerUnit  int                  `gorm:"column:price_per_unit;not null;default:0"`
	StagePosition int                  `gorm:"column:stage_position;not null;default:0"`
	CreatedAt     time.Time            `gorm:"column:created_at;not null"`
	FormedAt      *time.Time           `gorm:"column:formed_at"`

	CreatorID uint     `gorm:"column:creator_id;not null"`
	Creator   Engineer `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (LifeCycleStage) TableName() string {
	return "life_cycle_stages"
}

type LifeCycleStageWithStats struct {
	LifeCycleStage
	LikesCount int64
	IsLiked    bool
}
