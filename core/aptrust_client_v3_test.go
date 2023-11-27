package core_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}

func TestAPTrustClientConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(registryWorkItemHandler))
	defer server.Close()

	repo := core.NewRemoteRepository()
	repo.PluginID = constants.PluginIdAPTrustClientv3
	repo.Url = server.URL

	repo.APIToken = "valid-token"
	client := core.NewAPTrustClientV3(repo)
	err := client.TestConnection()
	require.Nil(t, err)

	repo.APIToken = "invalid-token"
	client = core.NewAPTrustClientV3(repo)
	err = client.TestConnection()
	require.NotNil(t, err)
}

func TestAPTrustClientWorkItemReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(registryWorkItemHandler))
	defer server.Close()

	repo := core.NewRemoteRepository()
	repo.PluginID = constants.PluginIdAPTrustClientv3
	repo.Url = server.URL

	repo.APIToken = "valid-token"
	client := core.NewAPTrustClientV3(repo)
	html, err := client.RunHTMLReport("Work Items")
	require.Nil(t, err)
	require.NotNil(t, html)
	assert.Equal(t, expectedWorkItemHTML, html)
}

func TestAPTrustClientObjectReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(registryObjectHandler))
	defer server.Close()

	repo := core.NewRemoteRepository()
	repo.PluginID = constants.PluginIdAPTrustClientv3
	repo.Url = server.URL

	repo.APIToken = "valid-token"
	client := core.NewAPTrustClientV3(repo)
	html, err := client.RunHTMLReport("Recent Objects")
	require.Nil(t, err)
	require.NotNil(t, html)
	assert.Equal(t, expectedObjectListHTML, html)
}

func TestAPTrustClientUnknownReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(registryObjectHandler))
	defer server.Close()

	repo := core.NewRemoteRepository()
	repo.PluginID = constants.PluginIdAPTrustClientv3
	repo.Url = server.URL
	client := core.NewAPTrustClientV3(repo)
	html, err := client.RunHTMLReport("This report does not exist")
	assert.Error(t, err)
	assert.Empty(t, html)
}

