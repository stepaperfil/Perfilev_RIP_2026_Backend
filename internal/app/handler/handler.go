package handler

import (
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"tco-backend/internal/app/repository"
)

const defaultEngineerID uint = 1

var (
	currentEngineerOnce sync.Once
	currentEngineerID   uint
)

// До авторизации (ЛР4) создатель зафиксирован: singleton всегда возвращает одного и того же инженера.
func CurrentEngineerID() uint {
	currentEngineerOnce.Do(func() {
		currentEngineerID = defaultEngineerID
	})
	return currentEngineerID
}

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, status int, err error) {
	logrus.Error(err)
	ctx.AbortWithStatus(status)
}

func (h *Handler) repositoryError(ctx *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	h.errorHandler(ctx, http.StatusInternalServerError, err)
}

func parseID(ctx *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	return uint(id), err
}
