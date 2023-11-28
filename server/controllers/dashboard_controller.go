package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

// DashboardShow shows the DART dashboard. This is essentially
// DART's homepage.
//
// GET /
func DashboardShow(c *gin.Context) {
	// Loop through remote repos. For any repo we can reach,
	// pull back all usable reports. A report will simply
	// return a blob of HTML to display.
	c.HTML(http.StatusOK, "dashboard/show.html", gin.H{})
}

// DashboardGetReport returns a report from a remote repository.
// This is an AJAX call. We want the dashboard page to display
// first, and then we'll load and display these reports as they
// become available.
//
// Params are RemoteRepoID and ReportName.
//
// GET /dashboard/report
func DashboardGetReport(c *gin.Context) {
	reportName := c.Query("ReportName")
	remoteRepoID := c.Query("RemoteRepoID")
	html, err := getRepoReport(remoteRepoID, reportName)
	result := "ok"
	status := http.StatusOK
	errMsg := ""
	if err != nil {
		result = "error"
		status = http.StatusBadRequest
		errMsg = err.Error()
	}
	data := gin.H{
		"result": result,
		"error":  errMsg,
		"html":   html,
	}
	c.JSON(status, data)
}

func getRepoReport(remoteRepoID, reportName string) (string, error) {
	result := core.ObjFind(remoteRepoID)
	if result.Error != nil {
		return "", result.Error
	}
	repo := result.RemoteRepository()
	client, err := core.GetRemoteRepoClient(repo)
	if err != nil {
		return "", err
	}
	return client.RunHTMLReport(reportName)
}
