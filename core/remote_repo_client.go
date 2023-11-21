package core

import (
	"sort"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// repoClients is a map of available remote repository clients.
var repoClients []NameIDPair

// repoClientConstructors maps repo client ids to their constructors
var repoClientConstructors map[string]func(*RemoteRepository) RemoteRepoClient

// RemoteRepoClient interface defines the basics for talking
// to a remote repository. Each client talks to one specific type
// of remote repository, usually through a REST API. We don't
// want to know the internal workings of these clients. We just
// want to be able to get reports to display on the dashboard.
type RemoteRepoClient interface {
	ID() string
	Name() string
	APIVersion() string
	Description() string
	AvailableHTMLReports() []util.NameValuePair
	RunHTMLReport(string) (string, error)
	TestConnection() (bool, error)
}

// RegisterRepoClient registers a remote repo client, so that
// DART knows it's available.
func RegisterRepoClient(name, id string, constructor func(*RemoteRepository) RemoteRepoClient) {
	if repoClients == nil {
		repoClients = make([]NameIDPair, 0)
	}
	repoClients = append(repoClients, NameIDPair{Name: name, ID: id})
	if repoClientConstructors == nil {
		repoClientConstructors = make(map[string]func(*RemoteRepository) RemoteRepoClient)
	}
	repoClientConstructors[id] = constructor
}

// RepoClientList returns a list of NameIDPair objects, listing
// available remote repository clients. This is useful for
// creating HTML select lists.
func RepoClientList() []NameIDPair {
	sort.Slice(repoClients, func(a, b int) bool {
		return repoClients[a].Name < repoClients[b].Name
	})
	return repoClients
}

// GetRemoteRepoClient returns a client that can talk to a RemoteRepository.
func GetRemoteRepoClient(repo *RemoteRepository) (RemoteRepoClient, error) {
	constructor, ok := repoClientConstructors[repo.PluginID]
	if !ok {
		return nil, constants.ErrNoSuchClient
	}
	return constructor(repo), nil
}
