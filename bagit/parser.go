package bagit

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ParseTagFile parses the tags in a tag file that conforms to the
// BagIt spec, which uses colon-separated Label:Value tags that may
// span multiple lines. See the BagIt spec at
// https://tools.ietf.org/html/draft-kunze-bagit-17#section-2.2.2
// for more info.
//
// Param reader can be an open file, a buffer, or any other io.Reader.
// If the reader needs to be closed, the user is responsible for closing it.
//
// Param relFilePath is the relative path of the file in the bag. For example,
// "bag-info.txt" or "custom-tags/custom-info.txt".
//
// This returns a slice of Tag objects as parsed from the bag. Note that
// the BagIt spec permits some tags to appear more than once in a file,
// so you may get multiple tags with the same label.
func ParseTagFile(reader io.Reader, relFilePath string) ([]*Tag, error) {
	re := regexp.MustCompile(`^(\S*\:)?(\s*.*)?$`)
	tags := make([]*Tag, 0)
	scanner := bufio.NewScanner(reader)
	var tag *Tag
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if re.MatchString(line) {
			data := re.FindStringSubmatch(line)
			data[1] = strings.TrimSpace(data[1])
			data[1] = strings.Replace(data[1], ":", "", 1)
			if data[1] != "" {
				if tag != nil && tag.TagName != "" {
					tags = append(tags, tag)
				}
				tag = NewTag(relFilePath, data[1], strings.TrimSpace(data[2]))
				continue
			}
			value := strings.TrimSpace(data[2])
			if tag != nil {
				tag.Value = strings.Join([]string{tag.Value, value}, " ")
			}
		} else {
			return nil, fmt.Errorf(
				"Unable to parse tag data in %s line '%s'",
				relFilePath, line)
		}
	}
	// Add file's last tag to the list
	if tag != nil && tag.TagName != "" {
		tags = append(tags, tag)
	}
	// Handle internal scanner errors
	if scanner.Err() != nil {
		return nil, fmt.Errorf("Error reading tag file '%s': %v",
			relFilePath, scanner.Err().Error())
	}
	return tags, nil
}

// ParseManifest parses a checksum manifest, returning a slice of
// Checksums. Param reader should be an open reader. Returns a map
// in which file paths are keys and digests are values.
func ParseManifest(reader io.Reader) (map[string]string, error) {
	checksums := make(map[string]string)
	re := regexp.MustCompile(`^(\S*)\s*(.*)`)
	scanner := bufio.NewScanner(reader)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if re.MatchString(line) {
			data := re.FindStringSubmatch(line)
			digest := data[1]
			filepath := data[2]
			checksums[filepath] = digest
		} else {
			return nil, fmt.Errorf("Unable to parse line %d: %s", lineNum, line)
		}
		lineNum++
	}
	return checksums, nil
}
