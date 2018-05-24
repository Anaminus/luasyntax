package scanner

import (
	"github.com/anaminus/luasyntax/go/token"
)

type Error struct {
	Position token.Position
	Message  string
}

func (e Error) Error() string {
	if e.Position.Filename != "" || e.Position.IsValid() {
		return e.Position.String() + ": " + e.Message
	}
	return e.Message
}
