package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /internal_settings
func InternalSettingIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.InternalSettings
	c.HTML(http.StatusOK, "internal_setting/list.html", request.TemplateData)

}
