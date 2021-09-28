package util_test

import (
	"os"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringListContains(t *testing.T) {
	list := []string{"apple", "orange", "banana"}
	assert.True(t, util.StringListContains(list, "orange"))
	assert.False(t, util.StringListContains(list, "wedgie"))
	// Don't crash on nil list
	assert.False(t, util.StringListContains(nil, "mars"))
}

func TestStringListContainsAll(t *testing.T) {
	list1 := []string{"apple", "orange", "banana"}
	list2 := []string{"apple", "orange", "banana"}
	list3 := []string{"apple", "orange", "fig"}

	assert.True(t, util.StringListContainsAll(list1, list2))
	assert.False(t, util.StringListContainsAll(list1, list3))
}

func TestIntListContains(t *testing.T) {
	list := []int{3, 5, 7}
	assert.True(t, util.IntListContains(list, 3))
	assert.False(t, util.IntListContains(list, 8))
	// Don't crash on nil list
	assert.False(t, util.IntListContains(nil, 3))
}

func TestAlgorithmFromManifestName(t *testing.T) {
	names := map[string]string{
		"manifest-md5.txt":       "md5",
		"tagmanifest-sha256.txt": "sha256",
		"manifest-sha512.txt":    "sha512",
	}
	for filename, algorithm := range names {
		alg, err := util.AlgorithmFromManifestName(filename)
		assert.Nil(t, err)
		assert.Equal(t, algorithm, alg)
	}
	_, err := util.AlgorithmFromManifestName("bad-file-name.txt")
	assert.NotNil(t, err)
}

func TestLooksLikeURL(t *testing.T) {
	assert.True(t, util.LooksLikeURL("http://s3.amazonaws.com/bucket/key"))
	assert.True(t, util.LooksLikeURL("https://s3.amazonaws.com/bucket/key"))
	assert.False(t, util.LooksLikeURL("tpph\\backslash\\slackbash\\iaintnourl!"))
	assert.False(t, util.LooksLikeURL(""))
}

func TestLooksLikeUUID(t *testing.T) {
	assert.True(t, util.LooksLikeUUID("1552abf5-28f3-46a5-ba63-95302d08e209"))
	assert.True(t, util.LooksLikeUUID("88198c5a-ec91-4ce1-bfcc-0f607ebdcca3"))
	assert.True(t, util.LooksLikeUUID("88198C5A-EC91-4CE1-BFCC-0F607EBDCCA3"))
	assert.False(t, util.LooksLikeUUID("88198c5a-ec91-4ce1-bfcc-0f607ebdccx3"))
	assert.False(t, util.LooksLikeUUID("88198c5a-ec91-4ce1-bfcc-0f6c"))
	assert.False(t, util.LooksLikeUUID(""))
}

func TestLooksLikeManifest(t *testing.T) {
	assert.True(t, util.LooksLikeManifest("manifest-md5.txt"))
	assert.True(t, util.LooksLikeManifest("manifest-sha256.txt"))
	// No: is tag manifest
	assert.False(t, util.LooksLikeManifest("tagmanifest-md5.txt"))
	// No: is tag file
	assert.False(t, util.LooksLikeManifest("bag-info.txt"))
	// No: is payload file
	assert.False(t, util.LooksLikeManifest("data/manifest-sha256.txt"))
}

func TestLooksLikeTagManifest(t *testing.T) {
	assert.True(t, util.LooksLikeTagManifest("tagmanifest-md5.txt"))
	assert.True(t, util.LooksLikeTagManifest("tagmanifest-sha256.txt"))
	// No: is manifest
	assert.False(t, util.LooksLikeTagManifest("manifest-md5.txt"))
	// No: is tag file
	assert.False(t, util.LooksLikeTagManifest("bag-info.txt"))
	// No: is payload file
	assert.False(t, util.LooksLikeTagManifest("data/manifest-sha256.txt"))
}

func TestLooksLikePayloadFile(t *testing.T) {
	assert.True(t, util.LooksLikePayloadFile("data/file.txt"))
	assert.True(t, util.LooksLikePayloadFile("data/nested/file.txt"))
	assert.False(t, util.LooksLikePayloadFile("tagmanifest-sha256.txt"))
	assert.False(t, util.LooksLikePayloadFile("manifest-md5.txt"))
	assert.False(t, util.LooksLikePayloadFile("bag-info.txt"))
	assert.False(t, util.LooksLikePayloadFile("bagit.txt"))
}

func TestBagFileType(t *testing.T) {
	assert.Equal(t, constants.FileTypePayload, util.BagFileType("data/file.pdf"))
	assert.Equal(t, constants.FileTypeManifest, util.BagFileType("manifest-sha256.txt"))
	assert.Equal(t, constants.FileTypeTagManifest, util.BagFileType("tagmanifest-sha256.txt"))
	assert.Equal(t, constants.FileTypeTag, util.BagFileType("ramdom_file.xml"))
}

func TestContainsControlCharacter(t *testing.T) {
	assert.True(t, util.ContainsControlCharacter("\u0000 -- NULL"))
	assert.True(t, util.ContainsControlCharacter("\u0001 -- START OF HEADING"))
	assert.True(t, util.ContainsControlCharacter("\u0002 -- START OF TEXT"))
	assert.True(t, util.ContainsControlCharacter("\u0003 -- END OF TEXT"))
	assert.True(t, util.ContainsControlCharacter("\u0004 -- END OF TRANSMISSION"))
	assert.True(t, util.ContainsControlCharacter("\u0005 -- ENQUIRY"))
	assert.True(t, util.ContainsControlCharacter("\u0006 -- ACKNOWLEDGE"))
	assert.True(t, util.ContainsControlCharacter("\u0007 -- BELL"))
	assert.True(t, util.ContainsControlCharacter("\u0008 -- BACKSPACE"))
	assert.True(t, util.ContainsControlCharacter("\u0009 -- CHARACTER TABULATION"))
	assert.True(t, util.ContainsControlCharacter("\u000A -- LINE FEED (LF)"))
	assert.True(t, util.ContainsControlCharacter("\u000B -- LINE TABULATION"))
	assert.True(t, util.ContainsControlCharacter("\u000C -- FORM FEED (FF)"))
	assert.True(t, util.ContainsControlCharacter("\u000D -- CARRIAGE RETURN (CR)"))
	assert.True(t, util.ContainsControlCharacter("\u000E -- SHIFT OUT"))
	assert.True(t, util.ContainsControlCharacter("\u000F -- SHIFT IN"))
	assert.True(t, util.ContainsControlCharacter("\u0010 -- DATA LINK ESCAPE"))
	assert.True(t, util.ContainsControlCharacter("\u0011 -- DEVICE CONTROL ONE"))
	assert.True(t, util.ContainsControlCharacter("\u0012 -- DEVICE CONTROL TWO"))
	assert.True(t, util.ContainsControlCharacter("\u0013 -- DEVICE CONTROL THREE"))
	assert.True(t, util.ContainsControlCharacter("\u0014 -- DEVICE CONTROL FOUR"))
	assert.True(t, util.ContainsControlCharacter("\u0015 -- NEGATIVE ACKNOWLEDGE"))
	assert.True(t, util.ContainsControlCharacter("\u0016 -- SYNCHRONOUS IDLE"))
	assert.True(t, util.ContainsControlCharacter("\u0017 -- END OF TRANSMISSION BLOCK"))
	assert.True(t, util.ContainsControlCharacter("\u0018 -- CANCEL"))
	assert.True(t, util.ContainsControlCharacter("\u0019 -- END OF MEDIUM"))
	assert.True(t, util.ContainsControlCharacter("\u001A -- SUBSTITUTE"))
	assert.True(t, util.ContainsControlCharacter("\u001B -- ESCAPE"))
	assert.True(t, util.ContainsControlCharacter("\u001C -- INFORMATION SEPARATOR FOUR"))
	assert.True(t, util.ContainsControlCharacter("\u001D -- INFORMATION SEPARATOR THREE"))
	assert.True(t, util.ContainsControlCharacter("\u001E -- INFORMATION SEPARATOR TWO"))
	assert.True(t, util.ContainsControlCharacter("\u001F -- INFORMATION SEPARATOR ONE"))
	assert.True(t, util.ContainsControlCharacter("\u007F -- DELETE"))
	assert.True(t, util.ContainsControlCharacter("\u0080 -- <control>"))
	assert.True(t, util.ContainsControlCharacter("\u0081 -- <control>"))
	assert.True(t, util.ContainsControlCharacter("\u0082 -- BREAK PERMITTED HERE"))
	assert.True(t, util.ContainsControlCharacter("\u0083 -- NO BREAK HERE"))
	assert.True(t, util.ContainsControlCharacter("\u0084 -- <control>"))
	assert.True(t, util.ContainsControlCharacter("\u0085 -- NEXT LINE (NEL)"))
	assert.True(t, util.ContainsControlCharacter("\u0086 -- START OF SELECTED AREA"))
	assert.True(t, util.ContainsControlCharacter("\u0087 -- END OF SELECTED AREA"))
	assert.True(t, util.ContainsControlCharacter("\u0088 -- CHARACTER TABULATION SET"))
	assert.True(t, util.ContainsControlCharacter("\u0089 -- CHARACTER TABULATION WITH JUSTIFICATION"))
	assert.True(t, util.ContainsControlCharacter("\u008A -- LINE TABULATION SET"))
	assert.True(t, util.ContainsControlCharacter("\u008B -- PARTIAL LINE FORWARD"))
	assert.True(t, util.ContainsControlCharacter("\u008C -- PARTIAL LINE BACKWARD"))
	assert.True(t, util.ContainsControlCharacter("\u008D -- REVERSE LINE FEED"))
	assert.True(t, util.ContainsControlCharacter("\u008E -- SINGLE SHIFT TWO"))
	assert.True(t, util.ContainsControlCharacter("\u008F -- SINGLE SHIFT THREE"))
	assert.True(t, util.ContainsControlCharacter("\u0090 -- DEVICE CONTROL STRING"))
	assert.True(t, util.ContainsControlCharacter("\u0091 -- PRIVATE USE ONE"))
	assert.True(t, util.ContainsControlCharacter("\u0092 -- PRIVATE USE TWO"))
	assert.True(t, util.ContainsControlCharacter("\u0093 -- SET TRANSMIT STATE"))
	assert.True(t, util.ContainsControlCharacter("\u0094 -- CANCEL CHARACTER"))
	assert.True(t, util.ContainsControlCharacter("\u0095 -- MESSAGE WAITING"))
	assert.True(t, util.ContainsControlCharacter("\u0096 -- START OF GUARDED AREA"))
	assert.True(t, util.ContainsControlCharacter("\u0097 -- END OF GUARDED AREA"))
	assert.True(t, util.ContainsControlCharacter("\u0098 -- START OF STRING"))
	assert.True(t, util.ContainsControlCharacter("\u0099 -- <control>"))
	assert.True(t, util.ContainsControlCharacter("\u009A -- SINGLE CHARACTER INTRODUCER"))
	assert.True(t, util.ContainsControlCharacter("\u009B -- CONTROL SEQUENCE INTRODUCER"))
	assert.True(t, util.ContainsControlCharacter("\u009C -- STRING TERMINATOR"))
	assert.True(t, util.ContainsControlCharacter("\u009D -- OPERATING SYSTEM COMMAND"))
	assert.True(t, util.ContainsControlCharacter("\u009E -- PRIVACY MESSAGE"))
	assert.True(t, util.ContainsControlCharacter("\u009F -- APPLICATION PROGRAM COMMAND"))
	assert.True(t, util.ContainsControlCharacter("data/datastream\u007f.txt"))

	assert.False(t, util.ContainsControlCharacter("./this/is/a/valid/file/name.txt"))
}

func TestContainsEscapedControl(t *testing.T) {
	assert.True(t, util.ContainsEscapedControl("\\u0000 -- NULL"))
	assert.True(t, util.ContainsEscapedControl("\\u0001 -- START OF HEADING"))
	assert.True(t, util.ContainsEscapedControl("\\u0002 -- START OF TEXT"))
	assert.True(t, util.ContainsEscapedControl("\\u0003 -- END OF TEXT"))
	assert.True(t, util.ContainsEscapedControl("\\u0004 -- END OF TRANSMISSION"))
	assert.True(t, util.ContainsEscapedControl("\\u0005 -- ENQUIRY"))
	assert.True(t, util.ContainsEscapedControl("\\u0006 -- ACKNOWLEDGE"))
	assert.True(t, util.ContainsEscapedControl("\\u0007 -- BELL"))
	assert.True(t, util.ContainsEscapedControl("\\u0008 -- BACKSPACE"))
	assert.True(t, util.ContainsEscapedControl("\\u0009 -- CHARACTER TABULATION"))
	assert.True(t, util.ContainsEscapedControl("\\u000A -- LINE FEED (LF)"))
	assert.True(t, util.ContainsEscapedControl("\\u000B -- LINE TABULATION"))
	assert.True(t, util.ContainsEscapedControl("\\u000C -- FORM FEED (FF)"))
	assert.True(t, util.ContainsEscapedControl("\\u000D -- CARRIAGE RETURN (CR)"))
	assert.True(t, util.ContainsEscapedControl("\\u000E -- SHIFT OUT"))
	assert.True(t, util.ContainsEscapedControl("\\u000F -- SHIFT IN"))
	assert.True(t, util.ContainsEscapedControl("\\u0010 -- DATA LINK ESCAPE"))
	assert.True(t, util.ContainsEscapedControl("\\u0011 -- DEVICE CONTROL ONE"))
	assert.True(t, util.ContainsEscapedControl("\\u0012 -- DEVICE CONTROL TWO"))
	assert.True(t, util.ContainsEscapedControl("\\u0013 -- DEVICE CONTROL THREE"))
	assert.True(t, util.ContainsEscapedControl("\\u0014 -- DEVICE CONTROL FOUR"))
	assert.True(t, util.ContainsEscapedControl("\\u0015 -- NEGATIVE ACKNOWLEDGE"))
	assert.True(t, util.ContainsEscapedControl("\\u0016 -- SYNCHRONOUS IDLE"))
	assert.True(t, util.ContainsEscapedControl("\\u0017 -- END OF TRANSMISSION BLOCK"))
	assert.True(t, util.ContainsEscapedControl("\\u0018 -- CANCEL"))
	assert.True(t, util.ContainsEscapedControl("\\u0019 -- END OF MEDIUM"))
	assert.True(t, util.ContainsEscapedControl("\\u001A -- SUBSTITUTE"))
	assert.True(t, util.ContainsEscapedControl("\\u001B -- ESCAPE"))
	assert.True(t, util.ContainsEscapedControl("\\u001C -- INFORMATION SEPARATOR FOUR"))
	assert.True(t, util.ContainsEscapedControl("\\u001D -- INFORMATION SEPARATOR THREE"))
	assert.True(t, util.ContainsEscapedControl("\\u001E -- INFORMATION SEPARATOR TWO"))
	assert.True(t, util.ContainsEscapedControl("\\u001F -- INFORMATION SEPARATOR ONE"))
	assert.True(t, util.ContainsEscapedControl("\\u007F -- DELETE"))
	assert.True(t, util.ContainsEscapedControl("\\u0080 -- <control>"))
	assert.True(t, util.ContainsEscapedControl("\\u0081 -- <control>"))
	assert.True(t, util.ContainsEscapedControl("\\u0082 -- BREAK PERMITTED HERE"))
	assert.True(t, util.ContainsEscapedControl("\\u0083 -- NO BREAK HERE"))
	assert.True(t, util.ContainsEscapedControl("\\u0084 -- <control>"))
	assert.True(t, util.ContainsEscapedControl("\\u0085 -- NEXT LINE (NEL)"))
	assert.True(t, util.ContainsEscapedControl("\\u0086 -- START OF SELECTED AREA"))
	assert.True(t, util.ContainsEscapedControl("\\u0087 -- END OF SELECTED AREA"))
	assert.True(t, util.ContainsEscapedControl("\\u0088 -- CHARACTER TABULATION SET"))
	assert.True(t, util.ContainsEscapedControl("\\u0089 -- CHARACTER TABULATION WITH JUSTIFICATION"))
	assert.True(t, util.ContainsEscapedControl("\\u008A -- LINE TABULATION SET"))
	assert.True(t, util.ContainsEscapedControl("\\u008B -- PARTIAL LINE FORWARD"))
	assert.True(t, util.ContainsEscapedControl("\\u008C -- PARTIAL LINE BACKWARD"))
	assert.True(t, util.ContainsEscapedControl("\\u008D -- REVERSE LINE FEED"))
	assert.True(t, util.ContainsEscapedControl("\\u008E -- SINGLE SHIFT TWO"))
	assert.True(t, util.ContainsEscapedControl("\\u008F -- SINGLE SHIFT THREE"))
	assert.True(t, util.ContainsEscapedControl("\\u0090 -- DEVICE CONTROL STRING"))
	assert.True(t, util.ContainsEscapedControl("\\u0091 -- PRIVATE USE ONE"))
	assert.True(t, util.ContainsEscapedControl("\\u0092 -- PRIVATE USE TWO"))
	assert.True(t, util.ContainsEscapedControl("\\u0093 -- SET TRANSMIT STATE"))
	assert.True(t, util.ContainsEscapedControl("\\u0094 -- CANCEL CHARACTER"))
	assert.True(t, util.ContainsEscapedControl("\\u0095 -- MESSAGE WAITING"))
	assert.True(t, util.ContainsEscapedControl("\\u0096 -- START OF GUARDED AREA"))
	assert.True(t, util.ContainsEscapedControl("\\u0097 -- END OF GUARDED AREA"))
	assert.True(t, util.ContainsEscapedControl("\\u0098 -- START OF STRING"))
	assert.True(t, util.ContainsEscapedControl("\\u0099 -- <control>"))
	assert.True(t, util.ContainsEscapedControl("\\u009A -- SINGLE CHARACTER INTRODUCER"))
	assert.True(t, util.ContainsEscapedControl("\\u009B -- CONTROL SEQUENCE INTRODUCER"))
	assert.True(t, util.ContainsEscapedControl("\\u009C -- STRING TERMINATOR"))
	assert.True(t, util.ContainsEscapedControl("\\u009D -- OPERATING SYSTEM COMMAND"))
	assert.True(t, util.ContainsEscapedControl("\\u009E -- PRIVACY MESSAGE"))
	assert.True(t, util.ContainsEscapedControl("\\u009F -- APPLICATION PROGRAM COMMAND"))
	assert.True(t, util.ContainsEscapedControl("data/datastream\\u007f.txt"))

	assert.False(t, util.ContainsEscapedControl("./this/is/a/valid/file/name.txt"))
}

func TestUCFirst(t *testing.T) {
	assert.Equal(t, "Institution", util.UCFirst("institution"))
	assert.Equal(t, "Institution", util.UCFirst("INSTITUTION"))
	assert.Equal(t, "Institution", util.UCFirst("inStiTuTioN"))
}

func TestTarPathToBagPath(t *testing.T) {
	pathInBag, err := util.TarPathToBagPath("my_bag/bagit.txt")
	require.Nil(t, err)
	assert.Equal(t, "bagit.txt", pathInBag)

	pathInBag, err = util.TarPathToBagPath("my_bag/data/file.docx")
	require.Nil(t, err)
	assert.Equal(t, "data/file.docx", pathInBag)

	pathInBag, err = util.TarPathToBagPath("my_bag/data/img/photo.jpg")
	require.Nil(t, err)
	assert.Equal(t, "data/img/photo.jpg", pathInBag)

	// Should be an error. We're expecting a top-level directory.
	// bagit.txt and the data dir should be inside of that.
	pathInBag, err = util.TarPathToBagPath("bagit.txt")
	assert.NotNil(t, err)
}

func TestPathTo(t *testing.T) {
	programs := []string{
		"go",
		"ls",
		"which",
	}
	pathSep := string(os.PathSeparator)
	for _, program := range programs {
		pathToProgram, err := util.PathTo(program)
		require.Nil(t, err)
		assert.True(t, strings.HasSuffix(pathToProgram, pathSep+program))
	}
}

func TestStringIsShellSafe(t *testing.T) {
	assert.True(t, util.StringIsShellSafe("https://example.com?a=b"))
	assert.False(t, util.StringIsShellSafe("No\""))
	assert.False(t, util.StringIsShellSafe("No'"))
	assert.False(t, util.StringIsShellSafe("No;"))
	assert.False(t, util.StringIsShellSafe("No{"))
	assert.False(t, util.StringIsShellSafe("No}"))
	assert.False(t, util.StringIsShellSafe("No|"))
	assert.False(t, util.StringIsShellSafe("No$"))
	assert.False(t, util.StringIsShellSafe("No\t"))
	assert.False(t, util.StringIsShellSafe("No\r"))
	assert.False(t, util.StringIsShellSafe("No\n"))
	assert.False(t, util.StringIsShellSafe("No<"))
	assert.False(t, util.StringIsShellSafe("No>"))
}

func TestStripFileExtension(t *testing.T) {
	assert.Equal(t, "somebag", util.StripFileExtension("somebag.tar"))
	assert.Equal(t, "some_file", util.StripFileExtension("some_file.txt"))
}

func TestProjectRoot(t *testing.T) {
	projectRoot := util.ProjectRoot()
	assert.True(t, len(projectRoot) > 2)
	assert.True(t, strings.Contains(projectRoot, string(os.PathSeparator)))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 2, util.Min(2, 10))
	assert.Equal(t, 2, util.Min(10, 2))
}

