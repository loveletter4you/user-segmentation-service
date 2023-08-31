package controllers

import (
	"github.com/loveletter4you/user-segmentation-service/config"
	"github.com/loveletter4you/user-segmentation-service/internal/storage"
)

type Controller struct {
	storage *storage.Storage
}

func NewController() *Controller {
	return &Controller{
		storage: storage.NewStorage(),
	}
}

func (ctr *Controller) OpenConnection(config *config.Config) error {
	err := ctr.storage.Open(config)
	return err
}

func (ctr *Controller) CloseConnection() error {
	err := ctr.storage.Close()
	return err
}
