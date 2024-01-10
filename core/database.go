package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/profiles"
	"github.com/APTrust/dart-runner/util"
)

type NameIDPair struct {
	Name string
	ID   string
}

func InitSchema() error {
	schema := `create table if not exists dart (
		uuid text primary key not null,
		obj_type text not null,
		obj_name text not null,
		is_deletable bool not null default false,
		obj_json text not null,
		updated_at datetime not null
	);
	create unique index if not exists ix_unique_object_name on dart(obj_type, obj_name);
	create table if not exists artifacts (
		uuid text primary key not null,
		job_id text not null,
		bag_name text not null,
		item_type text not null,
		file_name text,
		file_type text,
		raw_data text not null,
		updated_at datetime not null
	);
	create index if not exists ix_artifact_bag_name on artifacts(bag_name);
	create index if not exists ix_artifact_job_id on artifacts(job_id);
	`
	_, err := Dart.DB.Exec(schema)
	return err
}

func InitDBForFirstUse() {
	_, err := GetAppSetting("Bagging Directory")
	if err == sql.ErrNoRows {
		paths := util.NewPaths()
		dir := filepath.Join(paths.Documents, "DART")
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			Dart.Log.Errorf("Could not create Bagging Directory at %s: %v", dir, err)
		}
		setting := NewAppSetting("Bagging Directory", dir)
		err = ObjSave(setting)
		if err != nil {
			Dart.Log.Errorf("Could not save Bagging Directory setting: %v", err)
		}
	}
	result := ObjList(constants.TypeBagItProfile, "obj_name", 100, 0)
	if result.Error != nil {
		Dart.Log.Errorf("Could not get profiles list: ", result.Error)
		return
	}
	foundAPTrustProfile := false
	foundBTRProfile := false
	foundEmptyProfile := false
	for _, profile := range result.BagItProfiles {
		switch profile.BagItProfileInfo.BagItProfileIdentifier {
		case constants.BTRProfileIdentifier:
			foundBTRProfile = true
		case constants.DefaultProfileIdentifier:
			foundAPTrustProfile = true
		case constants.EmptyProfileIdentifier:
			foundEmptyProfile = true
		}
	}
	if !foundAPTrustProfile {
		initProfileInDB("APTrust", profiles.APTrust_V_2_2)
	}
	if !foundBTRProfile {
		initProfileInDB("BTR", profiles.BTR_V_1_0)
	}
	if !foundEmptyProfile {
		initProfileInDB("Empty", profiles.Empty_V_1_0)
	}
}

func initProfileInDB(name, profileJson string) {
	profile, err := BagItProfileFromJSON(profileJson)
	if err == nil {
		err = ObjSave(profile)
	}
	if err != nil {
		Dart.Log.Errorf("Error loading %s profile into local DB: %v", name, err)
	}
}

// ObjSave validates an object and then saves it if it passes validation.
// If the object is invalid, this will return constants.ErrObjecValidation
// and you can get a map of specific validation errors from obj.Errors.
//
// In rare cases, this may return constants.ErrUniqueConstraint, which means
// the database already contains a different object with the same UUID.
func ObjSave(obj PersistentObject) error {
	return objSave(obj, true)
}

// ObjSaveWithoutValidation saves an object without validating it first.
// This is useful when creating new Jobs through the UI, since the user
// must adjust settings on a number of screens to build up a fully valid
// Job.
func ObjSaveWithoutValidation(obj PersistentObject) error {
	return objSave(obj, false)
}

func objSave(obj PersistentObject, validate bool) error {
	if validate && !obj.Validate() {
		return constants.ErrObjecValidation
	}
	if FindConflictingUUID(obj) != "" {
		return constants.ErrUniqueConstraint
	}
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	stmt := `insert into dart (uuid, obj_type, obj_name, is_deletable, obj_json, updated_at) values (?,?,?,?,?,?)
	on conflict do update set obj_name=excluded.obj_name, is_deletable=excluded.is_deletable, 
	obj_json=excluded.obj_json, updated_at=excluded.updated_at where uuid=excluded.uuid`
	_, err = Dart.DB.Exec(stmt, obj.ObjID(), obj.ObjType(), obj.ObjName(), obj.IsDeletable(), string(jsonBytes), time.Now().UTC())
	return err
}

func ObjFind(uuid string) *QueryResult {
	row := Dart.DB.QueryRow("select obj_type, obj_json from dart where uuid=?", uuid)
	return findOne(row)
}

