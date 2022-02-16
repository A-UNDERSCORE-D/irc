package chatcommand

type (
	// Callback type used by chatcommand.Handler
	Callback func(*Argument) error
	command  struct {
		name                string
		help                string
		requiredArgs        int
		requiredPermissions []string
		callback            func(*Argument) error
	}
)
