package controllers

import (
	"fmt"
	"runtime/debug"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

func AbortWithErrorHTML(c *gin.Context, status int, err error) {
	logRequestError(c, status, err)
	c.HTML(status, "error/show.html", getResponseData(err))
	c.Abort()
}

func AbortWithErrorJSON(c *gin.Context, status int, err error) {
	logRequestError(c, status, err)
	c.JSON(status, getResponseData(err))
	c.Abort()
}

func getResponseData(err error) gin.H {
	stack := debug.Stack()
	data := gin.H{
		"error":      err.Error(),
		"stackTrace": string(stack),
	}
	return data
}

func logRequestError(c *gin.Context, status int, err error) {
	core.Dart.Log.Error(fmt.Sprintf("Returned status %d for %s %s", status, c.Request.Method, c.Request.URL.RequestURI()))
	core.Dart.Log.Error(err.Error())
}