func TestEstimatedChunkSize(t *testing.T) {
	mb := float64(1024 * 1024)
	gb := float64(mb * 1024)

	// 5 MB is min. Should not go lower than that.
	assert.Equal(t, uint64(5*mb), util.EstimatedChunkSize(float64(3000)))

	// 500 MB is our max. Make sure we get that and not 800 MB here.
	assert.Equal(t, uint64(500*mb), util.EstimatedChunkSize(float64(8*1024*gb)))

	// 500 MB chunks for 2 TB upload
	assert.Equal(t, uint64(500*mb), util.EstimatedChunkSize(float64(8*1024*gb)))

	// 600 GB total size -> ~60 MB chunks
	assert.Equal(t, uint64(64424510), util.EstimatedChunkSize(float64(600*gb)))

	// 200 GB total size -> ~40 MB chunks
	assert.Equal(t, uint64(42949673), util.EstimatedChunkSize(float64(200*gb)))

	// 80 GB total size -> ~32 MB chunks
	assert.Equal(t, uint64(34359739), util.EstimatedChunkSize(float64(80*gb)))

	// 8 GB total size -> ~16 MB chunks
	assert.Equal(t, uint64(17179870), util.EstimatedChunkSize(float64(8*gb)))

	// 3 GB total size -> ~6 MB chunks
	assert.Equal(t, uint64(6442451), util.EstimatedChunkSize(float64(3*gb)))

	// 100 MB total size -> ~5 MB chunks
	assert.Equal(t, uint64(5242880), util.EstimatedChunkSize(float64(100*mb)))

}

