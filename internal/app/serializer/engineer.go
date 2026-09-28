package serializer

import "tco-backend/internal/app/ds"

type RegisterEngineerRequest struct {
	FullName string `json:"full_name" binding:"required,max=150"`
	Login    string `json:"login" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type LoginEngineerRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Engineer struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Login    string `json:"login"`
}

func NewEngineer(e ds.Engineer) Engineer {
	return Engineer{
		ID:       e.ID,
		FullName: e.FullName,
		Login:    e.Login,
	}
}
