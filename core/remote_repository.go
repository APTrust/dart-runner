package core

import (
	"fmt"
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

// RemoteRepositoryFind returns the RemoteRepository with the specified UUID,
// or sql.ErrNoRows if no matching record exists.
func RemoteRepositoryFind(uuid string) (*RemoteRepository, error) {
	result, err := ObjFind(uuid)
	if err != nil {
		return nil, err
	}
	return result.RemoteRepository, err
}

// RemoteRepositoryList returns a list of RemoteRepositorys with the specified
// order, offset and limit.
func RemoteRepositoryList(orderBy string, limit, offset int) ([]*RemoteRepository, error) {
	result, err := ObjList(constants.TypeRemoteRepository, orderBy, limit, offset)
	if err != nil {
		return nil, err
	}
	return result.RemoteRepositories, err
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

	form.AddField("ID", "ID", repo.ID, true)
	form.AddField("Name", "Name", repo.Name, true)
	form.AddField("Url", "URL", repo.Url, true)
	form.AddField("UserID", "User", repo.UserID, false)
	form.AddField("APIToken", "API Token", repo.APIToken, false)
	form.AddField("LoginExtra", "Login Extra", repo.LoginExtra, false)

	//pluginIdField := form.AddField("PluginID", "Plugin ID", repo.PluginID, false)
	//pluginIdField.AddChoice("", "")
	//pluginIdField.AddChoice("APTrustClient", PluginIdAPTrustClient)

	return form
}

func (repo *RemoteRepository) GetErrors() map[string]string {
	return repo.Errors
}

func (repo *RemoteRepository) IsDeletable() bool {
	return true
}
