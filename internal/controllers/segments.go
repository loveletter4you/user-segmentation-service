package controllers

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/loveletter4you/user-segmentation-service/internal/model"
	"net/http"
	"os"
	"strconv"
)

type userSegmentRequest struct {
	AppendSlugs []string `json:"appendSlugs" binding:"required"`
	DeleteSlugs []string `json:"deleteSlugs" binding:"required"`
	TimeToLive  uint     `json:"timeToLive"`
}

type segmentRequest struct {
	Slug       string `json:"slug" binding:"required"`
	Percent    int    `json:"percent"`
	TimeToLive uint   `json:"timeToLive"`
}

// Create segment by slug
// @Summary Create segment
// @Tags segments
// @Accept json
// @Produce json
// @Param slug body segmentRequest true "create segment body"
// @Success 200 {object} model.Segment
// @Failure 400
// @Router /segment [post]
func (ctr *Controller) CreateSegment(c *gin.Context) {
	var segmentRequest segmentRequest
	if err := c.BindJSON(&segmentRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if segmentRequest.Slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input slug"})
	}
	segment := model.Segment{Slug: segmentRequest.Slug}

	tx, err := ctr.storage.StartTransaction()
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if err := ctr.storage.Segment().CreateSegment(tx, &segment); err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if segmentRequest.Percent != 0 {
		segmentAutoInsert, err := ctr.storage.Segment().CreateSegmentAutoInsert(tx, segmentRequest.Slug,
			segmentRequest.Percent, segmentRequest.TimeToLive)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		users, err := ctr.storage.User().GetUsers(tx)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctr.storage.Segment().AutoCreateUserSegments(tx, users, segmentAutoInsert)
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, segment)
}

// Get all segments
// @Summary Get segments
// @Tags segments
// @Produce json
// @Success 200
// @Failure 400
// @Router /segments [get]
func (ctr *Controller) GetSegments(c *gin.Context) {
	segments, err := ctr.storage.Segment().GetSegments(nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := map[string]interface{}{
		"segments": segments,
	}
	c.JSON(http.StatusOK, response)
}

// Create user segment
// @Summary Create segment
// @Tags segments
// @Produce json
// @Param id path int true "id"
// @Param appendSlugs body userSegmentRequest true "create user segment body"
// @Failure 400
// @Router /user/{id}/segments [post]
func (ctr *Controller) CreateUserSegment(c *gin.Context) {
	userSegmentResponse := new(userSegmentRequest)

	if err := c.BindJSON(&userSegmentResponse); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addedSlugs := make([]string, 0)
	notAddedSlugs := make([]string, 0)
	deletedSlugs := make([]string, 0)
	notDeletedSlugs := make([]string, 0)

	tx, err := ctr.storage.StartTransaction()
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	for _, slug := range userSegmentResponse.AppendSlugs {
		_, err := ctr.storage.Segment().CreateUserSegment(tx, id, slug, userSegmentResponse.TimeToLive)
		if err != nil {
			notAddedSlugs = append(notAddedSlugs, slug)
			continue
		}
		addedSlugs = append(addedSlugs, slug)
	}
	for _, slug := range userSegmentResponse.DeleteSlugs {
		_, err := ctr.storage.Segment().DeleteUserSegment(tx, id, slug)
		if err != nil {
			notDeletedSlugs = append(notDeletedSlugs, slug)
			continue
		}
		deletedSlugs = append(deletedSlugs, slug)
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"addedSlugs":      addedSlugs,
		"notAddedSlugs":   notAddedSlugs,
		"deletedSlugs":    deletedSlugs,
		"notDeletedSlugs": notDeletedSlugs,
	}
	c.JSON(http.StatusOK, response)
}

// Get user segments
// @Summary Get user segments
// @Description Get active user segments
// @Tags segments
// @Param id path int true "id"
// @Produce json
// @Success 200
// @Failure 400
// @Router /user/{id}/segments [get]
func (ctr *Controller) GetUserSegments(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userSegments, err := ctr.storage.Segment().GetUserSegments(nil, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := map[string]interface{}{
		"userId":   id,
		"segments": userSegments,
	}
	c.JSON(http.StatusOK, response)
}

// Get user month report
// @Summary Get month report
// @Description Get active user segments
// @Tags segments
// @Produce json
// @Param id path int true "id"
// @Param month query int true "report month"
// @Param year query int true "report year"
// @Success 200
// @Failure 400
// @Router /user/{id}/report [get]
func (ctr *Controller) GetUserMonthReport(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	month, err := strconv.Atoi(c.Query("month"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	userSegments, err := ctr.storage.Segment().GetUserSegmentsMonthYear(nil, id, month, year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	records := make([][]string, 0)
	for _, userSegment := range userSegments {
		if int(userSegment.ActiveFrom.Month()) == month && userSegment.ActiveFrom.Year() == year {
			row := []string{
				strconv.Itoa(id),
				userSegment.Segment.Slug,
				"add",
				userSegment.ActiveFrom.Format("2006-01-02 15-04-05")}
			records = append(records, row)
		}
		if int(userSegment.ActiveTo.Month()) == month && userSegment.ActiveTo.Year() == year {
			row := []string{
				strconv.Itoa(id),
				userSegment.Segment.Slug,
				"delete",
				userSegment.ActiveFrom.Format("2006-01-02 15-04-05")}
			records = append(records, row)
		}
	}
	fileName := fmt.Sprintf("/static/user%d_%d-%d.csv", id, year, month)
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	w := csv.NewWriter(f)
	err = w.WriteAll(records)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := map[string]interface{}{
		"url": fileName,
	}
	c.JSON(http.StatusOK, response)

}
