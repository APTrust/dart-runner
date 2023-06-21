package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// GET /app_settings/delete/:id
// POST /app_settings/delete/:id
func AppSettingDelete(c *gin.Context) {
	id := c.Param("id")
	result := core.ObjFind(id)
	if result.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}
	err := core.ObjDelete(result.AppSetting())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, "/app_settings")
}

// GET /app_settings/edit/:id
func AppSettingEdit(c *gin.Context) {
	id := c.Param("id")
	result := core.ObjFind(id)
	if result.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}
	data := gin.H{
		"form": result.AppSetting().ToForm(),
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
	result := core.ObjList(constants.TypeAppSetting, "obj_name", limit, offset)
	if result.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}
	data := gin.H{
		"items": result.AppSettings,
	}
	c.HTML(http.StatusOK, "app_setting/list.html", data)
}

// GET /app_settings/new
func AppSettingNew(c *gin.Context) {
	setting := core.NewAppSetting("", "")
	data := gin.H{
		"form":                 setting.ToForm(),
		"suppressDeleteButton": true,
	}
	c.HTML(http.StatusOK, "app_setting/form.html", data)
}

// PUT /app_settings/edit/:id
// POST /app_settings/edit/:id
// POST /app_settings/new
func AppSettingSave(c *gin.Context) {
	setting := &core.AppSetting{}
	err := c.Bind(setting)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = core.ObjSave(setting)
	if err != nil {
		data := gin.H{
			"form": setting.ToForm(),
		}
		c.HTML(http.StatusOK, "app_setting/form.html", data)
		return
	}
	c.Redirect(http.StatusFound, "/app_settings")
}
