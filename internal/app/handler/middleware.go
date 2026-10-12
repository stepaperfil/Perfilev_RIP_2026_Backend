package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	goredis "github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"

	"tco-backend/internal/app/ds"
)

const (
	jwtPrefix     = "Bearer "
	engineerIDKey = "engineer_id"
)

var errNoToken = errors.New("no bearer token")

// Возвращает id инженера из токена. Токен в блеклисте или с неверной подписью считается недействительным.
func (h *Handler) parseToken(ctx *gin.Context) (uint, error) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		return 0, errNoToken
	}
	jwtStr = jwtStr[len(jwtPrefix):]

	err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
	if err == nil {
		return 0, errors.New("token is in blacklist")
	}
	if !errors.Is(err, goredis.Nil) {
		return 0, err
	}

	token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(h.Config.JWT.Secret), nil
	})
	if err != nil {
		return 0, err
	}

	return token.Claims.(*ds.JWTClaims).EngineerID, nil
}

// Для методов создателя: без действующего токена запрос прерывается с кодом 401.
func (h *Handler) WithAuthCheck(ctx *gin.Context) {
	engineerID, err := h.parseToken(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	ctx.Set(engineerIDKey, engineerID)
}

// Для методов чтения: доступны гостю, а при наличии токена учитывается текущий пользователь (is_mine, is_liked).
func (h *Handler) WithOptionalAuth(ctx *gin.Context) {
	engineerID, err := h.parseToken(ctx)
	if err == nil {
		ctx.Set(engineerIDKey, engineerID)
	}
}

// id текущего пользователя; 0 - гость.
func currentEngineerID(ctx *gin.Context) uint {
	return ctx.GetUint(engineerIDKey)
}
