package util

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/APTrust/dart-runner/constants"
)

// StringListContains returns true if the list of strings contains item.
func StringListContains(list []string, item string) bool {
	if list != nil {
		for i := range list {
			if list[i] == item {
				return true
			}
		}
	}
	return false
}

// StringListContainsAll returns true if all items in listToCheck are also
// in masterList. Be sure you pass the params in the right order. Note
// that this can get expensive if the lists are long.
func StringListContainsAll(masterList []string, listToCheck []string) bool {
	for _, item := range listToCheck {
		if !StringListContains(masterList, item) {
			return false
		}
	}
	return true
}

// IsEmptyStringList returns true if list contains no items or if all
// items in list are empty strings.
func IsEmptyStringList(list []string) bool {
	if list == nil || len(list) == 0 {
		return true
	}
	for _, s := range list {
		if s != "" {
			return false
		}
	}
	return true
}

// CopyMap returns a copy of map orig
func CopyMap[K, V comparable](orig map[K]V) map[K]V {
	newCopy := make(map[K]V)
	for k, v := range orig {
		newCopy[k] = v
	}
	return newCopy
}

// RemoveFromSlice removes the item at index index from slice,
// returning a new slice.
func RemoveFromSlice[T any](list []T, index int) []T {
	return append(list[:index], list[index+1:]...)
}

// IsEmpty returns true if string s is empty. Strings that are all
// whitespace are considered empty.
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// SlitAndTrim splits string s on the specified separator, then
// trims leading and trailing whitespace from each item in the
// resulting slice. Returns a slice of trimmed strings. Empty
// strings will be omitted from the returned list.
func SplitAndTrim(s, sep string) []string {
	values := strings.Split(s, sep)
	trimmedValues := make([]string, 0)
	for _, val := range values {
		str := strings.TrimSpace(val)
		if str != "" {
			trimmedValues = append(trimmedValues, str)
		}
	}
	return trimmedValues
}

// IsListType returns true if obj is a slice or array.
func IsListType(obj interface{}) bool {
	if obj == nil {
		return false
	}
	rt := reflect.TypeOf(obj)
	return rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array
}

// IsMapType returns true if obj is a map.
func IsMapType(obj interface{}) bool {
	if obj == nil {
		return false
	}
	rt := reflect.TypeOf(obj)
	return rt.Kind() == reflect.Map
}

// IntListContains returns true if list contains item.
func IntListContains(list []int, item int) bool {
	for i := range list {
		if list[i] == item {
			return true
		}
	}
	return false
}

func LooksLikeManifest(name string) bool {
	return strings.HasPrefix(name, "manifest-") && strings.HasSuffix(name, ".txt")
}

func LooksLikeTagManifest(name string) bool {
	return strings.HasPrefix(name, "tagmanifest-") && strings.HasSuffix(name, ".txt")
}

func LooksLikePayloadFile(name string) bool {
	return strings.HasPrefix(name, "data/")
}

func BagFileType(name string) string {
	if LooksLikePayloadFile(name) {
		return constants.FileTypePayload
	} else if LooksLikeManifest(name) {
		return constants.FileTypeManifest
	} else if LooksLikeTagManifest(name) {
		return constants.FileTypeTagManifest
	}
	return constants.FileTypeTag
}

// LooksLikeURL returns true if url looks like a URL.
func LooksLikeURL(url string) bool {
	reURL := regexp.MustCompile(`^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`)
	return reURL.Match([]byte(url))
}

// LooksLikeUUID returns true if uuid looks like a valid UUID.
func LooksLikeUUID(uuid string) bool {
	reUUID := regexp.MustCompile(`(?i)^([a-f\d]{8}(-[a-f\d]{4}){3}-[a-f\d]{12}?)$`)
	return reUUID.Match([]byte(uuid))
}

func AlgorithmFromManifestName(filename string) (string, error) {
	re := regexp.MustCompile(`manifest-(?P<Alg>[^\.]+).txt$`)
	match := re.FindStringSubmatch(filename)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", fmt.Errorf("Cannot get algorithm from filename %s", filename)
}

// ContainsControlCharacter returns true if string str contains a
// Unicode control character. We use this to test file names, which
// should not contain control characters.
func ContainsControlCharacter(str string) bool {
	runes := []rune(str)
	for _, _rune := range runes {
		if unicode.IsControl(_rune) {
			return true
		}
	}
	return false
}

