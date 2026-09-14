package api

import (
	"html/template"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"tco-backend/internal/app/handler"
	"tco-backend/internal/app/repository"
)

var assetVersion = strconv.FormatInt(time.Now().Unix(), 10)

func plus1(n int) int {
	return n + 1
}

func formatPrice(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.Itoa(n)

	var groups []string
	for len(s) > 3 {
		groups = append([]string{s[len(s)-3:]}, groups...)
		s = s[:len(s)-3]
	}
	groups = append([]string{s}, groups...)

	result := strings.Join(groups, " ")
	if neg {
		result = "-" + result
	}
	return result
}

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	h := handler.NewHandler(repo)

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"formatPrice":  formatPrice,
		"plus1":        plus1,
		"assetVersion": func() string { return assetVersion },
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/stages", h.GetLifecycleStages)
	r.GET("/stages/draft", h.GetDraftLifecycleStage)
	r.GET("/stages/:id", h.GetLifecycleStage)

	r.Run()
	log.Println("Server down")
}
