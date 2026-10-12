package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"tco-backend/internal/app/ds"
	"tco-backend/internal/app/serializer"
)

// RegisterEngineer godoc
// @Summary      Регистрация пользователя
// @Tags         engineers
// @Accept       json
// @Produce      json
// @Param        request  body  serializer.RegisterEngineerRequest  true  "Данные пользователя"
// @Success      201  {object}  serializer.Engineer
// @Failure      400
// @Failure      409
// @Router       /api/engineers [post]
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

// LoginEngineer godoc
// @Summary      Аутентификация
// @Description  Проверяет логин и пароль и возвращает JWT для заголовка Authorization: Bearer <token>.
// @Tags         engineers
// @Accept       json
// @Produce      json
// @Param        request  body  serializer.LoginEngineerRequest  true  "Логин и пароль"
// @Success      200  {object}  serializer.LoginEngineerResponse
// @Failure      400
// @Failure      401
// @Router       /api/engineers/login [post]
func (h *Handler) LoginEngineer(ctx *gin.Context) {
	var req serializer.LoginEngineerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	engineer, err := h.Repository.GetEngineerByLogin(req.Login)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusUnauthorized, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(engineer.PasswordHash), []byte(req.Password)); err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(h.Config.JWT.ExpiresIn).Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    "tco-backend",
		},
		EngineerID: engineer.ID,
	})
	strToken, err := token.SignedString([]byte(h.Config.JWT.Secret))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.LoginEngineerResponse{
		AccessToken: strToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.Config.JWT.ExpiresIn.Seconds()),
	})
}

// LogoutEngineer godoc
// @Summary      Деавторизация
// @Description  Добавляет текущий JWT в блеклист Redis до окончания срока его действия.
// @Tags         engineers
// @Security     BearerAuth
// @Success      200
// @Failure      401
// @Router       /api/engineers/logout [post]
func (h *Handler) LogoutEngineer(ctx *gin.Context) {
	jwtStr := strings.TrimPrefix(ctx.GetHeader("Authorization"), jwtPrefix)

	claims := &ds.JWTClaims{}
	if _, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.Secret), nil
	}); err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	ttl := time.Until(time.Unix(claims.ExpiresAt, 0))
	if err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, ttl); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusOK)
}
