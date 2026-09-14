package ds

type StageLike struct {
	ID uint `gorm:"primaryKey"`

	EngineerID uint `gorm:"column:engineer_id;not null;uniqueIndex:ux_engineer_stage"`
	StageID    uint `gorm:"column:stage_id;not null;uniqueIndex:ux_engineer_stage"`

	Engineer Engineer       `gorm:"foreignKey:EngineerID;references:ID;constraint:OnDelete:RESTRICT"`
	Stage    LifecycleStage `gorm:"foreignKey:StageID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (StageLike) TableName() string {
	return "stage_likes"
}
