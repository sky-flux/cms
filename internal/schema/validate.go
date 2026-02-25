package schema

import "regexp"

var slugRegex = regexp.MustCompile(`^[a-z0-9_-]{3,50}$`)

func ValidateSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}
