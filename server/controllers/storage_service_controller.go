package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// GET /storage_services/delete/:id
// POST /storage_services/delete/:id
func StorageServiceDelete(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	err := core.ObjDelete(request.QueryResult.StorageService())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, "/storage_services")

}

// GET /storage_services/edit/:id
func StorageServiceEdit(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	c.HTML(http.StatusOK, "storage_service/form.html", request.TemplateData)
}

// GET /storage_services
func StorageServiceIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.StorageServices
	c.HTML(http.StatusOK, "storage_service/list.html", request.TemplateData)
}

// GET /storage_services/new
func StorageServiceNew(c *gin.Context) {
	ss := core.NewStorageService()
	data := gin.H{
		"form":                 ss.ToForm(),
		"suppressDeleteButton": true,
	}
	c.HTML(http.StatusOK, "storage_service/form.html", data)
}

// PUT /storage_services/edit/:id
// POST /storage_services/edit/:id
// POST /storage_services/new
func StorageServiceSave(c *gin.Context) {
	ss := &core.StorageService{}
	err := c.Bind(ss)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = core.ObjSave(ss)
	if err != nil {
		objectExistsInDB, _ := core.ObjExists(ss.ID)
		data := gin.H{
			"form":             ss.ToForm(),
			"objectExistsInDB": objectExistsInDB,
		}
		c.HTML(http.StatusOK, "storage_service/form.html", data)
		return
	}
	c.Redirect(http.StatusFound, "/storage_services")

}
