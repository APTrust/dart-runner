package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GET /jobs/run/:id
func JobRunShow(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	job.UpdatePayloadStats()

	p := message.NewPrinter(language.English)
	byteCount := p.Sprintf("%d", job.ByteCount)

	data := gin.H{
		"job":           job,
		"byteCount":     byteCount,
		"pathSeparator": string(os.PathSeparator),
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

	messageChannel := make(chan *core.EventMessage)
	go func() {

		// TODO: Close message channel only after ALL parts of job (including ALL uploads) complete.

		//defer close(messageChannel)
		exitCode := core.RunJobWithMessageChannel(job, false, messageChannel)
		//c.SSEvent("message", fmt.Sprintf("Exit code = %d", returnCode))
		status := constants.StatusFailed
		if exitCode == constants.ExitOK {
			status = constants.StatusSuccess
		}
		eventMessage := &core.EventMessage{
			EventType: constants.EventTypeDisconnect,
			Message:   fmt.Sprintf("Job completed with exit code %d", exitCode),
			Status:    status,
		}
		c.SSEvent("message", eventMessage)
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
}
