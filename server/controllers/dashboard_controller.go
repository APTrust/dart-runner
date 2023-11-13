package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func DashboardShow(c *gin.Context) {
	// Loop through remote repos. For any repo we can reach,
	// pull back all usable reports. A report will simply
	// return a blob of HTML to display.
	c.HTML(http.StatusOK, "dashboard/show.html", gin.H{})
}
