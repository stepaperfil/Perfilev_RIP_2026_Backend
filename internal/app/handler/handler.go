package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"tco-backend/internal/app/config"
	"tco-backend/internal/app/redis"
	"tco-backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	Config     *config.Config
}

func NewHandler(r *repository.Repository, redisClient *redis.Client, cfg *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redisClient,
		Config:     cfg,
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
