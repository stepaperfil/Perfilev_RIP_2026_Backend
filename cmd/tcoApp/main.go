package main

import (
	"log"

	"tco-backend/internal/app/api"
)

func main() {
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
