package core

import (
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

func init() {

	// Don't register this for the alpa.
	// We don't want it to appear in the client list because
	// it's not ready yet. If that changes, we can uncomment
	// the line below.

	// RegisterRepoClient(constants.PluginNameLOCKSSClientv2, constants.PluginIdLOCKSSClientv2, RemoteRepoClientLOCKSS)
}

// LOCKSSClientV2 is a remote repository client that can talk to
// the LOCKSS API. This client satisfies the RemoteRepoClient
// interface and includes additional methods for depositing
// materials into LOCKSS repos.
type LOCKSSClientV2 struct {
	id                   string
	name                 string
	description          string
	version              string
	availableHTMLReports []util.NameValuePair
	config               *RemoteRepository
}

// NewLOCKSSClientV2 returns a new instance of the LOCKSS remote repo client.
func NewLOCKSSClientV2(config *RemoteRepository) *LOCKSSClientV2 {
	return &LOCKSSClientV2{
		id:          constants.PluginIdLOCKSSClientv2,
		name:        constants.PluginNameLOCKSSClientv2,
		description: "This client talks to the LOCKSS REST API.",
		version:     "v2",
		config:      config,
		availableHTMLReports: []util.NameValuePair{
			{Name: "TBD", Value: "To be determined..."},
		},
	}
}

// RemoteRepoClientLOCKSS returns a LOCKSSClient as a basic RemoteRepoClient
// to support automated discovery and generation of clients using the factory
// method GetRemoteRepoClient().
//
// If you need access to LOCKSS client methods outside of the RemoteRepoClient
// interface, which is just for reporting, use NewLOCKSSClient() instead.
func RemoteRepoClientLOCKSS(config *RemoteRepository) RemoteRepoClient {
	return NewLOCKSSClientV2(config)
}

// ID returns this client's UUID.
func (client *LOCKSSClientV2) ID() string {
	return client.id
}

// Name returns the client name.
func (client *LOCKSSClientV2) Name() string {
	return client.name
}

// Description returns a short description of this client.
func (client *LOCKSSClientV2) Description() string {
	return client.description
}

// APIVersion returns the version number of the API that this client
// can talk to.
func (client *LOCKSSClientV2) APIVersion() string {
	return client.version
}

// AvailableHTMLReports returns a list of available HTML report names.
// In the NameValuePair, Name is the name of the report and Value is
// a description.
func (client *LOCKSSClientV2) AvailableHTMLReports() []util.NameValuePair {
	return client.availableHTMLReports
}

// TestConnection tests a connection to the remote repo. It returns true
// or false to describe whether the connection succeeded. Check the error
// if the connection did not succeed.
func (client *LOCKSSClientV2) TestConnection() error {

	return nil
}

// RunHTMLReport runs the named report and returns HTML suitable for
// display on the DART dashboard. For a list of available report names,
// call AvailableHTMLReports().
func (client *LOCKSSClientV2) RunHTMLReport(name string) (string, error) {

	return "", nil
}

// TODO: Methods for creating objects, depositing to LOCKSS, etc.
