package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// GET /remote_repositories/delete/:id
// POST /remote_repositories/delete/:id
func RemoteRepositoryDelete(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	err := core.ObjDelete(request.QueryResult.RemoteRepository())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, "/remote_repositories")

}

// GET /remote_repositories/edit/:id
func RemoteRepositoryEdit(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	c.HTML(http.StatusOK, "remote_repository/form.html", request.TemplateData)
}

// GET /remote_repositories
func RemoteRepositoryIndex(c *gin.Context) {
	request := NewRequest(c)
	if request.HasErrors() {
		c.AbortWithError(http.StatusInternalServerError, request.Errors[0])
		return
	}
	request.TemplateData["items"] = request.QueryResult.RemoteRepositories
	c.HTML(http.StatusOK, "remote_repository/list.html", request.TemplateData)
}

// GET /remote_repositories/new
func RemoteRepositoryNew(c *gin.Context) {
	repo := core.NewRemoteRepository()
	data := gin.H{
		"form":                 repo.ToForm(),
		"suppressDeleteButton": true,
	}
	c.HTML(http.StatusOK, "remote_repository/form.html", data)

}

// PUT /remote_repositories/edit/:id
// POST /remote_repositories/edit/:id
// POST /remote_repositories/new
func RemoteRepositorySave(c *gin.Context) {
	repo := &core.RemoteRepository{}
	err := c.Bind(repo)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = core.ObjSave(repo)
	if err != nil {
		objectExistsInDB, _ := core.ObjExists(repo.ID)
		data := gin.H{
			"form":             repo.ToForm(),
			"objectExistsInDB": objectExistsInDB,
		}
		c.HTML(http.StatusOK, "remote_repository/form.html", data)
		return
	}
	c.Redirect(http.StatusFound, "/remote_repositories")
}
