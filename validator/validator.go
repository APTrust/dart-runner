package validator

import (
	"os"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/APTrust/dart-runner/util"
)

type Validator struct {
	PathToBag        string
	Profile          *bagit.Profile
	PayloadFiles     *FileMap
	PayloadManifests *FileMap
	TagFiles         *FileMap
	TagManifests     *FileMap
	Errors           []error
}

func NewValidator(pathToBag string) (*Validator, error) {
	if !util.FileExists(pathToBag) {
		return nil, os.ErrNotExist
	}
	return &Validator{
		PathToBag:        pathToBag,
		PayloadFiles:     NewFileMap(FileTypePayload),
		PayloadManifests: NewFileMap(FileTypeManifest),
		TagFiles:         NewFileMap(FileTypeTagFile),
		TagManifests:     NewFileMap(FileTypeTagManifest),
		Errors:           make([]error, 0),
	}, nil
}

func (v *Validator) ManifestAlgs() ([]DigestAlgorithm, error) {
	return algsFromFileMap(v.PayloadManifests)
}

func (v *Validator) TagManifestAlgs() ([]DigestAlgorithm, error) {
	return algsFromFileMap(v.TagManifests)
}

func algsFromFileMap(fileMap *FileMap) ([]DigestAlgorithm, error) {
	algs := make([]DigestAlgorithm, 0)
	for filename, _ := range fileMap.Files {
		algName, err := util.AlgorithmFromManifestName(filename)
		if err != nil {
			return nil, err
		}
		alg, err := AlgToEnum(algName)
		if err != nil {
			return nil, err
		}
		algs = append(algs, alg)
	}
	return algs, nil
}
