package chatcommand

import "strings"

type Argument struct {
	CommandName string
	Argument    string
}

func (a *Argument) Split() []string {
	if len(strings.TrimSpace(a.Argument)) == 0 {
		return nil
	}

	return strings.Split(a.Argument, " ")
}
