package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestNewLOCKSSClientV2(t *testing.T) {
	repo := core.NewRemoteRepository()
	repo.PluginID = constants.PluginIdLOCKSSClientv2

	client := core.NewLOCKSSClientV2(repo)
	assert.Equal(t, constants.PluginIdLOCKSSClientv2, client.ID())
	assert.Equal(t, constants.PluginNameLOCKSSClientv2, client.Name())
	assert.Equal(t, "v2", client.APIVersion())
	assert.NotEmpty(t, client.Description())
	reports := client.AvailableHTMLReports()
	assert.NotEmpty(t, reports)
	assert.Equal(t, "TBD", reports[0].Name) // This will change when we have real reports

	testLOCKSSClientConnection(t, client)
	testLOCKSSClientRunReport(t, client)
}

func testLOCKSSClientConnection(t *testing.T, client *core.LOCKSSClientV2) {
	// TODO: Test both successful and failing connection
	// using local dummy service.
}

func testLOCKSSClientRunReport(t *testing.T, client *core.LOCKSSClientV2) {
	// TODO: How to test this without LOCKSS service?
	// Or should we use a docker container?
}