func TestCleanBagName(t *testing.T) {
	expected := "some.file"
	assert.Equal(t, expected, util.CleanBagName("some.file.b001.of200.tar"))
	assert.Equal(t, expected, util.CleanBagName("some.file.b1.of2.tar"))
	assert.Equal(t, expected, util.CleanBagName("some.file.tar"))
	assert.Equal(t, expected, util.CleanBagName("some.file"))
}

func TestIsEmptyStringList(t *testing.T) {
	assert.True(t, util.IsEmptyStringList(nil))

	list1 := make([]string, 0)
	assert.True(t, util.IsEmptyStringList(list1))

	list2 := make([]string, 5)
	assert.True(t, util.IsEmptyStringList(list2))

	list3 := []string{
		"",
		"not empty",
		"",
	}
	assert.False(t, util.IsEmptyStringList(list3))
}

func TestFileCommonPrefix(t *testing.T) {
	list1 := []string{
		"/user/joe/photo",
		"/user/joe/docs/resume",
		"/user/joe/docs/letter",
		"/user/joe/photos/car",
	}
	list2 := []string{
		"/user/joe/photos/dog",
		"/user/joe/photos/car",
		"/user/joe/photos/house",
	}
	list3 := []string{
		"/home/linus/torvalds",
		"/user/joe/photos/car",
		"/etc/apache2/conf",
	}
	list4 := []string{
		"/home/linus/torvalds",
		"my_photos",
	}
	assert.Equal(t, "/user/joe/", util.FindCommonPrefix(list1))
	assert.Equal(t, "/user/joe/photos/", util.FindCommonPrefix(list2))
	assert.Equal(t, "/", util.FindCommonPrefix(list3))
	assert.Equal(t, "", util.FindCommonPrefix(list4))
}
