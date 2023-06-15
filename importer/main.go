package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/APTrust/dart-runner/core"
)

func main() {
	dataDir := core.Dart.Paths.DataDir
	appSettingsFile := path.Join(dataDir, "AppSetting.json")
	appSettingsJson, err := os.ReadFile(appSettingsFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	appSettings := make(map[string]*core.AppSetting)
	err = json.Unmarshal(appSettingsJson, &appSettings)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, setting := range appSettings {
		saveErr := setting.Save()
		if saveErr != nil {
			fmt.Printf("Error saving setting %s: %v\n", setting.Name, saveErr)
		} else {
			fmt.Printf("Saved setting %s\n", setting.Name)
		}
	}
}
