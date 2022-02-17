package oper

import (
	"fmt"
	"path/filepath"

	"awesome-dragon.science/go/irc/permissions"
	"awesome-dragon.science/go/irc/user"
)

var _ permissions.Handler = (*Handler)(nil)

// Oper represents an oper name, and a permission level
type Oper struct {
	Name        string
	Permissions []string
}

// Handler is a PermissionHandler that checks for opers and oper names
type Handler struct {
	Opers []Oper
}

// IsAuthorised implements Handler.IsAuthorized.
// It checks against the users oper name and nothing else. Configured oper names are globbed.
func (h *Handler) IsAuthorised(u *user.EphemeralUser, requiredPermissions []string) (bool, error) {
	if len(requiredPermissions) == 0 {
		return true, nil
	}

	if !u.Oper { // quick path if any permission is required, and the target is not an oper
		return false, nil
	}

	for _, o := range h.Opers {
		m, err := filepath.Match(o.Name, u.OperName)
		if err != nil {
			return false, fmt.Errorf("invalid oper mask %s: %w", o.Name, err)
		}

		if !m {
			continue
		}

		ok, err := permissions.AllPermissionMatch(o.Permissions, requiredPermissions)
		if err != nil {
			return false, fmt.Errorf("could not check permissions for %q (%q): %w", o.Name, u.OperName, err)
		}

		if ok {
			return true, nil
		}
	}

	return false, nil
}