// ObjByNameAndType returns the object of the specified type with the matching
// name. Note that the DB has a unique constraint on obj_type + obj_name, so this
// should return at most one row.
func ObjByNameAndType(objName, objType string) *QueryResult {
	row := Dart.DB.QueryRow("select obj_type, obj_json from dart where obj_name=? and obj_type=?", objName, objType)
	return findOne(row)
}

func findOne(row *sql.Row) *QueryResult {
	var objType string
	var objJson string
	qr := NewQueryResult(constants.ResultTypeSingle)
	qr.ResultType = constants.ResultTypeSingle
	qr.Error = row.Scan(&objType, &objJson)
	if qr.Error != nil {
		return qr
	}
	qr.ObjType = objType
	qr.ObjCount = 1
	switch objType {
	case constants.TypeAppSetting:
		item := &AppSetting{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.AppSettings = append(qr.AppSettings, item)
	case constants.TypeExportSettings:
		item := &ExportSettings{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.ExportSettings = append(qr.ExportSettings, item)
	case constants.TypeBagItProfile:
		item := &BagItProfile{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.BagItProfiles = append(qr.BagItProfiles, item)
	case constants.TypeInternalSetting:
		item := &InternalSetting{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.InternalSettings = append(qr.InternalSettings, item)
	case constants.TypeJob:
		item := &Job{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.Jobs = append(qr.Jobs, item)
	case constants.TypeRemoteRepository:
		item := &RemoteRepository{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.RemoteRepositories = append(qr.RemoteRepositories, item)
	case constants.TypeStorageService:
		item := &StorageService{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.StorageServices = append(qr.StorageServices, item)
	case constants.TypeUploadJob:
		item := &UploadJob{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.UploadJobs = append(qr.UploadJobs, item)
	case constants.TypeValidationJob:
		item := &ValidationJob{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.ValidationJobs = append(qr.ValidationJobs, item)
	case constants.TypeWorkflow:
		item := &Workflow{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		item.resolveStorageServices() // see doc comments on this
		qr.Workflows = append(qr.Workflows, item)
	case constants.TypeWorkflowBatch:
		item := &WorkflowBatch{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.WorkflowBatches = append(qr.WorkflowBatches, item)
	default:
		qr.Error = constants.ErrUnknownType
	}
	return qr
}

func ObjList(objType, orderBy string, limit, offset int) *QueryResult {
	qr := NewQueryResult(constants.ResultTypeList)
	qr.ObjType = objType
	qr.ResultType = constants.ResultTypeList
	qr.Offset = offset
	qr.Limit = limit
	qr.OrderBy = orderBy
	qr.ObjCount, qr.Error = ObjCount(objType)
	if qr.Error != nil {
		return qr
	}

	// TODO: Make a proper whitelist of allowed order by clauses.
	// Right now, we only use two. We can't pass order by as query param.
	strQuery := "select obj_json from dart where obj_type = ? order by obj_name limit ? offset ?"
	if orderBy == "updated_at desc" {
		strQuery = "select obj_json from dart where obj_type = ? order by updated_at desc limit ? offset ?"
	}

	var rows *sql.Rows
	rows, qr.Error = Dart.DB.Query(strQuery, objType, limit, offset)
	if qr.Error != nil {
		return qr
	}
	defer rows.Close()

	switch objType {
	case constants.TypeAppSetting:
		appSettingList(rows, qr)
	case constants.TypeExportSettings:
		exportSettingList(rows, qr)
	case constants.TypeBagItProfile:
		bagItProfileList(rows, qr)
	case constants.TypeInternalSetting:
		internalSettingList(rows, qr)
	case constants.TypeJob:
		jobList(rows, qr)
	case constants.TypeRemoteRepository:
		remoteRepositoryList(rows, qr)
	case constants.TypeStorageService:
		storageServiceList(rows, qr)
	case constants.TypeUploadJob:
		uploadJobList(rows, qr)
	case constants.TypeValidationJob:
		validationJobList(rows, qr)
	case constants.TypeWorkflow:
		workflowList(rows, qr)
	case constants.TypeWorkflowBatch:
		workflowBatchList(rows, qr)
	default:
		qr.Error = constants.ErrUnknownType
	}
	return qr
}

func ObjCount(objType string) (int, error) {
	count := 0
	err := Dart.DB.QueryRow("select count(*) from dart where obj_type = ?", objType).Scan(&count)
	return count, err
}

func ObjExists(objId string) (bool, error) {
	count := 0
	err := Dart.DB.QueryRow("select count(*) from dart where uuid = ?", objId).Scan(&count)
	return count == 1, err
}

// ObjNameIdList returns a list of all items of type objType in the
// database. The return value is a slice of NameIDPair objects containing
// the name and id (uuid) of each object. The slice is in alpha order
// by name.
func ObjNameIdList(objType string) []NameIDPair {
	nameIdPairs := make([]NameIDPair, 0)
	var rows *sql.Rows
	rows, err := Dart.DB.Query("select uuid, obj_name from dart where obj_type = ? order by obj_name", objType)
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			uuid := ""
			name := ""
			err = rows.Scan(&uuid, &name)
			if err == nil {
				nameIdPairs = append(nameIdPairs, NameIDPair{Name: name, ID: uuid})
			}
		}
	}
	return nameIdPairs
}

// ObjChoiceList returns a list of all items of type objType in the
// database. The return value is a slice of Choice objects in which
// the Label is the object name and the Value is the id (uuid) of each
// object. The slice is in alpha order by name. Choices whose values
// match a value in selectedIds will have their Selected attribute
// set to true.
func ObjChoiceList(objType string, selectedIds []string) []Choice {
	nameIdPairs := ObjNameIdList(objType)
	choices := make([]Choice, len(nameIdPairs))
	for i, pair := range nameIdPairs {
		choices[i] = Choice{
			Label:    pair.Name,
			Value:    pair.ID,
			Selected: util.StringListContains(selectedIds, pair.ID),
		}
	}
	return choices
}

// GetAppSetting returns the value of the AppSetting with the given name.
func GetAppSetting(name string) (string, error) {
	jsonData := make([]byte, 0)
	query := "select obj_json from dart where obj_type=? and obj_name=?"
	err := Dart.DB.QueryRow(query, constants.TypeAppSetting, name).Scan(&jsonData)
	if err == sql.ErrNoRows {
		return "", err
	}
	setting := &AppSetting{}
	err = json.Unmarshal(jsonData, setting)
	return setting.Value, err
}

// FindConflictingUUID returns the UUID of the object having the same name and type
// as obj. The dart table has a unique constraint on obj_type + obj_name. That should
// prevent inserts that conflict with the constraint, but it doesn't. The modernc sqlite
// driver does not report an error on this conflict. It fails silently, So we check for
// the conflict on our own with this function.
//
// If this returns a UUID, there's a conflict, and we can't do the insert. If it returns
// an empty string, we're fine, and the insert can proceed.
func FindConflictingUUID(obj PersistentObject) string {
	uuid := ""
	query := `select uuid from dart where obj_type=? and obj_name=? and uuid != ?`
	err := Dart.DB.QueryRow(query, obj.ObjType(), obj.ObjName(), obj.ObjID()).Scan(&uuid)
	if err == sql.ErrNoRows {
		return ""
	}
	return uuid
}

func appSettingList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &AppSetting{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.AppSettings = append(qr.AppSettings, item)
	}
}

func exportSettingList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &ExportSettings{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.ExportSettings = append(qr.ExportSettings, item)
	}
}

func bagItProfileList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &BagItProfile{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.BagItProfiles = append(qr.BagItProfiles, item)
	}
}

func internalSettingList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &InternalSetting{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.InternalSettings = append(qr.InternalSettings, item)
	}
}

func jobList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &Job{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.Jobs = append(qr.Jobs, item)
	}
}

func remoteRepositoryList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &RemoteRepository{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.RemoteRepositories = append(qr.RemoteRepositories, item)
	}
}

func storageServiceList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &StorageService{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.StorageServices = append(qr.StorageServices, item)
	}
}

func uploadJobList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &UploadJob{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.UploadJobs = append(qr.UploadJobs, item)
	}
}

func validationJobList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &ValidationJob{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.ValidationJobs = append(qr.ValidationJobs, item)
	}
}

