package repository

import (
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tco-backend/internal/app/ds"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) GetPublishedLifeCycleStages(maxPricePerUnit int, hasMaxPricePerUnit bool) ([]ds.LifeCycleStage, error) {
	query := r.db.Where("status = ?", string(ds.StatusPublished))

	if hasMaxPricePerUnit {
		query = query.Where("price_per_unit <= ?", maxPricePerUnit)
	}

	var lifeCycleStages []ds.LifeCycleStage
	err := query.Order("stage_position ASC, id ASC").Find(&lifeCycleStages).Error
	return lifeCycleStages, err
}

func (r *Repository) GetLifeCycleStage(id uint) (ds.LifeCycleStage, error) {
	var lifeCycleStage ds.LifeCycleStage
	err := r.db.Where("id = ? AND status <> ?", id, string(ds.StatusDeleted)).First(&lifeCycleStage).Error
	return lifeCycleStage, err
}

func (r *Repository) GetNextLifeCycleStage(id uint) (ds.LifeCycleStage, error) {
	lifeCycleStages, err := r.GetPublishedLifeCycleStages(0, false)
	if err != nil {
		return ds.LifeCycleStage{}, err
	}
	for i, s := range lifeCycleStages {
		if s.ID == id && i+1 < len(lifeCycleStages) {
			return lifeCycleStages[i+1], nil
		}
	}
	return ds.LifeCycleStage{}, gorm.ErrRecordNotFound
}

func (r *Repository) GetDraftLifeCycleStage(engineerID uint) (ds.LifeCycleStage, bool, error) {
	var lifeCycleStage ds.LifeCycleStage
	err := r.db.Where("creator_id = ? AND status = ?", engineerID, string(ds.StatusDraft)).First(&lifeCycleStage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.LifeCycleStage{}, false, nil
		}
		return ds.LifeCycleStage{}, false, err
	}
	return lifeCycleStage, true, nil
}

func (r *Repository) CreateOrGetDraftLifeCycleStage(engineerID uint, title, imageURL, videoURL string) (ds.LifeCycleStage, error) {
	existing, found, err := r.GetDraftLifeCycleStage(engineerID)
	if err != nil {
		return ds.LifeCycleStage{}, err
	}
	if found {
		return existing, nil
	}

	lifeCycleStage := ds.LifeCycleStage{
		Title:     title,
		Status:    ds.StatusDraft,
		ImageURL:  imageURL,
		VideoURL:  videoURL,
		CreatedAt: time.Now(),
		CreatorID: engineerID,
	}
	if err := r.db.Create(&lifeCycleStage).Error; err != nil {
		return ds.LifeCycleStage{}, err
	}
	return lifeCycleStage, nil
}

func (r *Repository) PublishLifeCycleStage(id uint, description string, pricePerUnit, stagePosition int) error {
	now := time.Now()
	return r.db.Model(&ds.LifeCycleStage{}).
		Where("id = ? AND status = ?", id, string(ds.StatusDraft)).
		Updates(map[string]interface{}{
			"description":    description,
			"price_per_unit": pricePerUnit,
			"stage_position": stagePosition,
			"status":         string(ds.StatusPublished),
			"formed_at":      now,
		}).Error
}

func (r *Repository) GetLifeCycleStageLikesCount(lifeCycleStageID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.LifeCycleStageLike{}).Where("life_cycle_stage_id = ?", lifeCycleStageID).Count(&count).Error
	return count, err
}

func (r *Repository) DeleteLifeCycleStageRawSQL(id uint) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	rows, err := sqlDB.Query(
		`UPDATE life_cycle_stages SET status = $1 WHERE id = $2 AND status <> $1 RETURNING id`,
		string(ds.StatusDeleted), id,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	updated := false
	for rows.Next() {
		updated = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !updated {
		return gorm.ErrRecordNotFound
	}
	return nil
}
