package filter

import (
	"fmt"
	"strings"

	"gydnc/model"
)

// FilterOptions defines the options for filtering tags
type FilterOptions struct {
	IncludeTags []string // Tags that must be present (can include wildcards)
	ExcludeTags []string // Tags that must not be present (can include wildcards)
}

// Filter represents a compiled filter that can be applied to entities
type Filter struct {
	options FilterOptions
}

// ParseFilterString parses a simple query syntax into filter options
// Supports formats like:
// "scope:code quality:safety" (include tags)
// "NOT deprecated" or "-deprecated" (exclude tags)
// "scope:* -deprecated" (wildcards and negation)
func ParseFilterString(query string) (FilterOptions, error) {
	options := FilterOptions{}

	if query == "" {
		return options, nil
	}

	// Split the query by spaces
	parts := strings.Fields(query)

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		// Check for NOT operator
		if part == "NOT" && i+1 < len(parts) {
			// Next part after NOT should be negated
			nextPart := parts[i+1]
			options.ExcludeTags = append(options.ExcludeTags, nextPart)

			// Skip the next part since we've processed it
			i++
			continue
		}

		// Handle exclude with dash prefix
		if strings.HasPrefix(part, "-") {
			options.ExcludeTags = append(options.ExcludeTags, part[1:])
			continue
		}

		// Handle include tags
		options.IncludeTags = append(options.IncludeTags, part)
	}

	return options, nil
}

// NewFilter creates a new filter with the given options
func NewFilter(options FilterOptions) *Filter {
	return &Filter{
		options: options,
	}
}

// NewFilterFromString creates a new filter from a query string
func NewFilterFromString(query string) (*Filter, error) {
	options, err := ParseFilterString(query)
	if err != nil {
		return nil, err
	}
	return NewFilter(options), nil
}

// Matches checks if an entity matches this filter
func (f *Filter) Matches(entity model.Entity) bool {
	// Check include tags (entity must have all specified tags)
	for _, tag := range f.options.IncludeTags {
		if !containsTag(entity.Tags, tag) {
			return false
		}
	}

	// Check exclude tags (entity must not have any of these tags)
	for _, tag := range f.options.ExcludeTags {
		if containsTag(entity.Tags, tag) {
			return false
		}
	}

	return true
}

// containsTag checks if the tag list contains the specified tag,
// with support for wildcards (e.g., "scope:*", "foo*", "*bar")
func containsTag(tags []string, searchTag string) bool {
	// Special case: wildcard only means "has at least one tag"
	if searchTag == "*" {
		return len(tags) > 0
	}

	// Check different wildcard patterns
	hasPrefixWildcard := strings.HasPrefix(searchTag, "*") && len(searchTag) > 1
	hasSuffixWildcard := strings.HasSuffix(searchTag, "*") && len(searchTag) > 1 && !strings.HasSuffix(searchTag, ":*")
	hasNamespaceWildcard := strings.HasSuffix(searchTag, ":*")

	// If no wildcards, do exact match
	if !hasPrefixWildcard && !hasSuffixWildcard && !hasNamespaceWildcard {
		for _, tag := range tags {
			if tag == searchTag {
				return true
			}
		}
		return false
	}

	// Handle wildcard patterns
	for _, tag := range tags {
		if hasNamespaceWildcard {
			// Namespace wildcard: "foo:*" matches anything with "foo:" prefix
			// Get the namespace prefix (everything up to the :*)
			prefix := searchTag[:len(searchTag)-1] // Remove the * but keep the :
			// Make sure it's an exact namespace match by checking for prefix and for ":" at the right position
			if strings.HasPrefix(tag, prefix) {
				return true
			}
		} else if hasPrefixWildcard {
			// Prefix wildcard: "*bar" matches anything ending with "bar"
			suffix := searchTag[1:] // Remove the *
			if strings.HasSuffix(tag, suffix) {
				return true
			}
		} else if hasSuffixWildcard {
			// Suffix wildcard: "foo*" matches anything starting with "foo"
			prefix := searchTag[:len(searchTag)-1] // Remove the *
			if strings.HasPrefix(tag, prefix) {
				return true
			}
		}
	}

	return false
}

// Filter applies the filter to a slice of entities and returns only the matching ones
func (f *Filter) Filter(entities []model.Entity) []model.Entity {
	var filtered []model.Entity

	for _, entity := range entities {
		if f.Matches(entity) {
			filtered = append(filtered, entity)
		}
	}

	return filtered
}

// ApplyFilter applies a filter string to a list of entities
func ApplyFilter(entities []model.Entity, filterString string) ([]model.Entity, error) {
	if filterString == "" {
		return entities, nil
	}

	filter, err := NewFilterFromString(filterString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filter string: %w", err)
	}

	return filter.Filter(entities), nil
}
