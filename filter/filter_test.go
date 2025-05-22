package filter

import (
	"reflect"
	"testing"

	"gydnc/model"
)

func TestParseFilterString(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected FilterOptions
	}{
		{
			name:  "Empty query",
			query: "",
			expected: FilterOptions{
				IncludeTags: nil,
				ExcludeTags: nil,
			},
		},
		{
			name:  "Simple tag filter",
			query: "scope:code",
			expected: FilterOptions{
				IncludeTags: []string{"scope:code"},
				ExcludeTags: nil,
			},
		},
		{
			name:  "Multiple tag filters",
			query: "scope:code quality:safety",
			expected: FilterOptions{
				IncludeTags: []string{"scope:code", "quality:safety"},
				ExcludeTags: nil,
			},
		},
		{
			name:  "Exclude tag filter with dash",
			query: "-deprecated",
			expected: FilterOptions{
				IncludeTags: nil,
				ExcludeTags: []string{"deprecated"},
			},
		},
		{
			name:  "Exclude tag filter with NOT keyword",
			query: "NOT deprecated",
			expected: FilterOptions{
				IncludeTags: nil,
				ExcludeTags: []string{"deprecated"},
			},
		},
		{
			name:  "Mix of tag filters with dash",
			query: "scope:code -deprecated",
			expected: FilterOptions{
				IncludeTags: []string{"scope:code"},
				ExcludeTags: []string{"deprecated"},
			},
		},
		{
			name:  "Mix of tag filters with NOT",
			query: "scope:code NOT deprecated",
			expected: FilterOptions{
				IncludeTags: []string{"scope:code"},
				ExcludeTags: []string{"deprecated"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options, err := ParseFilterString(tt.query)
			if err != nil {
				t.Fatalf("ParseFilterString() error = %v", err)
			}

			if !reflect.DeepEqual(options.IncludeTags, tt.expected.IncludeTags) {
				t.Errorf("IncludeTags = %v, want %v", options.IncludeTags, tt.expected.IncludeTags)
			}

			if !reflect.DeepEqual(options.ExcludeTags, tt.expected.ExcludeTags) {
				t.Errorf("ExcludeTags = %v, want %v", options.ExcludeTags, tt.expected.ExcludeTags)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	// Create some test entities
	entities := []model.Entity{
		{
			Alias:         "entity1",
			Title:         "Entity One",
			Description:   "This is entity one",
			SourceBackend: "default",
			Tags:          []string{"scope:code", "quality:safety"},
			CustomMetadata: map[string]interface{}{
				"type": "behavior",
				"tier": "must",
			},
		},
		{
			Alias:         "entity2",
			Title:         "Entity Two",
			Description:   "This is entity two",
			SourceBackend: "secondary",
			Tags:          []string{"scope:docs", "quality:clarity"},
			CustomMetadata: map[string]interface{}{
				"type": "recipe",
				"tier": "should",
			},
		},
		{
			Alias:         "entity3",
			Title:         "Deprecated Entity",
			Description:   "This is a deprecated entity",
			SourceBackend: "default",
			Tags:          []string{"scope:code", "deprecated"},
			CustomMetadata: map[string]interface{}{
				"type": "behavior",
				"tier": "should",
			},
		},
		{
			Alias:         "entity4",
			Title:         "Feature Entity",
			Description:   "This is a feature entity",
			SourceBackend: "default",
			Tags:          []string{"feature:wizard", "feature:awesome"},
			CustomMetadata: map[string]interface{}{
				"type": "recipe",
				"tier": "should",
			},
		},
	}

	tests := []struct {
		name     string
		query    string
		expected []model.Entity
	}{
		{
			name:     "No filter",
			query:    "",
			expected: entities,
		},
		{
			name:     "Filter by tag",
			query:    "scope:code",
			expected: []model.Entity{entities[0], entities[2]},
		},
		{
			name:     "Filter by exclude tag",
			query:    "-deprecated",
			expected: []model.Entity{entities[0], entities[1], entities[3]},
		},
		{
			name:     "Multiple include tags",
			query:    "scope:code quality:safety",
			expected: []model.Entity{entities[0]},
		},
		{
			name:     "Include and exclude tags",
			query:    "scope:code -deprecated",
			expected: []model.Entity{entities[0]},
		},
		{
			name:     "Filter with no matches",
			query:    "nonexistent",
			expected: []model.Entity{},
		},
		// Tests for wildcard patterns
		{
			name:     "Wildcard suffix",
			query:    "scope:*",
			expected: []model.Entity{entities[0], entities[1], entities[2]},
		},
		{
			name:     "Wildcard prefix",
			query:    "*wizard",
			expected: []model.Entity{entities[3]},
		},
		{
			name:     "Exclude with suffix wildcard",
			query:    "-scope:*",
			expected: []model.Entity{entities[3]},
		},
		{
			name:     "Exclude with NOT and suffix wildcard",
			query:    "NOT feature:*",
			expected: []model.Entity{entities[0], entities[1], entities[2]},
		},
		{
			name:     "Exclude with prefix wildcard",
			query:    "-*wizard",
			expected: []model.Entity{entities[0], entities[1], entities[2]},
		},
		{
			name:     "Complex filter with wildcards and negation",
			query:    "scope:* -deprecated",
			expected: []model.Entity{entities[0], entities[1]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewFilterFromString(tt.query)
			if err != nil {
				t.Fatalf("NewFilterFromString() error = %v", err)
			}

			filtered := filter.Filter(entities)

			// For empty slices, special handling to avoid comparison issues
			if len(filtered) == 0 && len(tt.expected) == 0 {
				return // Both are empty, test passes
			}

			if !reflect.DeepEqual(filtered, tt.expected) {
				t.Errorf("Filter() = %v, want %v", filtered, tt.expected)
			}
		})
	}
}

func TestContainsTag(t *testing.T) {
	tags := []string{"scope:code", "quality:safety", "domain:api", "feature:wizard", "backend:localfs"}

	tests := []struct {
		name      string
		searchTag string
		expected  bool
	}{
		{
			name:      "Exact match",
			searchTag: "scope:code",
			expected:  true,
		},
		{
			name:      "No match",
			searchTag: "scope:docs",
			expected:  false,
		},
		{
			name:      "Wildcard any tag",
			searchTag: "*",
			expected:  true,
		},
		{
			name:      "Namespace wildcard",
			searchTag: "scope:*",
			expected:  true,
		},
		{
			name:      "Namespace wildcard no match",
			searchTag: "topic:*",
			expected:  false,
		},
		{
			name:      "Prefix wildcard match",
			searchTag: "*wizard",
			expected:  true,
		},
		{
			name:      "Prefix wildcard no match",
			searchTag: "*unknown",
			expected:  false,
		},
		{
			name:      "Suffix wildcard match",
			searchTag: "feature:*",
			expected:  true,
		},
		{
			name:      "Suffix wildcard partial match",
			searchTag: "back*",
			expected:  true,
		},
		{
			name:      "Suffix wildcard no match",
			searchTag: "unknown*",
			expected:  false,
		},
		{
			name:      "Mixed case match should fail",
			searchTag: "SCOPE:code",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsTag(tags, tt.searchTag); got != tt.expected {
				t.Errorf("containsTag() = %v, want %v", got, tt.expected)
			}
		})
	}

	// Test with empty tags
	emptyTags := []string{}
	if containsTag(emptyTags, "*") {
		t.Errorf("containsTag() with empty tags and wildcard should return false")
	}
}
