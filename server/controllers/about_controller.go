package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AboutShow(c *gin.Context) {
	templateData := gin.H{
		"version":      "Version goes here",
		"appPath":      "App path goes here",
		"userDataPath": "User data path goes here",
		"logFilePath":  "Log file path goes here",
	}
	c.HTML(http.StatusOK, "about/index.html", templateData)
}
