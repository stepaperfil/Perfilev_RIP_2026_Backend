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
	defaultImageURL = "/static/img/default-life-cycle-stage.jpg"
	defaultVideoURL = "/static/video/default-life-cycle-stage.mp4"
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

type lifeCycleStageTileView struct {
	LifeCycleStage ds.LifeCycleStage
	ImageURL       string
	LikesCount     int64
}

func (h *Handler) GetLifeCycleStages(ctx *gin.Context) {
	maxPricePerUnitQuery := ctx.Query("maxPricePerUnit")
	hasMaxPricePerUnit := false
	maxPricePerUnit := 0
	if maxPricePerUnitQuery != "" {
		v, err := strconv.Atoi(maxPricePerUnitQuery)
		if err != nil {
			logrus.Error(err)
		} else {
			maxPricePerUnit = v
			hasMaxPricePerUnit = true
		}
	}

	lifeCycleStages, err := h.Repository.GetPublishedLifeCycleStages(maxPricePerUnit, hasMaxPricePerUnit)
	if err != nil {
		logrus.Error(err)
	}

	views := make([]lifeCycleStageTileView, 0, len(lifeCycleStages))
	for _, lifeCycleStage := range lifeCycleStages {
		imageURL, _ := withDefaults(lifeCycleStage.ImageURL, lifeCycleStage.VideoURL)

		likes, err := h.Repository.GetLifeCycleStageLikesCount(lifeCycleStage.ID)
		if err != nil {
			logrus.Error(err)
		}

		views = append(views, lifeCycleStageTileView{
			LifeCycleStage: lifeCycleStage,
			ImageURL:       imageURL,
			LikesCount:     likes,
		})
	}

	ctx.HTML(http.StatusOK, "life_cycle_stages.html", gin.H{
		"lifeCycleStages": views,
		"maxPricePerUnit": maxPricePerUnitQuery,
	})
}

func (h *Handler) GetLifeCycleStage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "некорректный id этапа")
		return
	}
	id := uint(id64)

	var lifeCycleStage ds.LifeCycleStage
	if ctx.Query("next") == "true" {
		lifeCycleStage, err = h.Repository.GetNextLifeCycleStage(id)
	} else {
		lifeCycleStage, err = h.Repository.GetLifeCycleStage(id)
	}
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "этап не найден")
		return
	}

	likes, err := h.Repository.GetLifeCycleStageLikesCount(lifeCycleStage.ID)
	if err != nil {
		logrus.Error(err)
	}

	imageURL, videoURL := withDefaults(lifeCycleStage.ImageURL, lifeCycleStage.VideoURL)

	ctx.HTML(http.StatusOK, "life_cycle_stage.html", gin.H{
		"lifeCycleStage": lifeCycleStage,
		"imageURL":       imageURL,
		"videoURL":       videoURL,
		"likesCount":     likes,
	})
}

func (h *Handler) GetDraftLifeCycleStage(ctx *gin.Context) {
	lifeCycleStage, found, err := h.Repository.GetDraftLifeCycleStage(currentEngineerID)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "ошибка получения черновика")
		return
	}

	imageURL, videoURL := withDefaults(lifeCycleStage.ImageURL, lifeCycleStage.VideoURL)

	ctx.HTML(http.StatusOK, "life_cycle_stage_draft.html", gin.H{
		"lifeCycleStage": lifeCycleStage,
		"found":          found,
		"imageURL":       imageURL,
		"videoURL":       videoURL,
	})
}

func (h *Handler) CreateDraftLifeCycleStage(ctx *gin.Context) {
	title := ctx.PostForm("title")

	if _, err := h.Repository.CreateOrGetDraftLifeCycleStage(currentEngineerID, title, "", ""); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "не удалось создать черновик")
		return
	}

	ctx.Redirect(http.StatusFound, "/life_cycle_stages/draft")
}

func (h *Handler) PublishLifeCycleStage(ctx *gin.Context) {
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

	if err := h.Repository.PublishLifeCycleStage(uint(id64), description, pricePerUnit, stagePosition); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "не удалось опубликовать этап")
		return
	}

	ctx.Redirect(http.StatusFound, "/life_cycle_stages")
}

func (h *Handler) DeleteLifeCycleStage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "некорректный id этапа")
		return
	}

	if err := h.Repository.DeleteLifeCycleStageRawSQL(uint(id64)); err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, "не удалось удалить этап")
		return
	}

	ctx.Redirect(http.StatusFound, "/life_cycle_stages")
}
