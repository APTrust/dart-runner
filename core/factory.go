package core

import (
	"fmt"
	"path"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

// This file contains factory methods to be used in
// unit and/or integration tests.

func CreateAppSettings(howMany int) ([]*AppSetting, error) {
	settings := make([]*AppSetting, howMany)
	for i := 0; i < howMany; i++ {
		name := fmt.Sprintf("%s (%d)", gofakeit.Word(), i)
		setting := NewAppSetting(name, gofakeit.Phrase())
		err := ObjSave(setting)
		if err != nil {
			return settings, err
		}
		settings[i] = setting
	}
	return settings, nil
}

func CreateBagItProfiles(howMany int) ([]*BagItProfile, error) {
	// Empty profile satisfies the bare requirements of the BagIt spec
	// without any customizations.
	profiles := make([]*BagItProfile, howMany)
	filename := path.Join(util.ProjectRoot(), "profiles", "empty_profile.json")
	emptyProfile, err := BagItProfileLoad(filename)
	if err != nil {
		return profiles, err
	}
	for i := 0; i < howMany; i++ {
		profile := BagItProfileClone(emptyProfile)
		profile.Name = fmt.Sprintf("%s %s (%d)", gofakeit.FarmAnimal(), gofakeit.BeerName(), i)
		err := ObjSave(profile)
		if err != nil {
			return profiles, err
		}
		profiles[i] = profile
	}
	return profiles, nil
}

func CreateRemoteRepos(howMany int) ([]*RemoteRepository, error) {
	repos := make([]*RemoteRepository, howMany)
	for i := 0; i < howMany; i++ {
		repo := NewRemoteRepository()
		repo.Name = fmt.Sprintf("%s (%d)", gofakeit.HackerPhrase(), i)
		repo.APIToken = uuid.NewString()
		repo.LoginExtra = gofakeit.Slogan()
		repo.PluginID = uuid.NewString()
		repo.Url = gofakeit.URL()
		repo.UserID = gofakeit.Email()
		err := ObjSave(repo)
		if err != nil {
			return repos, err
		}
		repos[i] = repo
	}
	return repos, nil

}

func CreateStorageServices(howMany int) ([]*StorageService, error) {
	services := make([]*StorageService, howMany)
	for i := 0; i < howMany; i++ {
		allowsUpload := true
		allowsDownload := false
		protocol := constants.ProtocolS3
		if i%2 == 0 {
			allowsDownload = true
			protocol = constants.ProtocolSFTP
		}

		ss := NewStorageService()
		ss.AllowsDownload = allowsDownload
		ss.AllowsUpload = allowsUpload
		ss.Bucket = gofakeit.BuzzWord()
		ss.Description = gofakeit.Phrase()
		ss.Host = gofakeit.DomainName()
		ss.Login = gofakeit.Email()
		ss.LoginExtra = gofakeit.Word()
		ss.Name = gofakeit.Phrase()
		ss.Password = gofakeit.BuzzWord()
		ss.Port = 1001 * i
		ss.Protocol = protocol
		err := ObjSave(ss)
		if err != nil {
			return services, err
		}
		services[i] = ss
	}
	return services, nil
}
