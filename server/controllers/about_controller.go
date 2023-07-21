package controllers

import (
	"net/http"
	"os/exec"
	"runtime"
	"strconv"

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

// GET /help
// This is an AJAX call.
// TODO: Make context-sensitive. Go to the right page!
func OpenHelpInNewWindow(c *gin.Context) {
	helpUrl := "https://aptrust.github.io/dart-docs/users/getting_started/"
	command := "open"
	if runtime.GOOS == "windows" {
		command = "start"
	}
	cmd := exec.Command(command, helpUrl)
	err := cmd.Start()
	if err != nil {
		data := map[string]string{
			"status": strconv.Itoa(http.StatusInternalServerError),
			"error":  err.Error(),
		}
		c.JSON(http.StatusInternalServerError, data)
		return
	}
	data := map[string]string{
		"status": strconv.Itoa(http.StatusOK),
		"result": "OK",
	}
	c.JSON(http.StatusOK, data)
}
