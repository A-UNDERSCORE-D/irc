package permissions

import "awesome-dragon.science/go/irc/user"

const Unauthorised = -1

// Handler represents any permission system
type Handler interface {
	GetLevel(user *user.EphemeralUser) (int, string)
}
