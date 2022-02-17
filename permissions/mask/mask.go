package permissions

import (
	"errors"
	"fmt"
	"path/filepath"

	"awesome-dragon.science/go/irc/permissions"
	"awesome-dragon.science/go/irc/user"
	"awesome-dragon.science/go/irc/util"
)

var _ permissions.Handler = (*Handler)(nil)

// Mask represents a nick!user@host mask with a set of available permissions
type Mask struct {
	Mask  string
	Perms []string
}

// Handler is an IRC mask based  permission handler
type Handler struct {
	masks []*Mask
}

// ErrNoMatch is returned when there is no matching entry to a given EphemeralUser
var ErrNoMatch = errors.New("no mask matches")

// IsAuthorized implements permissions.Handler
func (h *Handler) IsAuthorised(userToCheck *user.EphemeralUser, requiredPermissions []string) (bool, error) {
	for _, u := range h.masks {
		m, err := filepath.Match(u.Mask, util.UserHostCanonical(userToCheck.UserHost))
		if err != nil {
			return false, fmt.Errorf("invalid user mask %q: %w", u.Mask, err)
		}

		if !m {
			continue
		}

		ok, err := permissions.AllPermissionMatch(u.Perms, requiredPermissions)
		if err != nil {
			return false, fmt.Errorf("could not check permissions for %q (%q): %w", u.Mask, userToCheck.RealHost, err)
		}

		if ok {
			return true, nil
		}
	}

	return false, nil
}
