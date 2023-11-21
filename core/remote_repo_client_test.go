package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteRepoClient(t *testing.T) {

	// Note that each repo client includes an init() function
	// that registers the client. On startup, we should have
	// two registered clients, APTrust and LOCKSS. If
	// RepoClientList shows both, we know that RegisterRepoClient
	// is working.
	//
	// Note that if we add more clients, we'll need to change this test.

	clientList := core.RepoClientList()
	require.Equal(t, 2, len(clientList))

	// Make sure list is sorted. Even though we added APTrust client last,
	// it should appear first. This list displays in HTML select controls,
	// so it's handy to have it sorted.
	assert.Equal(t, constants.PluginNameAPTrustClientv3, clientList[0].Name)
	assert.Equal(t, constants.PluginNameLOCKSSClientv2, clientList[1].Name)

	// Now make sure we can get an instance of each client.
	// Note that these clients won't be usable without real
	// credentials
	aptrustRepo := core.NewRemoteRepository()
	aptrustRepo.PluginID = constants.PluginIdAPTrustClientv3
	aptrustClient, err := core.GetRemoteRepoClient(aptrustRepo)
	require.Nil(t, err)
	require.NotNil(t, aptrustClient)
	assert.Equal(t, constants.PluginIdAPTrustClientv3, aptrustClient.ID())
	assert.Equal(t, constants.PluginNameAPTrustClientv3, aptrustClient.Name())

	lockssRepo := core.NewRemoteRepository()
	lockssRepo.PluginID = constants.PluginIdLOCKSSClientv2
	lockssClient, err := core.GetRemoteRepoClient(lockssRepo)
	require.Nil(t, err)
	require.NotNil(t, lockssClient)
	assert.Equal(t, constants.PluginIdLOCKSSClientv2, lockssClient.ID())
	assert.Equal(t, constants.PluginNameLOCKSSClientv2, lockssClient.Name())

	// Finally, let's register one more client and see if it appears in the list
	core.RegisterRepoClient("AARDVARK Sorts First", constants.EmptyUUID, dummyClientConstructor)
	clientList = core.RepoClientList()
	require.Equal(t, 3, len(clientList))
	assert.Equal(t, "AARDVARK Sorts First", clientList[0].Name)
	assert.Equal(t, constants.EmptyUUID, clientList[0].ID)
}

func dummyClientConstructor(repo *core.RemoteRepository) core.RemoteRepoClient {
	return nil
}
