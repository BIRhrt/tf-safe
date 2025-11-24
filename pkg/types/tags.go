package types

import (
	"fmt"
	"regexp"
	"strings"
)

// Tags represents key-value pairs for backup metadata
type Tags map[string]string

// MaxTagKeyLength is the maximum length for a tag key
const MaxTagKeyLength = 128

// MaxTagValueLength is the maximum length for a tag value
const MaxTagValueLength = 256

// MaxTagsPerBackup is the maximum number of tags per backup
const MaxTagsPerBackup = 20

// tagKeyRegex validates tag keys (alphanumeric, hyphens, underscores)
var tagKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// NewTags creates a new Tags instance from a string slice
// Expected format: ["key1=value1", "key2=value2"]
func NewTags(tagStrings []string) (Tags, error) {
	tags := make(Tags)

	for _, tagStr := range tagStrings {
		key, value, err := parseTag(tagStr)
		if err != nil {
			return nil, err
		}
		tags[key] = value
	}

	if err := tags.Validate(); err != nil {
		return nil, err
	}

	return tags, nil
}

// ParseTagString parses a comma-separated tag string
// Format: "key1=value1,key2=value2"
func ParseTagString(tagString string) (Tags, error) {
	if tagString == "" {
		return make(Tags), nil
	}

	tagStrings := strings.Split(tagString, ",")
	return NewTags(tagStrings)
}

// parseTag parses a single tag string
func parseTag(tagStr string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(tagStr), "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid tag format '%s': expected key=value", tagStr)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" {
		return "", "", fmt.Errorf("tag key cannot be empty")
	}

	return key, value, nil
}

// Validate checks if tags conform to constraints
func (t Tags) Validate() error {
	if len(t) > MaxTagsPerBackup {
		return fmt.Errorf("too many tags: %d (max: %d)", len(t), MaxTagsPerBackup)
	}

	for key, value := range t {
		if err := ValidateTagKey(key); err != nil {
			return err
		}
		if err := ValidateTagValue(value); err != nil {
			return err
		}
	}

	return nil
}

// ValidateTagKey validates a tag key
func ValidateTagKey(key string) error {
	if key == "" {
		return fmt.Errorf("tag key cannot be empty")
	}

	if len(key) > MaxTagKeyLength {
		return fmt.Errorf("tag key '%s' exceeds maximum length of %d", key, MaxTagKeyLength)
	}

	if !tagKeyRegex.MatchString(key) {
		return fmt.Errorf("tag key '%s' contains invalid characters (allowed: alphanumeric, hyphen, underscore)", key)
	}

	return nil
}

// ValidateTagValue validates a tag value
func ValidateTagValue(value string) error {
	if len(value) > MaxTagValueLength {
		return fmt.Errorf("tag value exceeds maximum length of %d", MaxTagValueLength)
	}

	return nil
}

// String returns a comma-separated string representation of tags
func (t Tags) String() string {
	if len(t) == 0 {
		return ""
	}

	var parts []string
	for key, value := range t {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(parts, ",")
}

// Matches checks if the tags match ALL the given filter tags
func (t Tags) Matches(filter Tags) bool {
	for filterKey, filterValue := range filter {
		actualValue, exists := t[filterKey]
		if !exists || actualValue != filterValue {
			return false
		}
	}
	return true
}

// Contains checks if the tags contain at least one of the filter tags
func (t Tags) Contains(filter Tags) bool {
	for filterKey, filterValue := range filter {
		if actualValue, exists := t[filterKey]; exists && actualValue == filterValue {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of tags
func (t Tags) Clone() Tags {
	clone := make(Tags, len(t))
	for k, v := range t {
		clone[k] = v
	}
	return clone
}

// Merge merges another Tags map into this one (overwrites on conflict)
func (t Tags) Merge(other Tags) Tags {
	result := t.Clone()
	for k, v := range other {
		result[k] = v
	}
	return result
}
