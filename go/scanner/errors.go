package scanner

import (
	"github.com/anaminus/luasyntax/go/token"
)

// An Error indicates an error within a file.
type Error struct {
	// Position points to the location of the error.
	Position token.Position
	// Message describes the condition of the error.
	Message string
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.Position.Filename != "" || e.Position.IsValid() {
		return e.Position.String() + ": " + e.Message
	}
	return e.Message
}
