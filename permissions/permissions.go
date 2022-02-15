package permissions

import (
	"fmt"
	"path/filepath"

	"awesome-dragon.science/go/irc/user"
)

// Unauthorised is the level returned when a user has no special levels
const Unauthorised = -1

// Handler represents any string-based permission handler
type Handler interface {
	IsAuthorised(userToCheck *user.EphemeralUser, requiredPermissions []string) (bool, error)
}

// AnyPermissionMatch returns true if any pair of required and available match
func AnyPermissionMatch(available, required []string) (bool, error) {
	if len(required) == 0 {
		return true, nil
	}

	for _, r := range required {
		for _, a := range available {
			if a == r {
				return true, nil
			}

			match, err := filepath.Match(a, r)
			if err != nil {
				return false, fmt.Errorf("could not glob string %q and %q: %w", a, r, err)
			}

			if match {
				return true, nil
			}
		}
	}

	return false, nil
}

// AllPermissionMatch returns true if and only if every entry in `required` is met by an entry in `available`
func AllPermissionMatch(available, required []string) (bool, error) {
outer:
	for _, r := range required {
		for _, a := range available {
			match, err := filepath.Match(a, r)
			if err != nil {
				return false, fmt.Errorf("could not glob string %q and %q: %w", a, r, err)
			}

			if match || a == r {
				continue outer
			}
		}

		return false, nil
	}

	// Only possible if every single entry in r was matched
	return true, nil
}
