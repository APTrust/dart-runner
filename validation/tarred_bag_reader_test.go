package validation_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/util"
	"github.com/APTrust/dart-runner/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTarredBagScanMetadata(t *testing.T) {
	pathToBag := util.PathToUnitTestBag("example.edu.tagsample_good.tar")
	validator, err := validation.NewValidator(pathToBag)
	require.Nil(t, err)
	reader, err := validation.NewTarredBagReader(validator)
	require.Nil(t, err)
	err = reader.ScanMetadata()
	require.Nil(t, err)

	err = reader.ScanPayload()
	require.Nil(t, err)

	data, _ := json.MarshalIndent(validator, "", "  ")
	fmt.Println(string(data))
	assert.True(t, false)
}
