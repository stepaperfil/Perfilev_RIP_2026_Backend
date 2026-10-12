package api

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "tco-backend/docs"
	"tco-backend/internal/app/config"
	"tco-backend/internal/app/dsn"
	"tco-backend/internal/app/handler"
	"tco-backend/internal/app/redis"
	"tco-backend/internal/app/repository"
	"tco-backend/internal/app/storage"
)

func StartServer() {
	log.Println("Starting server")

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("ошибка конфигурации: %v", err)
	}

	minioClient, err := storage.NewClient()
	if err != nil {
		log.Fatalf("ошибка подключения к MinIO: %v", err)
	}

	repo, err := repository.NewRepository(dsn.FromEnv(), minioClient, storage.Bucket())
	if err != nil {
		log.Fatalf("ошибка подключения к БД: %v", err)
	}

	redisClient, err := redis.New(context.Background(), cfg.Redis)
	if err != nil {
		log.Fatalf("ошибка подключения к Redis: %v", err)
	}
	defer redisClient.Close()

	h := handler.NewHandler(repo, redisClient, cfg)

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		guest := api.Group("")
		guest.Use(h.WithOptionalAuth)
		guest.GET("/life_cycle_stages", h.GetLifeCycleStages)
		guest.GET("/life_cycle_stages/feed", h.GetLifeCycleStageFeed)
		guest.POST("/engineers", h.RegisterEngineer)
		guest.POST("/engineers/login", h.LoginEngineer)

		creator := api.Group("")
		creator.Use(h.WithAuthCheck)
		creator.GET("/life_cycle_stages/draft", h.GetDraftLifeCycleStage)
		creator.POST("/life_cycle_stages", h.CreateLifeCycleStage)
		creator.PUT("/life_cycle_stages/:id/publish", h.PublishLifeCycleStage)
		creator.DELETE("/life_cycle_stages/:id", h.DeleteLifeCycleStage)
		creator.POST("/life_cycle_stages/:id/like", h.LikeLifeCycleStage)
		creator.POST("/engineers/logout", h.LogoutEngineer)
	}

	if err := r.Run(); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}
