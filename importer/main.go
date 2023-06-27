package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/APTrust/dart-runner/core"
)

func main() {
	ImportAppSettings()
	ImportBagItProfiles()
	ImportRemoteRepositories()
	ImportStorageServices()
}

func ImportAppSettings() {
	jsonBytes := GetJson("AppSetting.json")
	appSettings := make(map[string]*core.AppSetting)
	ParseJson(jsonBytes, &appSettings)
	for _, setting := range appSettings {
		SaveObject(setting)
	}
}

func ImportBagItProfiles() {
	jsonBytes := GetJson("BagItProfile.json")
	profiles := make(map[string]*core.BagItProfile)
	ParseJson(jsonBytes, &profiles)
	for _, setting := range profiles {
		SaveObject(setting)
	}
}

func ImportRemoteRepositories() {
	jsonBytes := GetJson("RemoteRepository.json")
	repos := make(map[string]*core.RemoteRepository)
	ParseJson(jsonBytes, &repos)
	for _, repo := range repos {
		SaveObject(repo)
	}
}

func ImportStorageServices() {
	jsonBytes := GetJson("StorageService.json")
	services := make(map[string]*core.StorageService)
	ParseJson(jsonBytes, &services)
	for _, ss := range services {
		SaveObject(ss)
	}
}

func GetJson(filename string) []byte {
	dataDir := core.Dart.Paths.DataDir
	jsonFile := path.Join(dataDir, filename)
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return jsonData
}

func ParseJson(jsonBytes []byte, objMap interface{}) {
	err := json.Unmarshal(jsonBytes, objMap)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func SaveObject(obj core.PersistentObject) {
	saveErr := core.ObjSave(obj)
	if saveErr != nil {
		fmt.Printf("Error saving setting %s: %v\n", obj.ObjName(), saveErr)
		for key, value := range obj.GetErrors() {
			fmt.Println("  ", key, "->", value)
		}
	} else {
		fmt.Printf("Saved %s %s\n", obj.ObjType(), obj.ObjName())
	}
}
