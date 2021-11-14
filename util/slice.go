package util

// IntSliceContains returns whether or not the given int is found in the given slice
func IntSliceContains(i int, slice []int) bool {
	for _, v := range slice {
		if v == i {
			return true
		}
	}

	return false
}

// StringSliceContains is IntSliceContains, but for strings. I need generics -_-
func StringSliceContains(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}

	return false
}
