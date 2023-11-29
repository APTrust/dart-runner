package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
)

// RemoteRepository contains config settings describing how to
// connect to a remote repository, such as APTrust. Presumably,
// this is a repository into which you are ingesting data,
// and the repository has a REST API.
//
// The repo config allows you to connect to the repo so you can
// see the state of bags you uploaded. The logic for performing
// those requests and parsing the responses has to be implemented
// elsewhere. In DART 2.x, this was done with plugins, and APTrust
// was the only existing plugin. In DART 3.x, the way to add new
// repo implementations is to be determined. One suggestion is to
// generate clients with Swagger/OpenAPI.
type RemoteRepository struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Url        string            `json:"url"`
	UserID     string            `json:"userId"`
	APIToken   string            `json:"apiToken"`
	LoginExtra string            `json:"loginExtra"`
	PluginID   string            `json:"pluginId"`
	Errors     map[string]string `json:"errors"`
}

func NewRemoteRepository() *RemoteRepository {
	return &RemoteRepository{
		ID: uuid.NewString(),
	}
}

// GetUserID returns the UserID for logging into this remote repo.
// If the UserID comes from an environment variable, this returns
// the value of the ENV var. Otherwise, it returns the value of
// this object's UserID property.
//
// When authenticating with a remote repo, call this instead of
// accessing UserID directly.
func (repo *RemoteRepository) GetUserID() string {
	if strings.HasPrefix(repo.UserID, "env:") {
		parts := strings.SplitN(repo.UserID, ":", 2)
		userID := os.Getenv(parts[1])
		if userID == "" {
			Dart.Log.Warningf("UserID for repo '%s' is set to env var '%s', but the env var has no value", repo.Name, parts[1])
		}
		return userID
	}
	return repo.UserID
}

// GetUserAPIToken returns the API token for logging into this remote repo.
// If the token comes from an environment variable, this returns
// the value of the ENV var. Otherwise, it returns the value of
// this object's APIToken property.
//
// When authenticating with a remote repo, call this instead of
// accessing APIToken directly.
func (repo *RemoteRepository) GetAPIToken() string {
	if strings.HasPrefix(repo.APIToken, "env:") {
		parts := strings.SplitN(repo.APIToken, ":", 2)
		token := os.Getenv(parts[1])
		if token == "" {
			Dart.Log.Warningf("API token for repo '%s' is set to env var '%s', but the env var has no value", repo.Name, parts[1])
		}
		return token
	}
	return repo.APIToken
}

// HasPlaintextAPIToken returns true if this repo's API token
// is non-empty and does not come from an environment variable.
func (repo *RemoteRepository) HasPlaintextAPIToken() bool {
	token := strings.TrimSpace(repo.APIToken)
	return token != "" && !strings.HasPrefix(repo.APIToken, "env:")
}

// ObjID returns this remote repo's UUID.
func (repo *RemoteRepository) ObjID() string {
	return repo.ID
}

// ObjName returns the name of this remote repo.
func (repo *RemoteRepository) ObjName() string {
	return repo.Name
}

// ObjType returns this object's type.
func (repo *RemoteRepository) ObjType() string {
	return constants.TypeRemoteRepository
}

func (repo *RemoteRepository) String() string {
	return fmt.Sprintf("RemoteRepository: '%s'", repo.Name)
}

// Validate returns true if this RemoteRepository config contains
// valid settings, false if not. Check the Errors map if this returns
// false.
func (repo *RemoteRepository) Validate() bool {
	repo.Errors = make(map[string]string)
	if !util.LooksLikeUUID(repo.ID) {
		repo.Errors["ID"] = "ID must be a valid uuid."
	}
	if strings.TrimSpace(repo.Name) == "" {
		repo.Errors["Name"] = "Please enter a name."
	}
	if !util.LooksLikeHypertextURL(repo.Url) {
		repo.Errors["Url"] = "Repository URL must be a valid URL beginning with http:// or https://."
	}
	return len(repo.Errors) == 0
}

func (repo *RemoteRepository) ToForm() *Form {
	form := NewForm(constants.TypeRemoteRepository, repo.ID, repo.Errors)
	form.UserCanDelete = true

	form.AddField("ID", "ID", repo.ID, true)
	form.AddField("Name", "Name", repo.Name, true)
	form.AddField("Url", "URL", repo.Url, true)
	form.AddField("UserID", "User", repo.UserID, false)
	form.AddField("APIToken", "API Token", repo.APIToken, false)
	form.AddField("LoginExtra", "Login Extra", repo.LoginExtra, false)

	pluginIdField := form.AddField("PluginID", "Client Type", repo.PluginID, false)
	pluginIdField.Choices = MakeChoiceListFromPairs(RepoClientList(), repo.PluginID)

	for field, errMsg := range repo.Errors {
		form.Fields[field].Error = errMsg
	}

	return form
}

func (repo *RemoteRepository) GetErrors() map[string]string {
	return repo.Errors
}

func (repo *RemoteRepository) IsDeletable() bool {
	return true
}

func (repo *RemoteRepository) TestConnection() error {
	if !util.LooksLikeUUID(repo.PluginID) {
		return fmt.Errorf("Please choose a client plugin before testing connection.")
	}
	client, err := GetRemoteRepoClient(repo)
	if err != nil {
		Dart.Log.Errorf("Can't get client for repo '%s': %v", repo.Name, err)
		return err
	}
	err = client.TestConnection()
	if err != nil {
		Dart.Log.Errorf("Test connection failed for repo '%s': %v", repo.Name, err)
	} else {
		Dart.Log.Infof("Test connection succeeded for repo %s", repo.Name)
	}
	return err
}

// ReportsAvailable returns a list of reports that can be retrieved
// from this remote repository.
func (repo *RemoteRepository) ReportsAvailable() ([]util.NameValuePair, error) {
	reports := make([]util.NameValuePair, 0)
	if !util.LooksLikeUUID(repo.PluginID) {
		return reports, fmt.Errorf("no client exists for this repo because plugin id is not a uuid")
	}
	client, err := GetRemoteRepoClient(repo)
	if err != nil {
		return reports, err
	}
	if client == nil {
		return reports, fmt.Errorf("client for repo %s is nil", repo.Name)
	}
	return client.AvailableHTMLReports(), nil
}
