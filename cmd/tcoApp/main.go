package main

import (
	"log"

	"github.com/joho/godotenv"

	"tco-backend/internal/app/api"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("файл .env не найден, используются значения по умолчанию")
	}

	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
