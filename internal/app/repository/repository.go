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

func (r *Repository) GetPublishedStages(maxPrice int, hasMaxPrice bool) ([]ds.LifecycleStage, error) {
	query := r.db.Where("status = ?", string(ds.StatusPublished))

	if hasMaxPrice {
		query = query.Where("price_per_unit <= ?", maxPrice)
	}

	var stages []ds.LifecycleStage
	err := query.Order("stage_position ASC, id ASC").Find(&stages).Error
	return stages, err
}

func (r *Repository) GetStage(id uint) (ds.LifecycleStage, error) {
	var stage ds.LifecycleStage
	err := r.db.Where("id = ? AND status <> ?", id, string(ds.StatusDeleted)).First(&stage).Error
	return stage, err
}

func (r *Repository) GetNextStage(id uint) (ds.LifecycleStage, error) {
	stages, err := r.GetPublishedStages(0, false)
	if err != nil {
		return ds.LifecycleStage{}, err
	}
	for i, s := range stages {
		if s.ID == id && i+1 < len(stages) {
			return stages[i+1], nil
		}
	}
	return ds.LifecycleStage{}, gorm.ErrRecordNotFound
}

func (r *Repository) GetDraftStage(engineerID uint) (ds.LifecycleStage, bool, error) {
	var stage ds.LifecycleStage
	err := r.db.Where("creator_id = ? AND status = ?", engineerID, string(ds.StatusDraft)).First(&stage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.LifecycleStage{}, false, nil
		}
		return ds.LifecycleStage{}, false, err
	}
	return stage, true, nil
}

func (r *Repository) CreateOrGetDraftStage(engineerID uint, title, imageURL, videoURL string) (ds.LifecycleStage, error) {
	existing, found, err := r.GetDraftStage(engineerID)
	if err != nil {
		return ds.LifecycleStage{}, err
	}
	if found {
		return existing, nil
	}

	stage := ds.LifecycleStage{
		Title:     title,
		Status:    ds.StatusDraft,
		ImageURL:  imageURL,
		VideoURL:  videoURL,
		CreatedAt: time.Now(),
		CreatorID: engineerID,
	}
	if err := r.db.Create(&stage).Error; err != nil {
		return ds.LifecycleStage{}, err
	}
	return stage, nil
}

func (r *Repository) PublishStage(id uint, description string, pricePerUnit, stagePosition int) error {
	now := time.Now()
	return r.db.Model(&ds.LifecycleStage{}).
		Where("id = ? AND status = ?", id, string(ds.StatusDraft)).
		Updates(map[string]interface{}{
			"description":    description,
			"price_per_unit": pricePerUnit,
			"stage_position": stagePosition,
			"status":         string(ds.StatusPublished),
			"formed_at":      now,
		}).Error
}

func (r *Repository) GetLikesCount(stageID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.StageLike{}).Where("stage_id = ?", stageID).Count(&count).Error
	return count, err
}

func (r *Repository) DeleteStageRawSQL(id uint) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	rows, err := sqlDB.Query(
		`UPDATE lifecycle_stages SET status = $1 WHERE id = $2 AND status <> $1 RETURNING id`,
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
