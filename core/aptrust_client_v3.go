package core

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	apt_network "github.com/APTrust/preservation-services/network"
	"github.com/gin-gonic/gin"
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
	registry             *apt_network.RegistryClient
}

// NewAPTrustClientV3 returns a new instance of the APTrust remote repo client.
func NewAPTrustClientV3(repo *RemoteRepository) *APTrustClientV3 {
	return &APTrustClientV3{
		id:          constants.PluginIdAPTrustClientv3,
		name:        constants.PluginNameAPTrustClientv3,
		description: "This client talks to the APTrust Registry REST API.",
		version:     "v3",
		config:      repo,
		availableHTMLReports: []util.NameValuePair{
			{Name: "Work Items", Value: "Returns a list of recent work items."},
			{Name: "Recent Objects", Value: "Returns a list of recently ingested intellectual objects."},
		},
	}
}

// RemoteRepoClientAPTrust returns an APTrustlient as a basic RemoteRepoClient
// to support automated discovery and generation of clients using the factory
// method GetRemoteRepoClient().
//
// If you need access to APTrust client methods outside of the RemoteRepoClient
// interface, which is just for reporting, use NewLOCKSSClient() instead.
func RemoteRepoClientAPTrust(repo *RemoteRepository) RemoteRepoClient {
	return NewAPTrustClientV3(repo)
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
	err := client.connect()
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Add("per_page", "1")
	resp := client.registry.WorkItemList(params)
	if resp.Response.StatusCode == http.StatusUnauthorized || resp.Response.StatusCode == http.StatusForbidden {
		return constants.ErrRepoUnauthorized
	}
	// Other errors should be OK here. They indicate that we did successfully authenticate.
	return nil
}

// RunHTMLReport runs the named report and returns HTML suitable for
// display on the DART dashboard. For a list of available report names,
// call AvailableHTMLReports().
func (client *APTrustClientV3) RunHTMLReport(name string) (string, error) {
	switch name {
	case "Work Items":
		return client.runWorkItemReport()
	case "Recent Objects":
		return client.runRecentObjectReport()
	default:
		return "", constants.ErrUnknownReport
	}
}

func (client *APTrustClientV3) runWorkItemReport() (string, error) {
	err := client.connect()
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Add("per_page", "30")
	resp := client.registry.WorkItemList(params)
	if resp.Response.StatusCode == http.StatusUnauthorized || resp.Response.StatusCode == http.StatusForbidden {
		return "", constants.ErrRepoUnauthorized
	}
	if resp.Error != nil {
		return "", resp.Error
	}
	data := gin.H{
		"items":       resp.WorkItems(),
		"repoBaseUrl": client.repoBaseUrl(),
	}
	sb := &strings.Builder{}
	_template := template.Must(template.New("work_items_report").Parse(workItemsTemplate))
	err = _template.Execute(sb, data)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (client *APTrustClientV3) runRecentObjectReport() (string, error) {
	err := client.connect()
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Add("per_page", "30")
	resp := client.registry.IntellectualObjectList(params)
	if resp.Response.StatusCode == http.StatusUnauthorized || resp.Response.StatusCode == http.StatusForbidden {
		return "", constants.ErrRepoUnauthorized
	}
	if resp.Error != nil {
		return "", resp.Error
	}
	data := gin.H{
		"items":       resp.IntellectualObjects(),
		"repoBaseUrl": client.repoBaseUrl(),
	}
	sb := &strings.Builder{}
	_template := template.Must(template.New("intellectual_objects_report").Parse(intellectualObjectsTemplate))
	err = _template.Execute(sb, data)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (client *APTrustClientV3) connect() error {
	if client.registry != nil {
		return nil
	}
	if !client.isValidDomain() {
		return fmt.Errorf("Domain is not valid for APTrust repositories.")
	}
	registryClient, err := apt_network.NewRegistryClient(
		client.config.Url,
		client.version,
		client.config.UserID,
		client.config.APIToken,
		Dart.Log,
	)
	client.registry = registryClient
	return err
}

func (client *APTrustClientV3) isValidDomain() bool {
	parsedUrl, err := url.Parse(client.config.Url)
	if err != nil {
		return false
	}
	host := parsedUrl.Hostname()
	// Allow localhost for testing.
	isLocalHost := host == "localhost" || host == "127.0.0.1"
	// And these are the real APTrust repo domains.
	isAPTrustHost := host == "repo.aptrust.org" || host == "demo.aptrust.org" || host == "staging.aptrust.org"
	return isLocalHost || isAPTrustHost
}

func (client *APTrustClientV3) repoBaseUrl() string {
	parsedUrl, err := url.Parse(client.config.Url)
	// This shouldn't happen because we don't call this till
	// after we connect, and if we've connected, the URL can't be bad.
	if err != nil {
		Dart.Log.Errorf("APTrustClientV3: bad repo url %s", client.config.Url)
		return client.config.Url
	}
	host := parsedUrl.Hostname()

	// Allow this for local testing.
	if host == "localhost" || host == "127.0.0.1" {
		return fmt.Sprintf("http://%s", host)
	}
	return fmt.Sprintf("https://%s", host)
}

var workItemsTemplate = `
<h3>Recent Work Items</h3>
<table class="table table-hover">
  <thead class="thead-inverse">
    <tr>
      <th>Name</th>
      <th>Stage</th>
      <th>Status</th>
    </tr>
  </thead>
  <tbody>
    {{ $repoBaseUrl := .repoBaseUrl }}
    {{ range $index, $item := .items }}
    <tr>
      <td><a href="{{ $repoBaseUrl }}/work_items/show/{{ $item.ID }}" target="_blank">{{ $item.Name }}</a></td>
      <td>{{ $item.Stage }}</td>
      <td>{{ $item.Status }}</td>
    </tr>
    {{ end }}
  </tbody>
</table>
`

var intellectualObjectsTemplate = `
<h3>Recently Ingested Objects</h3>
<table class="table table-hover">
  <thead class="thead-inverse">
    <tr>
      <th>Identifier</th>
      <th>Storage Option</th>
    </tr>
  </thead>
  <tbody>
    {{ $repoBaseUrl := .repoBaseUrl }}
    {{ range $index, $item := .items }}
    <tr>
      <td><a href="{{ $repoBaseUrl }}/objects/show/{{ $item.ID }}" target="_blank">{{ $item.Identifier }}</a></td>
      <td>{{ $item.StorageOption }}</td>
    </tr>
    {{ end }}
  </tbody>
</table>
`
