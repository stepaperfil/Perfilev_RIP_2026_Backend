package ds

type LifeCycleStageLike struct {
	ID uint `gorm:"primaryKey"`

	EngineerID       uint `gorm:"column:engineer_id;not null;uniqueIndex:ux_engineer_life_cycle_stage"`
	LifeCycleStageID uint `gorm:"column:life_cycle_stage_id;not null;uniqueIndex:ux_engineer_life_cycle_stage"`

	Engineer       Engineer       `gorm:"foreignKey:EngineerID;references:ID;constraint:OnDelete:RESTRICT"`
	LifeCycleStage LifeCycleStage `gorm:"foreignKey:LifeCycleStageID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (LifeCycleStageLike) TableName() string {
	return "life_cycle_stage_likes"
}
