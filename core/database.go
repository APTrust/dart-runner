package core

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type Artifact struct {
	ID        string
	BagName   string
	ItemType  string // File or WorkResult
	FileName  string // name of manifest or tag file
	FileType  string // manifest or tag file
	RawData   string // file content or work result json
	UpdatedAt time.Time
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
		bag_name text not null,
		item_type text not null,
		file_name text,
		file_type text,
		raw_data text not null,
		updated_at datetime not null
	);
	create index if not exists ix_artifact_bag_name on artifacts(bag_name);
	`
	_, err := Dart.DB.Exec(schema)
	return err
}

func ObjSave(obj PersistentObject) error {
	if !obj.Validate() {
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
	var objType string
	var objJson string
	qr := NewQueryResult(constants.ResultTypeSingle)
	qr.ResultType = constants.ResultTypeSingle
	row := Dart.DB.QueryRow("select obj_type, obj_json from dart where uuid=?", uuid)
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
	case constants.TypeStorageService:
		item := &StorageService{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.StorageServices = append(qr.StorageServices, item)
	case constants.TypeRemoteRepository:
		item := &RemoteRepository{}
		qr.Error = json.Unmarshal([]byte(objJson), item)
		qr.RemoteRepositories = append(qr.RemoteRepositories, item)
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
	var rows *sql.Rows
	rows, qr.Error = Dart.DB.Query("select obj_json from dart where obj_type = ? order by ? limit ? offset ?", objType, orderBy, limit, offset)
	if qr.Error != nil {
		return qr
	}
	defer rows.Close()

	switch objType {
	case constants.TypeAppSetting:
		appSettingList(rows, qr)
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

func ObjDelete(obj PersistentObject) error {
	if !obj.IsDeletable() {
		return constants.ErrNotDeletable
	}
	_, err := Dart.DB.Exec("delete from dart where uuid=? and obj_type=?", obj.ObjID(), obj.ObjType())
	return err
}

func ArtifactSave(a *Artifact) error {
	stmt := `insert into artifacts (uuid, bag_name, item_type, file_name, file_type, raw_data, updated_at) values (?,?,?,?,?,?,?)
	on conflict do update set bag_name=excluded.bag_name, item_type=excluded.item_type, 
	file_name=excluded.file_name, file_type=excluded.file_type, raw_data=excluded.raw_data, 
	updated_at=excluded.updated_at where uuid=excluded.uuid`
	_, err := Dart.DB.Exec(stmt, a.ID, a.BagName, a.ItemType, a.FileName, a.FileType, a.RawData, time.Now().UTC())
	return err
}

func ArtifactFind(uuid string) (*Artifact, error) {
	row := Dart.DB.QueryRow("select uuid, bag_name, item_type, file_name, file_type, raw_data, updated_at from artifacts where uuid=?", uuid)
	artifact := Artifact{}
	err := row.Scan(
		&artifact.ID,
		&artifact.BagName,
		&artifact.ItemType,
		&artifact.FileName,
		&artifact.FileType,
		&artifact.RawData,
		&artifact.UpdatedAt,
	)
	return &artifact, err
}

func ArtifactList(bagName string) ([]*Artifact, error) {
	rows, err := Dart.DB.Query("select uuid, bag_name, item_type, file_name, file_type, raw_data, updated_at from artifacts where bag_name=? order by file_name", bagName)
	if err != nil {
		return nil, err
	}
	artifacts := make([]*Artifact, 0)
	for rows.Next() {
		artifact := Artifact{}
		err = rows.Scan(
			&artifact.ID,
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
