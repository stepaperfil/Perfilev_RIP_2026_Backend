package repository

import "fmt"

type LifecycleStage struct {
	ID          int
	Title       string
	Description string
	ImageKey    string
	VideoKey    string
	Status      string

	Likes []int

	PricePerUnit  int
	StagePosition int
}

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

func (r *Repository) GetLifecycleStages() ([]LifecycleStage, error) {
	stages := []LifecycleStage{
		{
			ID:            1,
			Title:         "Закупка станка ЧПУ DMU 50",
			Description:   "Покупка нового фрезерного станка с ЧПУ для цеха №2, включая доставку и монтаж",
			ImageKey:      "purchase-dmu50.jpg",
			VideoKey:      "purchase-dmu50.mp4",
			Status:        "published",
			Likes:         makeLikes(24),
			PricePerUnit:  11475400,
			StagePosition: 1,
		},
		{
			ID:            2,
			Title:         "Ввод в эксплуатацию",
			Description:   "Пусконаладочные работы, калибровка осей и обучение оператора станка",
			ImageKey:      "commissioning.jpg",
			VideoKey:      "commissioning.mp4",
			Status:        "published",
			Likes:         makeLikes(9),
			PricePerUnit:  94860,
			StagePosition: 2,
		},
		{
			ID:            3,
			Title:         "Плановое ТО №1",
			Description:   "Регулярное техническое обслуживание по регламенту производителя",
			ImageKey:      "maintenance-1.jpg",
			VideoKey:      "maintenance.mp4",
			Status:        "published",
			Likes:         makeLikes(15),
			PricePerUnit:  79050,
			StagePosition: 3,
		},
		{
			ID:            4,
			Title:         "Замена комплектующих",
			Description:   "Плановая замена режущего инструмента на фрезерном станке ЧПУ, цех №2",
			ImageKey:      "steel.jpg",
			VideoKey:      "repair.mp4",
			Status:        "draft",
			Likes:         nil,
			PricePerUnit:  20000,
			StagePosition: 3,
		},
		{
			ID:            5,
			Title:         "Внеплановый ремонт №1",
			Description:   "Диагностика и устранение неисправности шпиндельного узла",
			ImageKey:      "repair-1.jpg",
			VideoKey:      "repair.mp4",
			Status:        "published",
			Likes:         makeLikes(6),
			PricePerUnit:  63240,
			StagePosition: 3,
		},
		{
			ID:            6,
			Title:         "Плановое ТО №2",
			Description:   "Регулярное техническое обслуживание по регламенту производителя",
			ImageKey:      "maintenance-2.jpg",
			VideoKey:      "maintenance.mp4",
			Status:        "published",
			Likes:         makeLikes(12),
			PricePerUnit:  79050,
			StagePosition: 3,
		},
		{
			ID:            7,
			Title:         "Модернизация ЧПУ",
			Description:   "Ретрофит системы ЧПУ и приводов для продления срока службы станка",
			ImageKey:      "retrofit.jpg",
			VideoKey:      "retrofit.mp4",
			Status:        "published",
			Likes:         makeLikes(11),
			PricePerUnit:  8000000,
			StagePosition: 3,
		},
		{
			ID:            8,
			Title:         "Плановое ТО №3",
			Description:   "Регулярное техническое обслуживание по регламенту производителя",
			ImageKey:      "maintenance-3.jpg",
			VideoKey:      "maintenance.mp4",
			Status:        "published",
			Likes:         makeLikes(8),
			PricePerUnit:  79050,
			StagePosition: 3,
		},
		{
			ID:            9,
			Title:         "Внеплановый ремонт №2",
			Description:   "Диагностика и устранение неисправности механизма подачи охлаждающей жидкости",
			ImageKey:      "repair-2.jpg",
			VideoKey:      "repair.mp4",
			Status:        "published",
			Likes:         makeLikes(4),
			PricePerUnit:  63240,
			StagePosition: 3,
		},
		{
			ID:            10,
			Title:         "Утилизация",
			Description:   "Демонтаж и сдача станка на утилизацию по окончании срока службы",
			ImageKey:      "disposal.jpg",
			VideoKey:      "disposal.mp4",
			Status:        "published",
			Likes:         makeLikes(3),
			PricePerUnit:  110000,
			StagePosition: 4,
		},
	}

	if len(stages) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return stages, nil
}

func makeLikes(n int) []int {
	likes := make([]int, n)
	for i := range likes {
		likes[i] = 100 + i
	}
	return likes
}

func (r *Repository) GetLifecycleStage(id int) (LifecycleStage, error) {
	stages, err := r.GetLifecycleStages()
	if err != nil {
		return LifecycleStage{}, err
	}

	for _, stage := range stages {
		if stage.ID == id {
			return stage, nil
		}
	}
	return LifecycleStage{}, fmt.Errorf("этап не найден")
}

func (r *Repository) GetDraftLifecycleStage() (LifecycleStage, error) {
	stages, err := r.GetLifecycleStages()
	if err != nil {
		return LifecycleStage{}, err
	}

	for _, stage := range stages {
		if stage.Status == "draft" {
			return stage, nil
		}
	}
	return LifecycleStage{}, fmt.Errorf("черновик не найден")
}

func (r *Repository) GetPublishedLifecycleStages() ([]LifecycleStage, error) {
	stages, err := r.GetLifecycleStages()
	if err != nil {
		return nil, err
	}

	var result []LifecycleStage
	for _, stage := range stages {
		if stage.Status == "published" {
			result = append(result, stage)
		}
	}
	return result, nil
}

func (r *Repository) GetLifecycleStagesByMaxPrice(maxPrice int) ([]LifecycleStage, error) {
	stages, err := r.GetPublishedLifecycleStages()
	if err != nil {
		return []LifecycleStage{}, err
	}

	var result []LifecycleStage
	for _, stage := range stages {
		if stage.PricePerUnit <= maxPrice {
			result = append(result, stage)
		}
	}
	return result, nil
}

func (r *Repository) GetNextLifecycleStage(id int) (LifecycleStage, error) {
	stages, err := r.GetPublishedLifecycleStages()
	if err != nil {
		return LifecycleStage{}, err
	}

	for i, stage := range stages {
		if stage.ID == id && i+1 < len(stages) {
			return stages[i+1], nil
		}
	}
	return LifecycleStage{}, fmt.Errorf("следующий этап не найден")
}