// ContainsEscapedControl returns true if string str contains
// something that looks like an escaped UTF-8 control character.
// The Mac OS file system seems to silently escape UTF-8 control
// characters. That causes problems when we try to copy a file
// over to another file system that won't accept the control
// character in a file name. The bag validator looks for file names
// matching these patterns and rejects them.
func ContainsEscapedControl(str string) bool {
	reControl := regexp.MustCompile("\\\\[Uu]00[0189][0-9A-Fa-f]|\\\\[Uu]007[Ff]")
	return reControl.MatchString(str)
}

// StripNonPrintable returns a copy of str with non-printable
// characters removed.
func StripNonPrintable(str string) string {
	str = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, str)
	return str
}

// UCFirst returns string str with the first letter capitalized
// and all others lower case.
func UCFirst(str string) string {
	return strings.Title(strings.ToLower(str))
}

// TarPathToBagPath, given the path of a file inside a tarball, returns
// the path of the file in a bag. The name param generally comes from
// the Name property of a tar file header. For example, in a tar file
// called my_bag.tar the paths would translate as follows:
//
// Input                      ->  Output
// my_bag/bagit.txt           ->  bagit.txt
// my_bag/data/file.docx      ->  data/file.docx
// my_bag/data/img/photo.jpg  ->  data/img/photo.jpg
//
// This function assumes (perhaps dangerously) that tarred bags follow
// the recommdation of pre-1.0 versions of the BagIt spec that say
// a tarred bag should deserialize to a single top-level directory.
// This function does not assume that the directory will match the
// bag name.
func TarPathToBagPath(name string) (string, error) {
	prefix := strings.Split(name, "/")[0] + "/"
	pathInBag := strings.Replace(name, prefix, "", 1)
	if pathInBag == name {
		return "", fmt.Errorf("Illegal path, '%s'. Should start with '%s'.", name, prefix)
	}
	return pathInBag, nil
}

// PathTo returns the path to the specified program.
// Note that this can be spotty on Windows. We'll need
// to watch this, log "not found" errors and work from
// there.
func PathTo(program string) (string, error) {
	cmd := "which"
	args := []string{program}
	if runtime.GOOS == "windows" {
		cmd = "where.exe"
	}
	stdout, stderr, exitcode := ExecCommand(cmd, args, os.Environ(), nil)
	if exitcode != 0 {
		return "", errors.New(string(stderr))
	}
	// Windows where.exe can return multiple lines if the
	// system has multiple copies of an executable in the path.
	// We'll go with the first line.
	lines := strings.Split(string(stdout), "\n")
	return strings.TrimSpace(lines[0]), nil
}

// StringIsShellSafe returns true if string looks safe to pass
// to shell.
func StringIsShellSafe(s string) bool {
	unsafeChars := "\"';{}|$` \t\r\n<>"
	return !strings.ContainsAny(s, unsafeChars)
}

// StripFileExtension returns filename, minus the extension.
// For example, "my_bag.tar" returns "my_bag".
func StripFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return filename[0 : len(filename)-len(ext)]
}

