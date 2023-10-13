package util_test

import (
	"fmt"
	"html/template"
	"testing"
	"time"

	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDate, _ = time.Parse(time.RFC3339, "2021-04-16T12:24:16Z")
var textString = "The Academic Preservation Trust (APTrust) is committed to the creation and management of a sustainable environment for digital preservation."
var truncatedString = "The Academic Preservation Trust..."

func TestTruncate(t *testing.T) {
	assert.Equal(t, truncatedString, util.Truncate(textString, 31))
	assert.Equal(t, "hello", util.Truncate("hello", 80))
}

func TestDateUS(t *testing.T) {
	assert.Equal(t, "Apr 16, 2021", util.DateUS(testDate))
	assert.Equal(t, "", util.DateUS(time.Time{}))
}

func TestDateTimeUS(t *testing.T) {
	assert.Equal(t, "Apr 16, 2021 12:24:16", util.DateTimeUS(testDate))
	assert.Equal(t, "", util.DateUS(time.Time{}))
}

func TestDateISO(t *testing.T) {
	assert.Equal(t, "2021-04-16", util.DateISO(testDate))
	assert.Equal(t, "", util.DateISO(time.Time{}))
}

func TestDateTimeISO(t *testing.T) {
	assert.Equal(t, "2021-04-16T12:24:16Z", util.DateTimeISO(testDate))
	assert.Equal(t, "", util.DateTimeISO(time.Time{}))
}

func TestStrEq(t *testing.T) {
	assert.True(t, util.StrEq("4", int8(4)))
	assert.True(t, util.StrEq("200", int16(200)))
	assert.True(t, util.StrEq("200", int32(200)))
	assert.True(t, util.StrEq("200", int64(200)))

	assert.True(t, util.StrEq("true", true))
	assert.True(t, util.StrEq("true", "true"))
	assert.True(t, util.StrEq(true, true))
	assert.True(t, util.StrEq(true, "true"))
	assert.True(t, util.StrEq(false, "false"))

	assert.False(t, util.StrEq("true", false))
	assert.False(t, util.StrEq("200", 909))
}

func TestStrStartsWith(t *testing.T) {
	assert.True(t, util.StrStartsWith("alligator", "all"))
	assert.False(t, util.StrStartsWith("crocodile", "all"))
}

func TestEscapeAttr(t *testing.T) {
	assert.Equal(t, template.HTMLAttr("O'Blivion's"), util.EscapeAttr("O'Blivion's"))
}

func TestEscapeHTML(t *testing.T) {
	assert.Equal(t, template.HTML("<em>escape!</em>"), util.EscapeHTML("<em>escape!</em>"))
}

func TestHumanSize(t *testing.T) {
	assert.Equal(t, "2.0 kB", util.HumanSize(2*1024))
	assert.Equal(t, "2.0 MB", util.HumanSize(2*1024*1024))
	assert.Equal(t, "2.0 GB", util.HumanSize(2*1024*1024*1024))
	assert.Equal(t, "2.0 TB", util.HumanSize(2*1024*1024*1024*1024))
}

func TestFileIconFor(t *testing.T) {
	// Should return item defined in map
	assert.Equal(
		t,
		template.HTML(util.FileIconMap[".pdf"]),
		util.FileIconFor("/Users/josie/files/usermanual.pdf"))

	// If item is not defined in map, should return IconMissing
	assert.Equal(
		t,
		template.HTML(util.FileIconGeneric),
		util.FileIconFor("** missing **"))
}

var longString = "Somewhere in la Mancha, in a place whose name I do not care to remember, a gentleman lived not long ago, one of those who has a lance and ancient shield on a shelf and keeps a skinny nag and a greyhound for racing."

func TestTruncateMiddle(t *testing.T) {
	assert.Equal(t, "Somewher... racing.", util.TruncateMiddle(longString, 20))
	assert.Equal(t, "Somewhere in ...d for racing.", util.TruncateMiddle(longString, 30))
	assert.Equal(t, longString, util.TruncateMiddle(longString, 500))
}

func TestTruncateStart(t *testing.T) {
	assert.Equal(t, "...a greyhound for racing.", util.TruncateStart(longString, 20))
	assert.Equal(t, "...y nag and a greyhound for racing.", util.TruncateStart(longString, 30))
	assert.Equal(t, longString, util.TruncateStart(longString, 500))
	assert.Equal(t, longString, util.TruncateStart(longString, 5000))
}

func TestDict(t *testing.T) {
	expected := map[string]interface{}{
		"key1": 1,
		"key2": "two",
	}
	dict, err := util.Dict("key1", 1, "key2", "two")
	require.Nil(t, err)
	assert.Equal(t, expected, dict)

	_, err = util.Dict("key1", 1, "key2")
	assert.Error(t, fmt.Errorf("invalid parameter length: should be an even number"), err)

	_, err = util.Dict(1, "key2")
	assert.Error(t, fmt.Errorf("wrong data type: key '1' should be a string"), err)
}

func TestDisplayDate(t *testing.T) {
	timestamp := time.Date(2023, time.August, 16, 8, 16, 0, 0, time.UTC)
	assert.Equal(t, "16 Aug 23 08:16 UTC", util.DisplayDate(timestamp))
	assert.Empty(t, util.DisplayDate(time.Time{}))
}