func workflowList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &Workflow{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		item.resolveStorageServices() // see doc comments on this
		qr.Workflows = append(qr.Workflows, item)
	}
}

func workflowBatchList(rows *sql.Rows, qr *QueryResult) {
	for rows.Next() {
		var jsonBytes []byte
		qr.Error = rows.Scan(&jsonBytes)
		if qr.Error != nil {
			return
		}
		item := &WorkflowBatch{}
		qr.Error = json.Unmarshal(jsonBytes, item)
		if qr.Error != nil {
			return
		}
		qr.WorkflowBatches = append(qr.WorkflowBatches, item)
	}
}

func ObjDelete(obj PersistentObject) error {
	if !obj.IsDeletable() {
		return constants.ErrNotDeletable
	}
	_, err := Dart.DB.Exec("delete from dart where uuid=? and obj_type=?", obj.ObjID(), obj.ObjType())
	return err
}

func ArtifactSave(a *Artifact) error {
	stmt := `insert into artifacts (uuid, job_id, bag_name, item_type, file_name, file_type, raw_data, updated_at) values (?,?,?,?,?,?,?,?)
	on conflict do update set bag_name=excluded.bag_name, item_type=excluded.item_type, 
	file_name=excluded.file_name, file_type=excluded.file_type, raw_data=excluded.raw_data, 
	updated_at=excluded.updated_at where uuid=excluded.uuid`
	_, err := Dart.DB.Exec(stmt, a.ID, a.JobID, a.BagName, a.ItemType, a.FileName, a.FileType, a.RawData, time.Now().UTC())
	return err
}

