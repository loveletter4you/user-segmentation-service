package main

import (
	"github.com/loveletter4you/user-segmentation-service/internal/app"
)

const configPath = "config/config.yaml"

// @title User segment app
// @version 1.0
// @description Api server dynamic user segments

// @host      localhost:8080
// @BasePath  /api/
func main() {
	app.Run(configPath)
}
