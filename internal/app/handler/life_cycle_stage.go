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

// GetLifeCycleStages godoc
// @Summary      Список этапов жизненного цикла
// @Description  Опубликованные этапы с фильтром по цене за единицу. Доступно гостю.
// @Tags         life_cycle_stages
// @Produce      json
// @Param        max_price_per_unit  query  int  false  "Максимальная цена за единицу"
// @Success      200  {array}  serializer.LifeCycleStageListItem
// @Failure      400
// @Router       /api/life_cycle_stages [get]
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

	engineerID := currentEngineerID(ctx)
	lifeCycleStages, err := h.Repository.GetPublishedLifeCycleStages(engineerID, maxPricePerUnit)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.NewLifeCycleStageList(lifeCycleStages, engineerID))
}

// GetLifeCycleStageFeed godoc
// @Summary      Лента этапов
// @Description  Без id - первый этап ленты, с id - этот этап, с id и next=true - следующий за ним. Доступно гостю.
// @Tags         life_cycle_stages
// @Produce      json
// @Param        id    query  int     false  "ID этапа"
// @Param        next  query  string  false  "true - следующий этап"
// @Success      200  {object}  serializer.LifeCycleStageFeedItem
// @Failure      400
// @Failure      404
// @Router       /api/life_cycle_stages/feed [get]
func (h *Handler) GetLifeCycleStageFeed(ctx *gin.Context) {
	engineerID := currentEngineerID(ctx)

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

// GetDraftLifeCycleStage godoc
// @Summary      Черновик текущего пользователя
// @Tags         life_cycle_stages
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  serializer.LifeCycleStage
// @Failure      401
// @Failure      404
// @Router       /api/life_cycle_stages/draft [get]
func (h *Handler) GetDraftLifeCycleStage(ctx *gin.Context) {
	lifeCycleStage, err := h.Repository.GetDraftLifeCycleStage(currentEngineerID(ctx))
	if err != nil {
		h.repositoryError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.NewLifeCycleStage(lifeCycleStage))
}

// CreateLifeCycleStage godoc
// @Summary      Добавление этапа
// @Description  Создаёт черновик этапа текущего пользователя, фото и видео загружаются в MinIO.
// @Tags         life_cycle_stages
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        title  formData  string  true  "Название"
// @Param        image  formData  file    true  "Изображение"
// @Param        video  formData  file    true  "Видео"
// @Success      201  {object}  serializer.LifeCycleStage
// @Failure      400
// @Failure      401
// @Failure      409
// @Router       /api/life_cycle_stages [post]
func (h *Handler) CreateLifeCycleStage(ctx *gin.Context) {
	engineerID := currentEngineerID(ctx)

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

// LikeLifeCycleStage godoc
// @Summary      Лайк этапа
// @Description  like = 1 ставит лайк текущего пользователя, like = 0 снимает его.
// @Tags         life_cycle_stages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                                   true  "ID этапа"
// @Param        request  body  serializer.LikeLifeCycleStageRequest  true  "Лайк"
// @Success      200  {object}  serializer.LifeCycleStageFeedItem
// @Failure      400
// @Failure      401
// @Failure      404
// @Router       /api/life_cycle_stages/{id}/like [post]
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

	engineerID := currentEngineerID(ctx)
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

// PublishLifeCycleStage godoc
// @Summary      Публикация черновика этапа
// @Tags         life_cycle_stages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                                      true  "ID этапа"
// @Param        request  body  serializer.PublishLifeCycleStageRequest  true  "Данные публикации"
// @Success      200  {object}  serializer.LifeCycleStage
// @Failure      400
// @Failure      401
// @Failure      404
// @Router       /api/life_cycle_stages/{id}/publish [put]
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

	lifeCycleStage, err := h.Repository.PublishLifeCycleStage(currentEngineerID(ctx), id, req.Description, *req.PricePerUnit, *req.StagePosition)
	if err != nil {
		h.repositoryError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.NewLifeCycleStage(lifeCycleStage))
}

// DeleteLifeCycleStage godoc
// @Summary      Удаление этапа
// @Description  Логическое удаление этапа текущего пользователя.
// @Tags         life_cycle_stages
// @Security     BearerAuth
// @Param        id  path  int  true  "ID этапа"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      404
// @Router       /api/life_cycle_stages/{id} [delete]
func (h *Handler) DeleteLifeCycleStage(ctx *gin.Context) {
	id, err := parseID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteLifeCycleStage(currentEngineerID(ctx), id); err != nil {
		h.repositoryError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}
