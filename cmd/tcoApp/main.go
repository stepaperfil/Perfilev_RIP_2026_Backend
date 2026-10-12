package main

import (
	"log"

	"github.com/joho/godotenv"

	"tco-backend/internal/app/api"
)

// @title           TCO API
// @version         1.0
// @description     Веб-сервис расчёта стоимости владения активом: этапы жизненного цикла и пользователи.

// @host      localhost:8080
// @BasePath  /
// @schemes   http

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Введите "Bearer <токен из /api/engineers/login>"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("файл .env не найден, используются значения по умолчанию")
	}

	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