func ArtifactFind(uuid string) (*Artifact, error) {
	row := Dart.DB.QueryRow("select uuid, job_id, bag_name, item_type, file_name, file_type, raw_data, updated_at from artifacts where uuid=?", uuid)
	artifact := Artifact{}
	err := row.Scan(
		&artifact.ID,
		&artifact.JobID,
		&artifact.BagName,
		&artifact.ItemType,
		&artifact.FileName,
		&artifact.FileType,
		&artifact.RawData,
		&artifact.UpdatedAt,
	)
	return &artifact, err
}

func ArtifactNameIDList(jobID string) ([]NameIDPair, error) {
	nameIdPairs := make([]NameIDPair, 0)
	var rows *sql.Rows
	rows, err := Dart.DB.Query("select uuid, file_name, updated_at from artifacts where job_id = ? order by updated_at desc, file_name asc", jobID)
	// Jobs imported from DART v2 and jobs that have not run
	// will have no artifacts. That's fine. We'll just return
	// and empty list.
	if err != nil && err != sql.ErrNoRows {
		return nameIdPairs, err
	}
	defer rows.Close()
	for rows.Next() {
		uuid := ""
		name := ""
		var updatedAt time.Time
		err = rows.Scan(&uuid, &name, &updatedAt)
		if err == nil {
			itemName := fmt.Sprintf("%s -> %s", updatedAt.Format(time.DateTime), name)
			nameIdPairs = append(nameIdPairs, NameIDPair{Name: itemName, ID: uuid})
		}
	}
	return nameIdPairs, nil
}

func ArtifactListByJobID(jobID string) ([]*Artifact, error) {
	query := "select uuid, job_id, bag_name, item_type, file_name, file_type, raw_data, updated_at from artifacts where job_id=? order by updated_at desc, file_name asc"
	return artifactList(query, jobID)
}

func ArtifactListByJobName(bagName string) ([]*Artifact, error) {
	query := "select uuid, job_id, bag_name, item_type, file_name, file_type, raw_data, updated_at from artifacts where bag_name=? order by updated_at desc, file_name asc"
	return artifactList(query, bagName)
}

// ArtifactsDeleteByJobID deletes all artifacts associated with JobID.
func ArtifactsDeleteByJobID(jobID string) error {
	query := "delete from artifacts where job_id=?"
	_, err := Dart.DB.Exec(query, jobID)
	return err
}

func artifactList(query string, params ...interface{}) ([]*Artifact, error) {
	rows, err := Dart.DB.Query(query, params...)
	if err != nil {
		return nil, err
	}
	artifacts := make([]*Artifact, 0)
	for rows.Next() {
		artifact := Artifact{}
		err = rows.Scan(
			&artifact.ID,
			&artifact.JobID,
			&artifact.BagName,
			&artifact.ItemType,
			&artifact.FileName,
			&artifact.FileType,
			&artifact.RawData,
			&artifact.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, &artifact)
	}
	return artifacts, err
}

func ArtifactDelete(uuid string) error {
	_, err := Dart.DB.Exec("delete from artifacts where uuid=?", uuid)
	return err
}

// ClearDartTable is for testing use only
func ClearDartTable() error {
	if !util.TestsAreRunning() {
		return constants.ErrInvalidOperation
	}
	_, err := Dart.DB.Exec("delete from dart")
	return err
}

// ClearArtifactsTable is for testing use only
func ClearArtifactsTable() error {
	if !util.TestsAreRunning() {
		return constants.ErrInvalidOperation
	}
	_, err := Dart.DB.Exec("delete from artifacts")
	return err
}
