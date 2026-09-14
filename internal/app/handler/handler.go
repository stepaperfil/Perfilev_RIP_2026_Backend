package handler

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"tco-backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func minioURL(key string) string {
	if key == "" {
		return ""
	}
	base := os.Getenv("MINIO_PUBLIC_URL")
	if base == "" {
		base = "http://localhost:9000/tco-assets"
	}
	return base + "/" + key
}

type tileView struct {
	Stage      repository.LifecycleStage
	ImageURL   string
	LikesCount int
}

func (h *Handler) GetLifecycleStages(ctx *gin.Context) {
	var stages []repository.LifecycleStage
	var err error

	maxPriceQuery := ctx.Query("maxPrice")
	if maxPriceQuery == "" {
		stages, err = h.Repository.GetPublishedLifecycleStages()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		maxPrice, convErr := strconv.Atoi(maxPriceQuery)
		if convErr != nil {
			logrus.Error(convErr)
			stages, err = h.Repository.GetPublishedLifecycleStages()
		} else {
			stages, err = h.Repository.GetLifecycleStagesByMaxPrice(maxPrice)
		}
		if err != nil {
			logrus.Error(err)
		}
	}

	views := make([]tileView, 0, len(stages))
	for _, stage := range stages {
		views = append(views, tileView{
			Stage:      stage,
			ImageURL:   minioURL(stage.ImageKey),
			LikesCount: len(stage.Likes),
		})
	}

	ctx.HTML(http.StatusOK, "tiles.html", gin.H{
		"stages":   views,
		"maxPrice": maxPriceQuery,
	})
}

func (h *Handler) GetLifecycleStage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "некорректный id этапа")
		return
	}

	var stage repository.LifecycleStage
	if ctx.Query("next") == "true" {
		stage, err = h.Repository.GetNextLifecycleStage(id)
	} else {
		stage, err = h.Repository.GetLifecycleStage(id)
	}
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "этап не найден")
		return
	}

	if stage.Status != "published" {
		ctx.String(http.StatusNotFound, "этап не найден")
		return
	}

	ctx.HTML(http.StatusOK, "stage.html", gin.H{
		"stage":      stage,
		"imageURL":   minioURL(stage.ImageKey),
		"videoURL":   minioURL(stage.VideoKey),
		"likesCount": len(stage.Likes),
	})
}

func (h *Handler) GetDraftLifecycleStage(ctx *gin.Context) {
	stage, err := h.Repository.GetDraftLifecycleStage()
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "в коллекции нет черновика")
		return
	}

	ctx.HTML(http.StatusOK, "draft.html", gin.H{
		"stage":    stage,
		"imageURL": minioURL(stage.ImageKey),
		"videoURL": minioURL(stage.VideoKey),
	})
}
