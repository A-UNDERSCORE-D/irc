package event

import "strings"

// MultiError contains multiple errors all of which occurred during an OnMessage call.
type MultiError struct {
	Errors []error
}

func (m *MultiError) Error() string {
	strErrors := []string{}
	for _, e := range m.Errors {
		strErrors = append(strErrors, e.Error())
	}

	if len(strErrors) == 0 {
		return "No errors?!"
	}

	return "Multiple errors: " + strings.Join(strErrors, ",")
}
