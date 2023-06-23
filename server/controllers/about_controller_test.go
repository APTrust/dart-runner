package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAboutShow(t *testing.T) {
	expected := []string{
		"Version",
		"App Location",
		"Data Location",
		"Log File",
		"Academic Preservation Trust",
		"GitHub",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/about", nil)
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	html := w.Body.String()

	ok, notFound := AssertContainsAllStrings(html, expected)
	assert.True(t, ok, "Missing from page: %v", notFound)
}
