package controllers

import "github.com/loveletter4you/user-segmentation-service/internal/storage"

type Controller struct {
	storage *storage.Storage
}

func NewController() *Controller {
	return &Controller{
		storage: storage.NewStorage(),
	}
}

func (ctr *Controller) OpenConnection(dbHost, dbPort, dbRoot, dbPassword, dbName string,
	connectionAttempt, connectionTimeout int) error {
	err := ctr.storage.Open(dbHost, dbPort, dbRoot, dbPassword, dbName, connectionAttempt, connectionTimeout)
	return err
}

func (ctr *Controller) CloseConnection() error {
	err := ctr.storage.Close()
	return err
}
