package main

import (
	"PRReviewer/config"
	"PRReviewer/internal/app"
)

func main() {
	cfg := config.MustLoadConfig()

	application := app.New(cfg)

	application.Run()

}
