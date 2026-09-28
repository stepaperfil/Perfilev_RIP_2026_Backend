package main

import (
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tco-backend/internal/app/ds"
	"tco-backend/internal/app/dsn"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("файл .env не найден, используются значения по умолчанию")
	}

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}

	if err := db.AutoMigrate(&ds.Engineer{}, &ds.LifeCycleStage{}, &ds.LifeCycleStageLike{}); err != nil {
		log.Fatalf("ошибка миграции: %v", err)
	}

	defaultEngineer := ds.Engineer{ID: 1, FullName: "Инженер по умолчанию", Login: "default.engineer", PasswordHash: "$2a$10$placeholderPlaceholderPlaceholderPlaceh"}
	if err := db.FirstOrCreate(&defaultEngineer, ds.Engineer{ID: 1}).Error; err != nil {
		log.Fatalf("не удалось создать инженера по умолчанию: %v", err)
	}

	log.Println("Миграция выполнена успешно")
}
