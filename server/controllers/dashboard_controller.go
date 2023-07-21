package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func DashboardShow(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard/show.html", gin.H{})
}
