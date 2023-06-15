package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// AppSettingCreate creates a new AppSetting.
// Handles submission of new AppSetting form.
// POST /app_settings/new
func AppSettingCreate(c *gin.Context) {

}

// GET /app_settings/delete/:id
// POST /app_settings/delete/:id
func AppSettingDelete(c *gin.Context) {

}

// GET /app_settings/edit/:id
func AppSettingEdit(c *gin.Context) {
	id := c.Param("id")
	setting, err := core.AppSettingFind(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	data := gin.H{
		"form": setting.ToForm(),
	}
	c.HTML(http.StatusOK, "app_setting/form.html", data)
}

// GET /app_settings
func AppSettingIndex(c *gin.Context) {
	offset := c.GetInt("offset")
	limit := c.GetInt("limit")
	if limit < 1 {
		limit = 25
	}
	items, err := core.AppSettingList("obj_name", limit, offset)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	data := gin.H{
		"items": items,
	}
	c.HTML(http.StatusOK, "app_setting/list.html", data)
}

// GET /app_settings/new
func AppSettingNew(c *gin.Context) {

}

// GET /app_settings/show/:id
func AppSettingShow(c *gin.Context) {

}

// PUT /app_settings/edit/:id
// POST /app_settings/edit/:id
func AppSettingUpdate(c *gin.Context) {
	setting := &core.AppSetting{}
	err := c.Bind(setting)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = setting.Save()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, "/app_settings")
}
