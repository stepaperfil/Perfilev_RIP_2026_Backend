package api

import (
	"log"

	"github.com/gin-gonic/gin"

	"tco-backend/internal/app/dsn"
	"tco-backend/internal/app/handler"
	"tco-backend/internal/app/repository"
	"tco-backend/internal/app/storage"
)

func StartServer() {
	log.Println("Starting server")

	minioClient, err := storage.NewClient()
	if err != nil {
		log.Fatalf("ошибка подключения к MinIO: %v", err)
	}

	repo, err := repository.NewRepository(dsn.FromEnv(), minioClient, storage.Bucket())
	if err != nil {
		log.Fatalf("ошибка подключения к БД: %v", err)
	}

	h := handler.NewHandler(repo)

	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/life_cycle_stages", h.GetLifeCycleStages)
		api.GET("/life_cycle_stages/feed", h.GetLifeCycleStageFeed)
		api.GET("/life_cycle_stages/draft", h.GetDraftLifeCycleStage)
		api.POST("/life_cycle_stages", h.CreateLifeCycleStage)
		api.PUT("/life_cycle_stages/:id/publish", h.PublishLifeCycleStage)
		api.DELETE("/life_cycle_stages/:id", h.DeleteLifeCycleStage)
		api.POST("/life_cycle_stages/:id/like", h.LikeLifeCycleStage)

		api.POST("/engineers", h.RegisterEngineer)
		api.POST("/engineers/login", h.LoginEngineer)
		api.POST("/engineers/logout", h.LogoutEngineer)
	}

	if err := r.Run(); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}
