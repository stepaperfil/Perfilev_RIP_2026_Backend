package repository

import "tco-backend/internal/app/ds"

func (r *Repository) EngineerLoginExists(login string) (bool, error) {
	var count int64
	err := r.db.Model(&ds.Engineer{}).Where("login = ?", login).Count(&count).Error
	return count > 0, err
}

func (r *Repository) GetEngineerByLogin(login string) (ds.Engineer, error) {
	var engineer ds.Engineer
	err := r.db.Where("login = ?", login).First(&engineer).Error
	return engineer, err
}

func (r *Repository) CreateEngineer(fullName, login, passwordHash string) (ds.Engineer, error) {
	engineer := ds.Engineer{
		FullName:     fullName,
		Login:        login,
		PasswordHash: passwordHash,
	}
	err := r.db.Create(&engineer).Error
	return engineer, err
}
