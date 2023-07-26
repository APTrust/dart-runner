package util

import (
	"fmt"
	"html/template"
	"path"
	"strings"
	"time"
)

// Dict returns an interface map suitable for passing into
// sub templates.
func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("invalid parameter length: should be an even number")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("wrong data type: key '%v' should be a string", values[i])
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

// DisplayDate returns a datetime in human-readable format.
// This returns an empty string if time is empty.
func DisplayDate(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.Format(time.RFC822)
}

// Truncate truncates the value to the given length, appending
// an ellipses to the end. If value contains HTML tags, they
// will be stripped because truncating HTML can result in unclosed
// tags that will ruin the page layout.
func Truncate(value string, length int) string {
	if len(value) < length {
		return value
	}
	fmtStr := fmt.Sprintf("%%.%ds...", length)
	return fmt.Sprintf(fmtStr, value)
}

// DateUS returns a date in format "Jan 2, 2006"
func DateUS(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format("Jan _2, 2006")
}

// DateUS returns a date in format "Jan 2, 2006 15:04:05"
func DateTimeUS(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format("Jan _2, 2006 15:04:05")
}

// DateISO returns a date in format "2006-01-02"
func DateISO(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format("2006-01-02")
}

// DateTimeISO returns a date in format "2006-01-02T15:04:05Z"
func DateTimeISO(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format(time.RFC3339)
}

// UnixToISO converts a Unix timestamp to ISO format.
func UnixToISO(ts int64) string {
	return time.Unix(ts, 0).Format(time.RFC3339)
}

// YesNo returns "Yes" if value is true, "No" if value is false.
func YesNo(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

// StrEq compares the string representation of two values and returns
// true if they are equal.
func StrEq(val1, val2 interface{}) bool {
	str1 := fmt.Sprintf("%v", val1)
	str2 := fmt.Sprintf("%v", val2)
	return str1 == str2
}

// EscapeAttr escapes an HTML attribute value.
// This helps avoid the ZgotmplZ problem.
func EscapeAttr(s string) template.HTMLAttr {
	return template.HTMLAttr(s)
}

// EscapeHTML returns an escaped HTML string.
// This helps avoid the ZgotmplZ problem.
func EscapeHTML(s string) template.HTML {
	return template.HTML(s)
}

// HumanSize returns a number of bytes in a human-readable format.
// Note that we use base 1024, not base 1000, because AWS uses 1024
// to calculate the storage size of the objects we're reporting on.
func HumanSize(size int64) string {
	return ToHumanSize(size, 1024)
}

// FileIconFor returns a FontAwesome icon for the specified file type,
// as defined in util.FileIconMap. If the icon map has no entry for this type,
// this returns util.FileIconGeneric.
func FileIconFor(filepath string) template.HTML {
	fileExtention := strings.ToLower(path.Ext(filepath))
	icon := FileIconMap[fileExtention]
	if icon == "" {
		icon = FileIconGeneric
	}
	return template.HTML(icon)
}

// TruncateMiddle trims str to maxLen by removing them from the
// middle of the string. It adds dots to the middle of the string
// to indicate text was trimmed.
func TruncateMiddle(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	half := (maxLen - 3) / 2
	end := len(str) - half
	return str[0:half] + "..." + str[end:]
}

// TruncateStart trims str to maxLen by removing them from the
// start of the string. It adds leading dots to indicate some
// text was trimmed.
func TruncateStart(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	end := (len(str) - 3) - maxLen
	if end < 0 {
		end = 0
	}
	return "..." + str[end:]
}
