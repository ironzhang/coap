package layer

import "fmt"

type Error struct {
	Layer   string
	Cause   error
	Details string
}

func (e Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %v(%s)", e.Layer, e.Cause, e.Details)
	}
	return fmt.Sprintf("%s: %v", e.Layer, e.Cause)
}
