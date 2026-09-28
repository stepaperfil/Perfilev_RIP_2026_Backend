package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"tco-backend/internal/app/serializer"
)

func (h *Handler) RegisterEngineer(ctx *gin.Context) {
	var req serializer.RegisterEngineerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	exists, err := h.Repository.EngineerLoginExists(req.Login)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if exists {
		h.errorHandler(ctx, http.StatusConflict, errors.New("логин уже занят"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	engineer, err := h.Repository.CreateEngineer(req.FullName, req.Login, string(passwordHash))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, serializer.NewEngineer(engineer))
}

// Заглушка до ЛР4: проверка логина и пароля появится вместе с авторизацией.
func (h *Handler) LoginEngineer(ctx *gin.Context) {
	var req serializer.LoginEngineerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// Заглушка до ЛР4: завершение сессии появится вместе с авторизацией.
func (h *Handler) LogoutEngineer(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}
