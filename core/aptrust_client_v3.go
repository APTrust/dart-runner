package core

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	apt_network "github.com/APTrust/preservation-services/network"
)

func init() {
	RegisterRepoClient(constants.PluginNameAPTrustClientv3, constants.PluginIdAPTrustClientv3, RemoteRepoClientAPTrust)
}

// APTrustClientV3 is a remote repository client that can talk to
// the APTrust Registry API.
type APTrustClientV3 struct {
	id                   string
	name                 string
	description          string
	version              string
	availableHTMLReports []util.NameValuePair
	config               *RemoteRepository
}

// NewAPTrustClientV3 returns a new instance of the APTrust remote repo client.
func NewAPTrustClientV3(config *RemoteRepository) *APTrustClientV3 {
	return &APTrustClientV3{
		id:          constants.PluginIdAPTrustClientv3,
		name:        constants.PluginNameAPTrustClientv3,
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
	return NewAPTrustClientV3(config)
}

// ID returns this client's UUID.
func (client *APTrustClientV3) ID() string {
	return client.id
}

// Name returns the client name.
func (client *APTrustClientV3) Name() string {
	return client.name
}

// Description returns a short description of this client.
func (client *APTrustClientV3) Description() string {
	return client.description
}

// APIVersion returns the version number of the API that this client
// can talk to.
func (client *APTrustClientV3) APIVersion() string {
	return client.version
}

// AvailableHTMLReports returns a list of available HTML report names.
// In the NameValuePair, Name is the name of the report and Value is
// a description.
func (client *APTrustClientV3) AvailableHTMLReports() []util.NameValuePair {
	return client.availableHTMLReports
}

// TestConnection tests a connection to the remote repo. It returns true
// or false to describe whether the connection succeeded. Check the error
// if the connection did not succeed.
func (client *APTrustClientV3) TestConnection() error {
	registryClient, err := apt_network.NewRegistryClient(
		client.config.Url,
		client.version,
		client.config.UserID,
		client.config.APIToken,
		Dart.Log,
	)
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Add("per_page", "1")
	resp := registryClient.WorkItemList(params)
	if resp.Response.StatusCode == http.StatusUnauthorized || resp.Response.StatusCode == http.StatusForbidden {
		return fmt.Errorf("Server returned status %d. Be sure your user id and API token are correct.", resp.Response.StatusCode)
	}
	// Other errors should be OK here. They indicate that we did successfully authenticate.
	return nil
}

// RunHTMLReport runs the named report and returns HTML suitable for
// display on the DART dashboard. For a list of available report names,
// call AvailableHTMLReports().
func (client *APTrustClientV3) RunHTMLReport(name string) (string, error) {

	return "", nil
}
