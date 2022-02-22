// Package function implements a simple single-function based permission handler
package function

import (
	"awesome-dragon.science/go/irc/permissions"
	"awesome-dragon.science/go/irc/user"
)

var _ permissions.Handler = Handler(nil)

// Handler implements permissions.Handler
type Handler func(userToCheck *user.EphemeralUser, requiredPermissions []string) (bool, error)

// IsAuthorised implements permissions.Handler
func (h Handler) IsAuthorised(userToCheck *user.EphemeralUser, requiredPermissions []string) (bool, error) {
	return h(userToCheck, requiredPermissions)
}
