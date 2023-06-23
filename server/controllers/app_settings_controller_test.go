package controllers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSettingsList(t *testing.T) {
	defer core.ClearDartTable()
	s1 := core.NewAppSetting("Setting 1", "Value 1")
	s2 := core.NewAppSetting("Setting 2", "Value 2")
	assert.NoError(t, core.ObjSave(s1))
	assert.NoError(t, core.ObjSave(s2))

	expected := []string{
		"Application Settings",
		"Name",
		"Value",
		"New",
		"Setting 1",
		"Value 1",
		"Setting 2",
		"Value 2",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/app_settings", nil)
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	html := w.Body.String()

	ok, notFound := AssertContainsAllStrings(html, expected)
	assert.True(t, ok, "Missing from page: %v", notFound)
}

func TestAppSettingNew(t *testing.T) {
	expected := []string{
		"Application Setting",
		"AppSetting_Name",
		"AppSetting_Value",
		`name="Name"`,
		`name="Value"`,
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/app_settings/new", nil)
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	html := w.Body.String()

	ok, notFound := AssertContainsAllStrings(html, expected)
	assert.True(t, ok, "Missing from page: %v", notFound)
}

func TestAppSettingSaveEditDelete(t *testing.T) {
	defer core.ClearDartTable()
	testNewWithMisingParams(t)
	testNewSaveEditDeleteWithGoodParams(t)
}

func testNewWithMisingParams(t *testing.T) {
	expected := []string{
		"NameError",
		"Name cannot be empty",
		"ValueError",
		"Value cannot be empty",
	}

	data := url.Values{}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/app_settings/new", strings.NewReader(data.Encode()))
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	html := w.Body.String()
	ok, notFound := AssertContainsAllStrings(html, expected)
	assert.True(t, ok, "Missing from page: %v", notFound)
}

func testNewSaveEditDeleteWithGoodParams(t *testing.T) {
	expected := []string{
		"Application Settings",
		"Web Test Name 1",
		"Web Test Value 1",
	}

	data := url.Values{}
	data.Set("ID", uuid.NewString())
	data.Set("Name", "Web Test Name 1")
	data.Set("Value", "Web Test Value 1")
	data.Set("UserCanDelete", "true")

	// Submit the New App Setting form with valid params.
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/app_settings/new", strings.NewReader(data.Encode()))
	require.Nil(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/app_settings", w.Header().Get("Location"))

	// Make sure it was created
	w = httptest.NewRecorder()
	id := data.Get("ID")
	itemUrl := fmt.Sprintf("/app_settings/edit/%s", id)
	req, err = http.NewRequest(http.MethodGet, itemUrl, nil)
	require.Nil(t, err)
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	html := w.Body.String()
	ok, notFound := AssertContainsAllStrings(html, expected)
	assert.True(t, ok, "Missing from page: %v", notFound)

	// Submit the Edit App Setting form with updated params.
	w = httptest.NewRecorder()
	data.Set("Name", "Web Test Name Edited")
	data.Set("Value", "Web Test Value Edited")
	itemUrl = fmt.Sprintf("/app_settings/edit/%s", id)
	req, err = http.NewRequest(http.MethodPost, itemUrl, strings.NewReader(data.Encode()))
	require.Nil(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/app_settings", w.Header().Get("Location"))

	// Make sure it was updated
	expected[1] = "Web Test Name Edited"
	expected[2] = "Web Test Value Edited"
	w = httptest.NewRecorder()
	itemUrl = fmt.Sprintf("/app_settings/edit/%s", id)
	req, err = http.NewRequest(http.MethodGet, itemUrl, nil)
	require.Nil(t, err)
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	html = w.Body.String()
	ok, notFound = AssertContainsAllStrings(html, expected)
	assert.True(t, ok, "Missing from page: %v", notFound)

	// Test App Setting Delete
	w = httptest.NewRecorder()
	itemUrl = fmt.Sprintf("/app_settings/delete/%s", id)
	req, err = http.NewRequest(http.MethodPost, itemUrl, nil)
	require.Nil(t, err)
	dartServer.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/app_settings", w.Header().Get("Location"))

}
