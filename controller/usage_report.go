package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"veloera/model"
	"veloera/service"
)

type createReportRequest struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

func GetUsageReports(c *gin.Context) {
	reports, _, err := model.GetUsageReports(0, 100)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": reports})
}

func GetUsageReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	report, err := model.GetUsageReportById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": report})
}

func CreateUsageReport(c *gin.Context) {
	req := createReportRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	data, err := service.GenerateUsageReport(req.Items)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	b, _ := json.Marshal(data)
	report := model.UsageReport{
		Name:    req.Name,
		Options: strings.Join(req.Items, ","),
		Data:    string(b),
	}
	if err = report.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": report})
}
