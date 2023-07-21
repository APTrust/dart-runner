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
	logFile, err := core.Dart.Paths.LogFile()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	templateData := gin.H{
		"version":      "Version goes here",
		"appPath":      "App path goes here",
		"userDataPath": core.Dart.Paths.DataDir,
		"logFilePath":  logFile,
	}
	c.HTML(http.StatusOK, "about/index.html", templateData)
}

// GET /open_external
// This is an AJAX call.
// TODO: Make context-sensitive. Go to the right page!
func OpenExternalUrl(c *gin.Context) {
	externalUrl := c.Query("url")
	command := "open"
	if runtime.GOOS == "windows" {
		command = "start"
	}
	cmd := exec.Command(command, externalUrl)
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

// GET /tail_log
func TailLog(c *gin.Context) {
	// powershell -command Get-Content Desktop\sample.log -Wait
	// tail -f
}

// GET /open_log
func OpenLog(c *gin.Context) {
	logFile, err := core.Dart.Paths.LogFile()
	if err != nil {
		data := map[string]string{
			"status": strconv.Itoa(http.StatusInternalServerError),
			"error":  err.Error(),
		}
		c.JSON(http.StatusInternalServerError, data)
		return
	}
	command := "open"
	if runtime.GOOS == "windows" {
		command = "start"
	}
	cmd := exec.Command(command, logFile)
	err = cmd.Start()
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

// GET /open_log
func OpenLogFolder(c *gin.Context) {
	command := "open"
	if runtime.GOOS == "windows" {
		command = "start"
	}
	cmd := exec.Command(command, core.Dart.Paths.LogDir)
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
