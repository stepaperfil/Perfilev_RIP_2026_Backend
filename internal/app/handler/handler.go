package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"tco-backend/internal/app/ds"
	"tco-backend/internal/app/repository"
)

const currentEngineerID uint = 1

const (
	defaultImageURL = "/static/img/default-stage.jpg"
	defaultVideoURL = "/static/video/default-stage.mp4"
)

func withDefaults(imageURL, videoURL string) (string, string) {
	if imageURL == "" {
		imageURL = defaultImageURL
	}
	if videoURL == "" {
		videoURL = defaultVideoURL
	}
	return imageURL, videoURL
}

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

type tileView struct {
	Stage      ds.LifecycleStage
	ImageURL   string
	LikesCount int64
}

func (h *Handler) GetLifecycleStages(ctx *gin.Context) {
	maxPriceQuery := ctx.Query("maxPrice")
	hasMaxPrice := false
	maxPrice := 0
	if maxPriceQuery != "" {
		v, err := strconv.Atoi(maxPriceQuery)
		if err != nil {
			logrus.Error(err)
		} else {
			maxPrice = v
			hasMaxPrice = true
		}
	}

	stages, err := h.Repository.GetPublishedStages(maxPrice, hasMaxPrice)
	if err != nil {
		logrus.Error(err)
	}

	views := make([]tileView, 0, len(stages))
	for _, stage := range stages {
		imageURL, _ := withDefaults(stage.ImageURL, stage.VideoURL)

		likes, err := h.Repository.GetLikesCount(stage.ID)
		if err != nil {
			logrus.Error(err)
		}

		views = append(views, tileView{
			Stage:      stage,
			ImageURL:   imageURL,
			LikesCount: likes,
		})
	}

	ctx.HTML(http.StatusOK, "tiles.html", gin.H{
		"stages":   views,
		"maxPrice": maxPriceQuery,
	})
}

func (h *Handler) GetLifecycleStage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "некорректный id этапа")
		return
	}
	id := uint(id64)

	var stage ds.LifecycleStage
	if ctx.Query("next") == "true" {
		stage, err = h.Repository.GetNextStage(id)
	} else {
		stage, err = h.Repository.GetStage(id)
	}
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "этап не найден")
		return
	}

	likes, err := h.Repository.GetLikesCount(stage.ID)
	if err != nil {
		logrus.Error(err)
	}

	imageURL, videoURL := withDefaults(stage.ImageURL, stage.VideoURL)

	ctx.HTML(http.StatusOK, "stage.html", gin.H{
		"stage":      stage,
		"imageURL":   imageURL,
		"videoURL":   videoURL,
		"likesCount": likes,
	})
}

func (h *Handler) GetDraftLifecycleStage(ctx *gin.Context) {
	stage, found, err := h.Repository.GetDraftStage(currentEngineerID)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "ошибка получения черновика")
		return
	}

	imageURL, videoURL := withDefaults(stage.ImageURL, stage.VideoURL)

	ctx.HTML(http.StatusOK, "draft.html", gin.H{
		"stage":    stage,
		"found":    found,
		"imageURL": imageURL,
		"videoURL": videoURL,
	})
}

func (h *Handler) CreateDraftLifecycleStage(ctx *gin.Context) {
	title := ctx.PostForm("title")

	if _, err := h.Repository.CreateOrGetDraftStage(currentEngineerID, title, "", ""); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "не удалось создать черновик")
		return
	}

	ctx.Redirect(http.StatusFound, "/stages/draft")
}

func (h *Handler) PublishLifecycleStage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "некорректный id этапа")
		return
	}

	description := ctx.PostForm("description")

	pricePerUnit, err := strconv.Atoi(ctx.PostForm("pricePerUnit"))
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "некорректная цена за единицу")
		return
	}

	stagePosition, err := strconv.Atoi(ctx.PostForm("stagePosition"))
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "некорректный порядковый номер")
		return
	}

	if err := h.Repository.PublishStage(uint(id64), description, pricePerUnit, stagePosition); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "не удалось опубликовать этап")
		return
	}

	ctx.Redirect(http.StatusFound, "/stages")
}

func (h *Handler) DeleteLifecycleStage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "некорректный id этапа")
		return
	}

	if err := h.Repository.DeleteStageRawSQL(uint(id64)); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "не удалось удалить этап")
		return
	}

	ctx.Redirect(http.StatusFound, "/stages")
}
