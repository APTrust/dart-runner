package controllers_test

import (
	"strings"

	"github.com/APTrust/dart-runner/server"
	"github.com/gin-gonic/gin"
)

var dartServer *gin.Engine

func init() {
	dartServer = server.InitAppEngine(true)
}

func AssertContainsAllStrings(html string, expected []string) (allFound bool, notFound []string) {
	allFound = true
	for _, expectedString := range expected {
		if !strings.Contains(html, expectedString) {
			allFound = false
			notFound = append(notFound, expectedString)
		}
	}
	return allFound, notFound
}
