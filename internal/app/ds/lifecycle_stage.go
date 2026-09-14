package ds

import "time"

type StageStatus string

const (
	StatusDraft     StageStatus = "черновик"
	StatusPublished StageStatus = "опубликован"
	StatusDeleted   StageStatus = "удален"
)

type LifecycleStage struct {
	ID            uint        `gorm:"primaryKey"`
	Title         string      `gorm:"column:title;type:varchar(150);not null"`
	Description   string      `gorm:"column:description;type:text"`
	Status        StageStatus `gorm:"column:status;type:varchar(20);not null"`
	ImageURL      string      `gorm:"column:image_url;type:varchar(255)"`
	VideoURL      string      `gorm:"column:video_url;type:varchar(255)"`
	PricePerUnit  int         `gorm:"column:price_per_unit;not null;default:0"`
	StagePosition int         `gorm:"column:stage_position;not null;default:0"`
	CreatedAt     time.Time   `gorm:"column:created_at;not null"`
	FormedAt      *time.Time  `gorm:"column:formed_at"`

	CreatorID uint     `gorm:"column:creator_id;not null"`
	Creator   Engineer `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (LifecycleStage) TableName() string {
	return "lifecycle_stages"
}
