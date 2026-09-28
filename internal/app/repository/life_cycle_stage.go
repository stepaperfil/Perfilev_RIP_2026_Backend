package repository

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tco-backend/internal/app/ds"
)

const lifeCycleStageOrder = "life_cycle_stages.stage_position ASC, life_cycle_stages.id ASC"

func (r *Repository) publishedWithStats(engineerID uint) *gorm.DB {
	return r.db.Model(&ds.LifeCycleStage{}).
		Select(`life_cycle_stages.*,
			(SELECT COUNT(*) FROM life_cycle_stage_likes l WHERE l.life_cycle_stage_id = life_cycle_stages.id) AS likes_count,
			EXISTS (SELECT 1 FROM life_cycle_stage_likes l WHERE l.life_cycle_stage_id = life_cycle_stages.id AND l.engineer_id = ?) AS is_liked`,
			engineerID).
		Where("life_cycle_stages.status = ?", string(ds.StatusPublished))
}

func scanOneLifeCycleStage(query *gorm.DB) (ds.LifeCycleStageWithStats, error) {
	var stage ds.LifeCycleStageWithStats
	res := query.Limit(1).Scan(&stage)
	if res.Error != nil {
		return stage, res.Error
	}
	if res.RowsAffected == 0 {
		return stage, gorm.ErrRecordNotFound
	}
	return stage, nil
}

func (r *Repository) GetPublishedLifeCycleStages(engineerID uint, maxPricePerUnit *int) ([]ds.LifeCycleStageWithStats, error) {
	query := r.publishedWithStats(engineerID)
	if maxPricePerUnit != nil {
		query = query.Where("life_cycle_stages.price_per_unit <= ?", *maxPricePerUnit)
	}

	var lifeCycleStages []ds.LifeCycleStageWithStats
	err := query.Order(lifeCycleStageOrder).Scan(&lifeCycleStages).Error
	return lifeCycleStages, err
}

func (r *Repository) GetFirstFeedLifeCycleStage(engineerID uint) (ds.LifeCycleStageWithStats, error) {
	return scanOneLifeCycleStage(r.publishedWithStats(engineerID).Order(lifeCycleStageOrder))
}

func (r *Repository) GetFeedLifeCycleStage(engineerID, id uint) (ds.LifeCycleStageWithStats, error) {
	return scanOneLifeCycleStage(r.publishedWithStats(engineerID).Where("life_cycle_stages.id = ?", id))
}

func (r *Repository) GetNextFeedLifeCycleStage(engineerID, id uint) (ds.LifeCycleStageWithStats, error) {
	return scanOneLifeCycleStage(r.publishedWithStats(engineerID).
		Where("(life_cycle_stages.stage_position, life_cycle_stages.id) > (SELECT s.stage_position, s.id FROM life_cycle_stages s WHERE s.id = ?)", id).
		Order(lifeCycleStageOrder))
}

func (r *Repository) GetDraftLifeCycleStage(engineerID uint) (ds.LifeCycleStage, error) {
	var lifeCycleStage ds.LifeCycleStage
	err := r.db.Where("creator_id = ? AND status = ?", engineerID, string(ds.StatusDraft)).First(&lifeCycleStage).Error
	return lifeCycleStage, err
}

func (r *Repository) CreateDraftLifeCycleStage(engineerID uint, title, imageKey, videoKey string) (ds.LifeCycleStage, error) {
	lifeCycleStage := ds.LifeCycleStage{
		Title:     title,
		Status:    ds.StatusDraft,
		ImageKey:  imageKey,
		VideoKey:  videoKey,
		CreatedAt: time.Now(),
		CreatorID: engineerID,
	}
	err := r.db.Create(&lifeCycleStage).Error
	return lifeCycleStage, err
}

func (r *Repository) SetLifeCycleStageLike(engineerID, id uint, like bool) error {
	var published int64
	if err := r.db.Model(&ds.LifeCycleStage{}).
		Where("id = ? AND status = ?", id, string(ds.StatusPublished)).
		Count(&published).Error; err != nil {
		return err
	}
	if published == 0 {
		return gorm.ErrRecordNotFound
	}

	if like {
		return r.db.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&ds.LifeCycleStageLike{EngineerID: engineerID, LifeCycleStageID: id}).Error
	}
	return r.db.Where("engineer_id = ? AND life_cycle_stage_id = ?", engineerID, id).
		Delete(&ds.LifeCycleStageLike{}).Error
}

func (r *Repository) PublishLifeCycleStage(engineerID, id uint, description string, pricePerUnit, stagePosition int) (ds.LifeCycleStage, error) {
	res := r.db.Model(&ds.LifeCycleStage{}).
		Where("id = ? AND creator_id = ? AND status = ?", id, engineerID, string(ds.StatusDraft)).
		Updates(map[string]interface{}{
			"description":    description,
			"price_per_unit": pricePerUnit,
			"stage_position": stagePosition,
			"status":         string(ds.StatusPublished),
			"formed_at":      time.Now(),
		})
	if res.Error != nil {
		return ds.LifeCycleStage{}, res.Error
	}
	if res.RowsAffected == 0 {
		return ds.LifeCycleStage{}, gorm.ErrRecordNotFound
	}

	var lifeCycleStage ds.LifeCycleStage
	err := r.db.First(&lifeCycleStage, id).Error
	return lifeCycleStage, err
}

func (r *Repository) DeleteLifeCycleStage(engineerID, id uint) error {
	res := r.db.Model(&ds.LifeCycleStage{}).
		Where("id = ? AND creator_id = ? AND status <> ?", id, engineerID, string(ds.StatusDeleted)).
		Update("status", string(ds.StatusDeleted))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
