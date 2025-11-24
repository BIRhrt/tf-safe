package types

import (
	"testing"
)

func TestNewTags(t *testing.T) {
	tests := []struct {
		name      string
		tagStrs   []string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "valid tags",
			tagStrs:   []string{"env=prod", "version=v1.0"},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "empty tags",
			tagStrs:   []string{},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:    "invalid format - no equals",
			tagStrs: []string{"env"},
			wantErr: true,
		},
		{
			name:    "invalid format - empty key",
			tagStrs: []string{"=value"},
			wantErr: true,
		},
		{
			name:    "invalid key - special chars",
			tagStrs: []string{"env@prod=value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := NewTags(tt.tagStrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(tags) != tt.wantCount {
				t.Errorf("NewTags() got %d tags, want %d", len(tags), tt.wantCount)
			}
		})
	}
}

func TestParseTagString(t *testing.T) {
	tests := []struct {
		name     string
		tagStr   string
		wantErr  bool
		wantTags Tags
	}{
		{
			name:    "single tag",
			tagStr:  "env=prod",
			wantErr: false,
			wantTags: Tags{
				"env": "prod",
			},
		},
		{
			name:    "multiple tags",
			tagStr:  "env=prod,version=v1.0,team=platform",
			wantErr: false,
			wantTags: Tags{
				"env":     "prod",
				"version": "v1.0",
				"team":    "platform",
			},
		},
		{
			name:     "empty string",
			tagStr:   "",
			wantErr:  false,
			wantTags: Tags{},
		},
		{
			name:    "whitespace handling",
			tagStr:  "env = prod , version = v1.0",
			wantErr: false,
			wantTags: Tags{
				"env":     "prod",
				"version": "v1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := ParseTagString(tt.tagStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTagString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for k, v := range tt.wantTags {
					if tags[k] != v {
						t.Errorf("ParseTagString() tag %s = %s, want %s", k, tags[k], v)
					}
				}
			}
		})
	}
}

func TestTagsValidate(t *testing.T) {
	tests := []struct {
		name    string
		tags    Tags
		wantErr bool
	}{
		{
			name: "valid tags",
			tags: Tags{
				"env":     "production",
				"version": "v1.0",
				"team":    "platform",
			},
			wantErr: false,
		},
		{
			name:    "empty tags",
			tags:    Tags{},
			wantErr: false,
		},
		{
			name: "too many tags",
			tags: func() Tags {
				t := make(Tags)
				for i := 0; i < MaxTagsPerBackup+1; i++ {
					t[string(rune('a'+i))] = "value"
				}
				return t
			}(),
			wantErr: true,
		},
		{
			name: "tag key too long",
			tags: Tags{
				string(make([]byte, MaxTagKeyLength+1)): "value",
			},
			wantErr: true,
		},
		{
			name: "tag value too long",
			tags: Tags{
				"key": string(make([]byte, MaxTagValueLength+1)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tags.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Tags.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagsMatches(t *testing.T) {
	tags := Tags{
		"env":     "prod",
		"version": "v1.0",
		"team":    "platform",
	}

	tests := []struct {
		name   string
		filter Tags
		want   bool
	}{
		{
			name:   "exact match single tag",
			filter: Tags{"env": "prod"},
			want:   true,
		},
		{
			name:   "exact match multiple tags",
			filter: Tags{"env": "prod", "team": "platform"},
			want:   true,
		},
		{
			name:   "no match - wrong value",
			filter: Tags{"env": "staging"},
			want:   false,
		},
		{
			name:   "no match - missing key",
			filter: Tags{"region": "us-west-2"},
			want:   false,
		},
		{
			name:   "empty filter matches all",
			filter: Tags{},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tags.Matches(tt.filter); got != tt.want {
				t.Errorf("Tags.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagsContains(t *testing.T) {
	tags := Tags{
		"env":     "prod",
		"version": "v1.0",
		"team":    "platform",
	}

	tests := []struct {
		name   string
		filter Tags
		want   bool
	}{
		{
			name:   "contains single match",
			filter: Tags{"env": "prod"},
			want:   true,
		},
		{
			name:   "contains partial match",
			filter: Tags{"env": "prod", "region": "us-west-2"},
			want:   true,
		},
		{
			name:   "no match",
			filter: Tags{"env": "staging", "region": "us-west-2"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tags.Contains(tt.filter); got != tt.want {
				t.Errorf("Tags.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagsClone(t *testing.T) {
	original := Tags{
		"env":  "prod",
		"team": "platform",
	}

	clone := original.Clone()

	// Modify clone
	clone["env"] = "staging"
	clone["new"] = "value"

	// Original should be unchanged
	if original["env"] != "prod" {
		t.Errorf("Clone modified original: env = %s", original["env"])
	}
	if _, exists := original["new"]; exists {
		t.Errorf("Clone modified original: new key exists")
	}
}

func TestTagsMerge(t *testing.T) {
	tags1 := Tags{
		"env":  "prod",
		"team": "platform",
	}

	tags2 := Tags{
		"env":     "staging", // Should overwrite
		"version": "v2.0",    // Should add
	}

	merged := tags1.Merge(tags2)

	expected := Tags{
		"env":     "staging",
		"team":    "platform",
		"version": "v2.0",
	}

	for k, v := range expected {
		if merged[k] != v {
			t.Errorf("Merge failed: %s = %s, want %s", k, merged[k], v)
		}
	}

	// Original should be unchanged
	if tags1["env"] != "prod" {
		t.Errorf("Merge modified original")
	}
}

func TestValidateTagKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid alphanumeric", "env123", false},
		{"valid with hyphen", "my-env", false},
		{"valid with underscore", "my_env", false},
		{"empty key", "", true},
		{"too long", string(make([]byte, MaxTagKeyLength+1)), true},
		{"invalid chars - space", "my env", true},
		{"invalid chars - dot", "my.env", true},
		{"invalid chars - slash", "my/env", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTagKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTagKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTagValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid value", "production", false},
		{"empty value", "", false},
		{"value with spaces", "my value", false},
		{"too long", string(make([]byte, MaxTagValueLength+1)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTagValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTagValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
