package controllers

import (
	"fmt"
	"io"
	"net/http"
	"time"

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

// GET /jobs/run/:id
//
// By REST standards, this should be a POST. However, the Server
// Side Events standard for JavaScript only supports GET, so GET
// it is.
func JobRunExecute(c *gin.Context) {
	// Run the job in response to user clicking the Run button.
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()

	// TODO: Emit specific message types, such as
	//
	// - package/validate/upload started & completed
	// - package - file added
	// - validate - file checksum verified
	// - upload - bytes written

	messageChannel := make(chan string)
	go func() {
		defer close(messageChannel)
		returnCode := core.RunJobWithMessageChannel(job, false, messageChannel)
		c.SSEvent("message", fmt.Sprintf("Exit code = %d", returnCode))
		c.SSEvent("message", "EOF")
	}()

	streamer := func(w io.Writer) bool {
		if msg, ok := <-messageChannel; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	}

	// Building a small bag can take just milliseconds. In testing,
	// the front-end client (JavaScript EventSource) starts receiving data
	// in the millisecond window between connecting and defining event
	// handlers. That causes the front end to miss the first event.
	// We could handle this with Last-Event-ID, but that gets tricky
	// if we have to cache and re-request data. This is much simpler.
	// Just give the front-end time to attach its event handler.
	time.Sleep(200 * time.Millisecond)

	clientDisconnected := c.Stream(streamer)
	if clientDisconnected {
		// core.Dart.Log.Error("While running job, client disconnected in middle of stream.")
		fmt.Println("While running job, client disconnected in middle of stream.")
	}
	// returnCode := core.RunJob(job, false, false)
	// status := http.StatusOK
	// if returnCode != 0 {
	// 	status = http.StatusInternalServerError
	// }
	// c.JSON(status, job)
}

// func StreamJsonResponse(c *gin.Context, job *core.Job) bool {
// }
