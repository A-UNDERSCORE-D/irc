package permissions

import (
	"errors"
	"fmt"
	"path/filepath"

	"awesome-dragon.science/go/irc/user"
	"awesome-dragon.science/go/irc/util"
)

type SimpleMask struct {
	Mask  string
	Level int
}

// SimpleHandler is an IRC mask based  permission handler
type SimpleHandler struct {
	masks []SimpleMask
}

// ErrNoMatch is returned when there is no matching entry to a given EphemeralUser
var ErrNoMatch = errors.New("no mask matches")

// GetLevel returns the level on the first mask that matches, or an error indicating no match
func (s *SimpleHandler) GetLevel(userHost *user.EphemeralUser) (int, error) {
	for _, mask := range s.masks {
		matches, err := filepath.Match(mask.Mask, util.UserHostCanonical(userHost.UserHost))
		if err != nil {
			return Unauthorised, fmt.Errorf("could not match against mask: %w", err)
		}

		if matches {
			return mask.Level, nil
		}
	}

	return Unauthorised, fmt.Errorf("%w: %q", ErrNoMatch, util.UserHostCanonical(userHost.UserHost))
}
