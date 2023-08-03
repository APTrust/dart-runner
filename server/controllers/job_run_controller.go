package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// GET /jobs/run/:id
func JobRunShow(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	data := gin.H{
		"job": job,
	}
	c.HTML(http.StatusOK, "job/run.html", data)
}

// POST /jobs/run/:id
func JobRunExecute(c *gin.Context) {
	// Run the job in response to user clicking the Run button.
}
