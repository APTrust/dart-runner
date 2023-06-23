package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// GET /about
func AboutShow(c *gin.Context) {
	templateData := gin.H{
		"version":      "Version goes here",
		"appPath":      "App path goes here",
		"userDataPath": core.Dart.Paths.DataDir,
		"logFilePath":  core.Dart.Paths.LogDir,
	}
	c.HTML(http.StatusOK, "about/index.html", templateData)
}
