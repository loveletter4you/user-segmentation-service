package app

import (
	"github.com/loveletter4you/user-segmentation-service/config"
	"github.com/loveletter4you/user-segmentation-service/internal/httpserver"
	"log"
)

func Run(configPath string) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	r := httpserver.NewServer()
	err = r.StartServer(cfg)
	log.Fatal(err)
}