// PrintAndExit prints a message to STDERR and exits
func PrintAndExit(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

// ProjectRoot returns the project root.
func ProjectRoot() string {
	_, thisFile, _, _ := runtime.Caller(0)
	absPath, _ := filepath.Abs(filepath.Join(thisFile, "..", ".."))
	return absPath
}

// PathToTestData returns the path to the directory containing test data.
// This is used only in testing.
func PathToTestData() string {
	return filepath.Join(ProjectRoot(), "testdata")
}

func PathToUnitTestBag(bagName string) string {
	return filepath.Join(ProjectRoot(), "testdata", "bags", bagName)
}

// TestsAreRunning returns true when code is running under "go test"
func TestsAreRunning() bool {
	if strings.HasSuffix(os.Args[0], ".test") || os.Getenv("DART_ENV") == "test" {
		return true
	}
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

// RunningInCI returns true when code is running in the Travis CI
// environment.
func RunningInCI() bool {
	return os.Getenv("TRAVIS_BUILD_DIR") != ""
}

// Min returns the minimum of x or y without all the casting required
// by the math package.
func Min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

// EstimatedChunkSize returns the size we should use for each chunk
// in a multipart S3 upload. If we don't tell Minio what chunk size
// to use, and it doesn't know the size of the total upload, it
// tries to allocate 5 GB of RAM. This causes some restorations to
// fail with an out of memory error. https://trello.com/c/1hkP28x1
//
// Param totalSize is the total size of the object to upload.
// When restoring entire objects, we only know the approximate size,
// which will be IntellectualObject.FileSize plus one or more payload
// manifests and tag manifests of unknown size that we'll have to
// generate on the fly. In practice, we can guesstimate that the
// total size of a restored object will be about 1.01 - 1.1 times
// IntellectualObject.FileSize.
//
// Since S3 max upload size is 5 TB with 10k parts, the max this
// will return is 500MB for part size. Although we could return 5 GB,
// we don't want to because we can't allocate that much memory inside
// of memory-limited docker instances.
func EstimatedChunkSize(totalSize float64) uint64 {
	mb := float64(1024 * 1024)
	gb := float64(mb * 1024)
	minChunkSize := float64(5 * mb)
	maxChunkSize := float64(500 * mb)

	size := minChunkSize

	if totalSize >= float64(500*gb) {
		size = totalSize / 10000
	} else if totalSize >= float64(100*gb) {
		size = totalSize / 5000
	} else if totalSize >= float64(10*gb) {
		size = totalSize / 2500
	} else {
		size = totalSize / 500
	}

	// Size must be within bounds
	size = math.Min(size, maxChunkSize)
	size = math.Max(size, minChunkSize)

	return uint64(math.Ceil(size))
}

// MultipartSuffix pattern describes what APTrust's multipart
// bag suffix looks like. This is for APTrust-specific legacy
// support.
//
// The tar files that make up multipart bags include a suffix
// that follows this pattern. For example, after stripping off
// the .tar suffix, you'll have a name like "my_bag.b04.of12"
var MultipartSuffix = regexp.MustCompile(`\.b\d+\.of\d+$`)

// TarSuffix matches strings that end with .tar
var TarSuffix = regexp.MustCompile(`\.tar$|\.tar.gz$`)

// CleanBagName returns the clean bag name. That's the tar file name minus
// the tar extension and any ".bagN.ofN" suffix.
func CleanBagName(bagName string) string {
	nameMinusTarSuffix := TarSuffix.ReplaceAllString(bagName, "")
	return MultipartSuffix.ReplaceAllString(nameMinusTarSuffix, "")
}

// FindCommonPrefix finds the common prefix in a list of strings.
// This is used during bagging to trim off extraneous characters
// from file paths to help determine what their path should be inside
// the bag.
//
// For example, if we're bagging 100 files that all look like this:
//
// /user/linus/photos/image1.jpg
// /user/linus/photos/image2.jpg
// etc.
//
// We don't want the bag's payload directory to look like this:
//
// data/user/linus/photos/image1.jpg
// data/user/linus/photos/image2.jpg
// etc.
//
// We want it to look like this:
//
// data/image1.jpg
// data/image2.jpg
// etc.
//
// Note that this function has the side effect of sorting the list
// of paths. While side effects are generally undesirable, this one
// is OK in the bagging context as it makes the manifests easy to
// read and tar file structure predictable.
//
// Also note that there may be cases where the list has no common
// prefix. This will happen, for example, if the user is bagging two
// directories like "/var/www" and "images".
func FindCommonPrefix(paths []string) string {
	prefix := make([]rune, 0)
	sort.Strings(paths)
	lastPathSeparator := 0
	first := paths[0]
	last := paths[len(paths)-1]
	for i, rune := range first {
		if i > len(last) {
			break
		}
		if first[i] == last[i] {
			prefix = append(prefix, rune)
			if rune == os.PathSeparator {
				lastPathSeparator = i
			}
		} else {
			break
		}
	}
	if len(prefix) == 0 {
		return ""
	}
	commonPrefix := string(prefix[0:lastPathSeparator]) + string(os.PathSeparator)
	return commonPrefix
}

// ToHumanSize converts a raw byte count (size) to a human-friendly
// representation.
func ToHumanSize(size, unit int64) string {
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "kMGTPE"[exp])
}

// NewLine returns the correct newline character for the current system,
// either "\r\n" or "\n".
func NewLine() string {
	newline := "\n"
	if runtime.GOOS == "windows" {
		newline = "\r\n"
	}
	return newline
}

var reUrl = regexp.MustCompile(`(^http://localhost)|(https?://(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*))`)

// LooksLikeHypertextURL returns true if str looks like an
// HTTP or HTTPS URL.
func LooksLikeHypertextURL(str string) bool {
	return reUrl.MatchString(str)
}

// YesOrNo returns "Yes" if value is true, "No" if value is false.
func YesOrNo(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

// StringToBool converts a string to its boolean value, or returns
// an error if it can't convert the string. In addition to the values
// supported by strconv.ParseBool, this also supports "yes", "y",
// "no" and "n". Those strings are case-insensitive.
func StringToBool(value string) (bool, error) {
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		lcValue := strings.ToLower(value)
		if lcValue == "yes" || lcValue == "y" {
			boolValue = true
			err = nil
		} else if lcValue == "no" || lcValue == "n" {
			boolValue = false
			err = nil
		}
	}
	return boolValue, err
}
