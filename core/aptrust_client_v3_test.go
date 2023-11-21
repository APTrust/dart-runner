package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestNewAPTrustClientV3(t *testing.T) {
	repo := core.NewRemoteRepository()
	repo.PluginID = constants.PluginIdAPTrustClientv3

	client := core.NewAPTrustClientV3(repo)
	assert.Equal(t, constants.PluginIdAPTrustClientv3, client.ID())
	assert.Equal(t, constants.PluginNameAPTrustClientv3, client.Name())
	assert.Equal(t, "v3", client.APIVersion())
	assert.NotEmpty(t, client.Description())
	reports := client.AvailableHTMLReports()
	assert.NotEmpty(t, reports)
	assert.Equal(t, "Work Items", reports[0].Name)

	testAPTrustClientConnection(t, client)
	testAPTrustClientRunReport(t, client)
}

func testAPTrustClientConnection(t *testing.T, client *core.APTrustClientV3) {
	// TODO: Test both successful and failing connection
	// using local dummy service.
}

func testAPTrustClientRunReport(t *testing.T, client *core.APTrustClientV3) {
	// TODO: Test against dummy service? Or use a Docker container?
}
