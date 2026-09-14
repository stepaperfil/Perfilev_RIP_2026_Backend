package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tco-backend/internal/app/ds"
	"tco-backend/internal/app/dsn"
)

func main() {
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}

	if err := db.AutoMigrate(&ds.Engineer{}, &ds.LifecycleStage{}, &ds.StageLike{}); err != nil {
		log.Fatalf("ошибка миграции: %v", err)
	}

	defaultEngineer := ds.Engineer{ID: 1, FullName: "Инженер по умолчанию"}
	if err := db.FirstOrCreate(&defaultEngineer, ds.Engineer{ID: 1}).Error; err != nil {
		log.Fatalf("не удалось создать инженера по умолчанию: %v", err)
	}

	log.Println("Миграция выполнена успешно")
}
