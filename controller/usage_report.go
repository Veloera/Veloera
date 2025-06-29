package controller

import (
	"net/http"
	"strconv"
	"veloera/model"

	"github.com/gin-gonic/gin"
)

func GetUsageReports(c *gin.Context) {
	reports, err := model.GetAllUsageReports()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": reports})
}

func GetUsageReportDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	report, err := model.GetUsageReport(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": report})
}

func CreateUsageReport(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	report, err := model.GenerateUsageReport(req.Name)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": report})
}
