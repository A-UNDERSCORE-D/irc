package permissions

import (
	"fmt"
	"path/filepath"

	"awesome-dragon.science/go/irc/user"
)

// Oper represents an oper name, and a permission level
type Oper struct {
	Name  string
	Level int
}

// OperHandler is a PermissionHandler that checks for opers and oper names
type OperHandler struct {
	Opers []Oper
}

// GetLevel returns the authorisation level of a given user when compared
// against an Oper. If the user is not an oper at all, it is ignored.
// globs are supported, an oper name of "*" has a fast path
func (o *OperHandler) GetLevel(u *user.EphemeralUser) (int, error) {
	if !u.Oper {
		return Unauthorised, fmt.Errorf("%w: user is not an oper", ErrNoMatch)
	}

	for _, oper := range o.Opers {
		if oper.Name == "*" {
			return oper.Level, nil
		}

		matches, err := filepath.Match(oper.Name, u.OperName)
		if err != nil {
			return Unauthorised, fmt.Errorf("invalid glob: %w", err)
		}

		if matches {
			return oper.Level, nil
		}
	}

	return Unauthorised, ErrNoMatch
}
