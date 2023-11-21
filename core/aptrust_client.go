package core

import (
	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

const aptrustV3ClientName = "APTrust Registry Client (v3)"

func init() {
	RegisterRepoClient(aptrustV3ClientName, constants.PluginIdAPTrustClientv3, RemoteRepoClientAPTrust)
}

// APTrustClient is a remote repository client that can talk to
// the APTrust Registry API.
type APTrustClient struct {
	id                   string
	name                 string
	description          string
	version              string
	availableHTMLReports []util.NameValuePair
	config               *RemoteRepository
}

// NewAPTrustClient returns a new instance of the APTrust remote repo client.
func NewAPTrustClient(config *RemoteRepository) *APTrustClient {
	return &APTrustClient{
		id:          constants.PluginIdAPTrustClientv3,
		name:        aptrustV3ClientName,
		description: "This client talks to the APTrust Registry REST API.",
		version:     "v3",
		config:      config,
		availableHTMLReports: []util.NameValuePair{
			{Name: "Work Items", Value: "Returns a list of recent work items."},
		},
	}
}

// RemoteRepoClientAPTrust returns an APTrustlient as a basic RemoteRepoClient
// to support automated discovery and generation of clients using the factory
// method GetRemoteRepoClient().
//
// If you need access to APTrust client methods outside of the RemoteRepoClient
// interface, which is just for reporting, use NewLOCKSSClient() instead.
func RemoteRepoClientAPTrust(config *RemoteRepository) RemoteRepoClient {
	return NewAPTrustClient(config)
}

// ID returns this client's UUID.
func (client *APTrustClient) ID() string {
	return client.id
}

// Name returns the client name.
func (client *APTrustClient) Name() string {
	return client.name
}

// Description returns a short description of this client.
func (client *APTrustClient) Description() string {
	return client.description
}

// APIVersion returns the version number of the API that this client
// can talk to.
func (client *APTrustClient) APIVersion() string {
	return client.version
}

// AvailableHTMLReports returns a list of available HTML report names.
// In the NameValuePair, Name is the name of the report and Value is
// a description.
func (client *APTrustClient) AvailableHTMLReports() []util.NameValuePair {
	return client.availableHTMLReports
}

// TestConnection tests a connection to the remote repo. It returns true
// or false to describe whether the connection succeeded. Check the error
// if the connection did not succeed.
func (client *APTrustClient) TestConnection() (bool, error) {

	return true, nil
}

// RunHTMLReport runs the named report and returns HTML suitable for
// display on the DART dashboard. For a list of available report names,
// call AvailableHTMLReports().
func (client *APTrustClient) RunHTMLReport(name string) (string, error) {

	return "", nil
}
