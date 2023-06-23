package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// DELETE /app_settings/delete/:id
// POST /app_settings/delete/:id
func AppSettingDelete(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	err := core.ObjDelete(request.QueryResult.AppSetting())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, "/app_settings")
}

// GET /app_settings/edit/:id
func AppSettingEdit(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	c.HTML(http.StatusOK, "app_setting/form.html", request.TemplateData)
}

// GET /app_settings
func AppSettingIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.AppSettings
	c.HTML(http.StatusOK, "app_setting/list.html", request.TemplateData)
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
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	err = core.ObjSave(setting)
	if err != nil {
		objectExistsInDB, _ := core.ObjExists(setting.ID)
		data := gin.H{
			"form":             setting.ToForm(),
			"objectExistsInDB": objectExistsInDB,
		}
		c.HTML(http.StatusBadRequest, "app_setting/form.html", data)
		return
	}
	c.Redirect(http.StatusFound, "/app_settings")
}
