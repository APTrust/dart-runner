package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/server"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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

func NewPostRequest(endpointUrl string, params url.Values) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, endpointUrl, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params.Encode())))
	return req, err
}

func DoSimpleGetTest(t *testing.T, endpointUrl string, expected []string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, endpointUrl, nil)
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	html := w.Body.String()
	ok, notFound := AssertContainsAllStrings(html, expected)
	assert.True(t, ok, "Missing from page %s: %v", endpointUrl, notFound)
}
