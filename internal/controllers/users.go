package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/loveletter4you/user-segmentation-service/internal/model"
	"net/http"
)

// Create user
// @Summary Create user
// @Tags users
// @Produce json
// @Success 200 {object} model.User
// @Failure 500
// @Router /user [post]
func (ctr *Controller) CreateUser(c *gin.Context) {
	var user model.User

	tx, err := ctr.storage.StartTransaction()
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if err := ctr.storage.User().CreateUser(tx, &user); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := ctr.storage.Segment().AutoAddUserToSegments(tx, &user); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.New("can't auto add user to segments")})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Get users
// @Summary Get users
// @Tags users
// @Produce json
// @Success 200
// @Failure 500
// @Router /users [get]
func (ctr *Controller) GetUsers(c *gin.Context) {
	users, err := ctr.storage.User().GetUsers(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := map[string]interface{}{
		"users": users,
	}
	c.JSON(http.StatusOK, response)
}