// This handler will return unauthorized if it gets API token "invalid-token".
// Otherwise, it returns proper WorkItem list JSON.
// Use this one to TestConnection().
func registryWorkItemHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Pharos-API-Key")
	if token == "invalid-token" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"error": "bad auth token"}`)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, workItemResponseJson)
	}
}

// This handler returns IntellectualObject list JSON no matter what.
// Don't use this to TestConnection().
func registryObjectHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, objectResponseJson)
}

var workItemResponseJson = `
{
    "count": 933132,
    "next": "/member-api/v3/items?page=41\\u0026per_page=4",
    "previous": "/member-api/v3/items?page=39\\u0026per_page=4",
    "results": [
        {
            "id": 933127,
            "name": "KNOX_007565.tar",
            "etag": "d25fe92c730fe8c11bfe596626b32610-64",
            "institution_id": 8377,
            "institution_name": "School of Hard Knocks",
            "institution_identifier": "knox.edu",
            "intellectual_object_id": 271291,
            "object_identifier": "knox.edu/KNOX_007565",
            "alt_identifier": "",
            "bag_group_identifier": "",
            "storage_option": "Glacier-Deep-OH",
            "bagit_profile_identifier": "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json",
            "source_organization": "School of Hard Knocks",
            "internal_sender_identifier": "",
            "generic_file_id": 0,
            "generic_file_identifier": "",
            "bucket": "aptrust.receiving.knox.edu",
            "user": "system@aptrust.org",
            "note": "Finished cleanup. Ingest complete.",
            "action": "Ingest",
            "stage": "Cleanup",
            "status": "Success",
            "outcome": "Ingest complete",
            "bag_date": "2023-11-21T22:04:59Z",
            "date_processed": "2023-11-21T22:05:33.885224Z",
            "retry": false,
            "node": "",
            "pid": 0,
            "needs_admin_review": false,
            "queued_at": "2023-11-21T22:05:34.063015Z",
            "size": 531548160,
            "stage_started_at": "2023-11-21T22:06:37.977518Z",
            "aptrust_approver": "",
            "inst_approver": "",
            "created_at": "2023-11-21T22:05:34.056343Z",
            "updated_at": "2023-11-21T22:06:38.792231Z"
        },
        {
            "id": 933126,
            "name": "KNOX_007564.tar",
            "etag": "164a8cadb66505bd5dc1cee8332e0e68-51",
            "institution_id": 8377,
            "institution_name": "School of Hard Knocks",
            "institution_identifier": "knox.edu",
            "intellectual_object_id": 271290,
            "object_identifier": "knox.edu/KNOX_007564",
            "alt_identifier": "",
            "bag_group_identifier": "",
            "storage_option": "Glacier-Deep-OH",
            "bagit_profile_identifier": "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json",
            "source_organization": "School of Hard Knocks",
            "internal_sender_identifier": "",
            "generic_file_id": 0,
            "generic_file_identifier": "",
            "bucket": "aptrust.receiving.knox.edu",
            "user": "system@aptrust.org",
            "note": "Finished cleanup. Ingest complete.",
            "action": "Ingest",
            "stage": "Cleanup",
            "status": "Success",
            "outcome": "Ingest complete",
            "bag_date": "2023-11-21T22:04:34Z",
            "date_processed": "2023-11-21T22:05:07.427847Z",
            "retry": false,
            "node": "",
            "pid": 0,
            "needs_admin_review": false,
            "queued_at": "2023-11-21T22:05:07.620111Z",
            "size": 422533120,
            "stage_started_at": "2023-11-21T22:06:08.889292Z",
            "aptrust_approver": "",
            "inst_approver": "",
            "created_at": "2023-11-21T22:05:07.613897Z",
            "updated_at": "2023-11-21T22:06:09.704304Z"
        },
        {
            "id": 933125,
            "name": "KNOX_007563.tar",
            "etag": "cabe07ed093fb9111a0da25adb460860-47",
            "institution_id": 8377,
            "institution_name": "School of Hard Knocks",
            "institution_identifier": "knox.edu",
            "intellectual_object_id": 271289,
            "object_identifier": "knox.edu/KNOX_007563",
            "alt_identifier": "",
            "bag_group_identifier": "",
            "storage_option": "Glacier-Deep-OH",
            "bagit_profile_identifier": "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json",
            "source_organization": "School of Hard Knocks",
            "internal_sender_identifier": "",
            "generic_file_id": 0,
            "generic_file_identifier": "",
            "bucket": "aptrust.receiving.knox.edu",
            "user": "system@aptrust.org",
            "note": "Finished cleanup. Ingest complete.",
            "action": "Ingest",
            "stage": "Cleanup",
            "status": "Success",
            "outcome": "Ingest complete",
            "bag_date": "2023-11-21T22:04:10Z",
            "date_processed": "2023-11-21T22:04:40.47526Z",
            "retry": false,
            "node": "",
            "pid": 0,
            "needs_admin_review": false,
            "queued_at": "2023-11-21T22:04:40.593895Z",
            "size": 393338880,
            "stage_started_at": "2023-11-21T22:05:36.954567Z",
            "aptrust_approver": "",
            "inst_approver": "",
            "created_at": "2023-11-21T22:04:40.586934Z",
            "updated_at": "2023-11-21T22:05:37.862091Z"
        },
        {
            "id": 933124,
            "name": "KNOX_007562.tar",
            "etag": "aa1d92f7ac4fd1a099d172b9105bce60-48",
            "institution_id": 8377,
            "institution_name": "School of Hard Knocks",
            "institution_identifier": "knox.edu",
            "intellectual_object_id": 271288,
            "object_identifier": "knox.edu/KNOX_007562",
            "alt_identifier": "",
            "bag_group_identifier": "",
            "storage_option": "Glacier-Deep-OH",
            "bagit_profile_identifier": "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json",
            "source_organization": "School of Hard Knocks",
            "internal_sender_identifier": "",
            "generic_file_id": 0,
            "generic_file_identifier": "",
            "bucket": "aptrust.receiving.knox.edu",
            "user": "system@aptrust.org",
            "note": "Finished cleanup. Ingest complete.",
            "action": "Ingest",
            "stage": "Cleanup",
            "status": "Success",
            "outcome": "Ingest complete",
            "bag_date": "2023-11-21T22:03:46Z",
            "date_processed": "2023-11-21T22:04:15.197843Z",
            "retry": false,
            "node": "",
            "pid": 0,
            "needs_admin_review": false,
            "queued_at": "0001-01-01T00:00:00Z",
            "size": 402114560,
            "stage_started_at": "2023-11-21T22:05:25.012515Z",
            "aptrust_approver": "",
            "inst_approver": "",
            "created_at": "2023-11-21T22:04:15.332132Z",
            "updated_at": "2023-11-21T22:05:25.817518Z"
        }
    ]
}`

var objectResponseJson = `
{
    "count": 271402,
    "next": "/member-api/v3/objects?page=41\\u0026per_page=4",
    "previous": "/member-api/v3/objects?page=39\\u0026per_page=4",
    "results": [
        {
            "id": 271255,
            "title": "KNOX_000089",
            "description": "School of Hard Knocks",
            "identifier": "knox.edu/KNOX_000089",
            "alt_identifier": "",
            "access": "institution",
            "bag_name": "KNOX_000089",
            "institution_id": 8377,
            "created_at": "2023-11-21T16:56:24.50977Z",
            "updated_at": "2023-11-21T16:56:24.50977Z",
            "state": "A",
            "etag": "5211fa940f07806a57b9c0cbfdebdab5-811",
            "bag_group_identifier": "",
            "storage_option": "Glacier-Deep-OH",
            "bagit_profile_identifier": "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json",
            "source_organization": "School of Hard Knocks",
            "internal_sender_identifier": "",
            "internal_sender_description": "",
            "institution_name": "School of Hard Knocks",
            "institution_identifier": "knox.edu",
            "institution_type": "MemberInstitution",
            "institution_parent_id": 0,
            "file_count": 66,
            "size": 6800442726,
            "payload_file_count": 64,
            "payload_size": 6800442183
        },
        {
            "id": 271254,
            "title": "KNOX_000088",
            "description": "School of Hard Knocks",
            "identifier": "knox.edu/KNOX_000088",
            "alt_identifier": "",
            "access": "institution",
            "bag_name": "KNOX_000088",
            "institution_id": 8377,
            "created_at": "2023-11-21T16:50:28.157142Z",
            "updated_at": "2023-11-21T16:50:28.157142Z",
            "state": "A",
            "etag": "3c17d45cbf2887dd7066c077c0ed1655-814",
            "bag_group_identifier": "",
            "storage_option": "Glacier-Deep-OH",
            "bagit_profile_identifier": "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json",
            "source_organization": "School of Hard Knocks",
            "internal_sender_identifier": "",
            "internal_sender_description": "",
            "institution_name": "School of Hard Knocks",
            "institution_identifier": "knox.edu",
            "institution_type": "MemberInstitution",
            "institution_parent_id": 0,
            "file_count": 66,
            "size": 6827509158,
            "payload_file_count": 64,
            "payload_size": 6827508615
        },
        {
            "id": 271253,
            "title": "KNOX_000080",
            "description": "School of Hard Knocks",
            "identifier": "knox.edu/KNOX_000080",
            "alt_identifier": "",
            "access": "institution",
            "bag_name": "KNOX_000080",
            "institution_id": 8377,
            "created_at": "2023-11-21T16:37:51.085551Z",
            "updated_at": "2023-11-21T16:37:51.085551Z",
            "state": "A",
            "etag": "95a6fc79a8c3329451a6960a9e65a18e-525",
            "bag_group_identifier": "",
            "storage_option": "Glacier-Deep-OH",
            "bagit_profile_identifier": "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json",
            "source_organization": "School of Hard Knocks",
            "internal_sender_identifier": "",
            "internal_sender_description": "",
            "institution_name": "School of Hard Knocks",
            "institution_identifier": "knox.edu",
            "institution_type": "MemberInstitution",
            "institution_parent_id": 0,
            "file_count": 44,
            "size": 4401550565,
            "payload_file_count": 42,
            "payload_size": 4401550022
        },
        {
            "id": 271252,
            "title": "KNOX_000087",
            "description": "School of Hard Knocks",
            "identifier": "knox.edu/KNOX_000087",
            "alt_identifier": "",
            "access": "institution",
            "bag_name": "KNOX_000087",
            "institution_id": 8377,
            "created_at": "2023-11-21T16:37:29.988296Z",
            "updated_at": "2023-11-21T16:37:29.988296Z",
            "state": "A",
            "etag": "ca3635d9b8a1949970b9d23161440abf-326",
            "bag_group_identifier": "",
            "storage_option": "Glacier-Deep-OH",
            "bagit_profile_identifier": "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json",
            "source_organization": "School of Hard Knocks",
            "internal_sender_identifier": "",
            "internal_sender_description": "",
            "institution_name": "School of Hard Knocks",
            "institution_identifier": "knox.edu",
            "institution_type": "MemberInstitution",
            "institution_parent_id": 0,
            "file_count": 58,
            "size": 2728041490,
            "payload_file_count": 56,
            "payload_size": 2728040947
        }
    ]
}`

var expectedWorkItemHTML = `
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
    
    
    <tr>
      <td><a href="http://127.0.0.1/work_items/show/933127" target="_blank">KNOX_007565.tar</a></td>
      <td>Cleanup</td>
      <td>Success</td>
    </tr>
    
    <tr>
      <td><a href="http://127.0.0.1/work_items/show/933126" target="_blank">KNOX_007564.tar</a></td>
      <td>Cleanup</td>
      <td>Success</td>
    </tr>
    
    <tr>
      <td><a href="http://127.0.0.1/work_items/show/933125" target="_blank">KNOX_007563.tar</a></td>
      <td>Cleanup</td>
      <td>Success</td>
    </tr>
    
    <tr>
      <td><a href="http://127.0.0.1/work_items/show/933124" target="_blank">KNOX_007562.tar</a></td>
      <td>Cleanup</td>
      <td>Success</td>
    </tr>
    
  </tbody>
</table>
`
var expectedObjectListHTML = `
<h3>Recently Ingested Objects</h3>
<table class="table table-hover">
  <thead class="thead-inverse">
    <tr>
      <th>Identifier</th>
      <th>Storage Option</th>
    </tr>
  </thead>
  <tbody>
    
    
    <tr>
      <td><a href="http://127.0.0.1/objects/show/271255" target="_blank">knox.edu/KNOX_000089</a></td>
      <td>Glacier-Deep-OH</td>
    </tr>
    
    <tr>
      <td><a href="http://127.0.0.1/objects/show/271254" target="_blank">knox.edu/KNOX_000088</a></td>
      <td>Glacier-Deep-OH</td>
    </tr>
    
    <tr>
      <td><a href="http://127.0.0.1/objects/show/271253" target="_blank">knox.edu/KNOX_000080</a></td>
      <td>Glacier-Deep-OH</td>
    </tr>
    
    <tr>
      <td><a href="http://127.0.0.1/objects/show/271252" target="_blank">knox.edu/KNOX_000087</a></td>
      <td>Glacier-Deep-OH</td>
    </tr>
    
  </tbody>
</table>
`
