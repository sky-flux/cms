package templates

import "strconv"

// itoa converts an int to string for use in Templ expressions.
func itoa(n int) string {
	return strconv.Itoa(n)
}
