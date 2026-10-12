package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"tco-backend/internal/app/ds"
	"tco-backend/internal/app/serializer"
)

func (h *Handler) GetLifeCycleStages(ctx *gin.Context) {
	var maxPricePerUnit *int
	if q := ctx.Query("max_price_per_unit"); q != "" {
		v, err := strconv.Atoi(q)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		maxPricePerUnit = &v
	}

	engineerID := CurrentEngineerID()
	lifeCycleStages, err := h.Repository.GetPublishedLifeCycleStages(engineerID, maxPricePerUnit)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.NewLifeCycleStageList(lifeCycleStages, engineerID))
}

func (h *Handler) GetLifeCycleStageFeed(ctx *gin.Context) {
	engineerID := CurrentEngineerID()

	var lifeCycleStage ds.LifeCycleStageWithStats
	var err error
	if q := ctx.Query("id"); q == "" {
		lifeCycleStage, err = h.Repository.GetFirstFeedLifeCycleStage(engineerID)
	} else {
		id64, parseErr := strconv.ParseUint(q, 10, 64)
		if parseErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, parseErr)
			return
		}
		id := uint(id64)
		if ctx.Query("next") == "true" {
			lifeCycleStage, err = h.Repository.GetNextFeedLifeCycleStage(engineerID, id)
		} else {
			lifeCycleStage, err = h.Repository.GetFeedLifeCycleStage(engineerID, id)
		}
	}
	if err != nil {
		h.repositoryError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.NewLifeCycleStageFeedItem(lifeCycleStage))
}

func (h *Handler) GetDraftLifeCycleStage(ctx *gin.Context) {
	lifeCycleStage, err := h.Repository.GetDraftLifeCycleStage(CurrentEngineerID())
	if err != nil {
		h.repositoryError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.NewLifeCycleStage(lifeCycleStage))
}

func (h *Handler) CreateLifeCycleStage(ctx *gin.Context) {
	engineerID := CurrentEngineerID()

	_, err := h.Repository.GetDraftLifeCycleStage(engineerID)
	if err == nil {
		h.errorHandler(ctx, http.StatusConflict, errors.New("у инженера уже есть черновик"))
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	title := strings.TrimSpace(ctx.PostForm("title"))
	if title == "" || utf8.RuneCountInString(title) > 150 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректное название"))
		return
	}

	imageHeader, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	videoHeader, err := ctx.FormFile("video")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	imageType, err := detectContentType(imageHeader)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	imageExt, ok := imageExtensions[imageType]
	if !ok {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("неподдерживаемый тип изображения: "+imageType))
		return
	}

	videoType, err := detectContentType(videoHeader)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	videoExt, ok := videoExtensions[videoType]
	if !ok {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("неподдерживаемый тип видео: "+videoType))
		return
	}

	imageKey, err := newObjectKey(imageExt)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	videoKey, err := newObjectKey(videoExt)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := h.Repository.UploadFile(ctx, imageKey, imageHeader, imageType); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if err := h.Repository.UploadFile(ctx, videoKey, videoHeader, videoType); err != nil {
		h.removeFiles(ctx, imageKey)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	lifeCycleStage, err := h.Repository.CreateDraftLifeCycleStage(engineerID, title, imageKey, videoKey)
	if err != nil {
		h.removeFiles(ctx, imageKey, videoKey)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, serializer.NewLifeCycleStage(lifeCycleStage))
}

func (h *Handler) removeFiles(ctx *gin.Context, keys ...string) {
	for _, key := range keys {
		if err := h.Repository.RemoveFile(ctx, key); err != nil {
			logrus.Error(err)
		}
	}
}

func (h *Handler) LikeLifeCycleStage(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req serializer.LikeLifeCycleStageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	engineerID := CurrentEngineerID()
	if err := h.Repository.SetLifeCycleStageLike(engineerID, id, *req.Like == 1); err != nil {
		h.repositoryError(ctx, err)
		return
	}

	lifeCycleStage, err := h.Repository.GetFeedLifeCycleStage(engineerID, id)
	if err != nil {
		h.repositoryError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.NewLifeCycleStageFeedItem(lifeCycleStage))
}

func (h *Handler) PublishLifeCycleStage(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req serializer.PublishLifeCycleStageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	lifeCycleStage, err := h.Repository.PublishLifeCycleStage(CurrentEngineerID(), id, req.Description, *req.PricePerUnit, *req.StagePosition)
	if err != nil {
		h.repositoryError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.NewLifeCycleStage(lifeCycleStage))
}

func (h *Handler) DeleteLifeCycleStage(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteLifeCycleStage(CurrentEngineerID(), id); err != nil {
		h.repositoryError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}
