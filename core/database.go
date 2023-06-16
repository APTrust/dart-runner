package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type DBObject struct {
	ID        string
	Type      string
	Name      string
	Json      string
	UpdatedAt time.Time
}

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
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	stmt := `insert into dart (uuid, obj_type, obj_name, obj_json, updated_at) values (?,?,?,?,?)
	on conflict do update set obj_name=excluded.obj_name, obj_json=excluded.obj_json, updated_at=excluded.updated_at where uuid=excluded.uuid`
	_, err = Dart.DB.Exec(stmt, obj.ObjID(), obj.ObjType(), obj.ObjName(), string(jsonBytes), time.Now().UTC())
	return err
}

func ObjFind(uuid string) (*QueryResult, error) {
	var objType string
	var objJson string
	row := Dart.DB.QueryRow("select obj_type, obj_json from dart where uuid=?", uuid)
	err := row.Scan(&objType, &objJson)
	if err != nil {
		return nil, err
	}
	qr := NewQueryResult(objType)
	switch objType {
	case constants.TypeAppSetting:
		a := &AppSetting{}
		err = json.Unmarshal([]byte(objJson), a)
		qr.AppSetting = a
	case constants.TypeInternalSetting:
		i := &InternalSetting{}
		err = json.Unmarshal([]byte(objJson), i)
		qr.InternalSetting = i
	case constants.TypeStorageService:
		s := &StorageService{}
		err = json.Unmarshal([]byte(objJson), s)
		qr.StorageService = s
	case constants.TypeRemoteRepository:
		r := &RemoteRepository{}
		err = json.Unmarshal([]byte(objJson), r)
		qr.RemoteRepository = r
	default:
		return nil, fmt.Errorf("cannot convert unknown type %s to query result", objType)
	}

	return qr, err
}

func ObjList(objType, orderBy string, limit, offset int) (*QueryResult, error) {
	rows, err := Dart.DB.Query("select obj_json from dart where obj_type = ? order by ? limit ? offset ?", objType, orderBy, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	qr := NewQueryResult(objType)
	switch objType {
	case constants.TypeAppSetting:
		qr.AppSettings, err = appSettingsList(rows)
	case constants.TypeInternalSetting:
		qr.InternalSettings, err = internalSettingList(rows)
	case constants.TypeRemoteRepository:
		qr.RemoteRepositories, err = remoteRepositoryList(rows)
	case constants.TypeStorageService:
		qr.StorageServices, err = storageServiceList(rows)
	default:
		return nil, fmt.Errorf("cannot convert unknown type %s to query result", objType)
	}
	return qr, err
}

func appSettingsList(rows *sql.Rows) ([]*AppSetting, error) {
	list := make([]*AppSetting, 0)
	for rows.Next() {
		var jsonStr string
		err := rows.Scan(&jsonStr)
		if err != nil {
			return nil, err
		}
		item := &AppSetting{}
		err = json.Unmarshal([]byte(jsonStr), item)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func internalSettingList(rows *sql.Rows) ([]*InternalSetting, error) {
	list := make([]*InternalSetting, 0)
	for rows.Next() {
		var jsonStr string
		err := rows.Scan(&jsonStr)
		if err != nil {
			return nil, err
		}
		item := &InternalSetting{}
		err = json.Unmarshal([]byte(jsonStr), item)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func remoteRepositoryList(rows *sql.Rows) ([]*RemoteRepository, error) {
	list := make([]*RemoteRepository, 0)
	for rows.Next() {
		var jsonStr string
		err := rows.Scan(&jsonStr)
		if err != nil {
			return nil, err
		}
		item := &RemoteRepository{}
		err = json.Unmarshal([]byte(jsonStr), item)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func storageServiceList(rows *sql.Rows) ([]*StorageService, error) {
	list := make([]*StorageService, 0)
	for rows.Next() {
		var jsonStr string
		err := rows.Scan(&jsonStr)
		if err != nil {
			return nil, err
		}
		item := &StorageService{}
		err = json.Unmarshal([]byte(jsonStr), item)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func ObjDelete(uuid string) error {
	_, err := Dart.DB.Exec("delete from dart where uuid=?", uuid)
	return err
}

func ArtifactSave(a *Artifact) error {
	stmt := `insert into artifacts (uuid, bag_name, item_type, file_name, file_type, raw_data, updated_at) values (?,?,?,?,?,?,?);`
	_, err := Dart.DB.Exec(stmt, a.ID, a.BagName, a.ItemType, a.FileName, a.FileType, a.RawData, time.Now().UTC())
	return err
}

func ArtifactGet(uuid string) (*Artifact, error) {
	row := Dart.DB.QueryRow("select uuid, bag_name, item_type, file_name, file_type, raw_data, updated_at from artifacts where uuid=?", uuid)
	artifact := &Artifact{}
	err := row.Scan(
		artifact.ID,
		artifact.BagName,
		artifact.ItemType,
		artifact.FileName,
		artifact.FileType,
		artifact.RawData,
		artifact.UpdatedAt,
	)
	return artifact, err
}

func ArtifactList(bagName string) ([]*Artifact, error) {
	rows, err := Dart.DB.Query("select uuid, bag_name, item_type, file_name, file_type, raw_data, updated_at from artifacts where bag_name=? order by updated_at desc, item_type, file_name", bagName)
	if err != nil {
		return nil, err
	}
	artifacts := make([]*Artifact, 0)
	for rows.Next() {
		artifact := &Artifact{}
		err = rows.Scan(
			artifact.ID,
			artifact.BagName,
			artifact.ItemType,
			artifact.FileName,
			artifact.FileType,
			artifact.RawData,
			artifact.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
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
		return ErrInvalidOperation
	}
	_, err := Dart.DB.Exec("delete from dart")
	return err
}

// ClearArtifactsTable is for testing use only
func ClearArtifactsTable() error {
	if !util.TestsAreRunning() {
		return ErrInvalidOperation
	}
	_, err := Dart.DB.Exec("delete from artifacts")
	return err
}
