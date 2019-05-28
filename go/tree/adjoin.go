package tree

import (
	"github.com/anaminus/luasyntax/go/token"
)

// AdjoinSeparator calls left.Type.AdjoinSeparator(right.Type) to get a
// separating character that allows the left token to precede the right.
//
// Returns -1 if the two tokens are allowed to be adjacent.
//
// When a CONCAT precedes a NUMBERFLOAT, -1 is returned if the bytes of the
// number token do not start with a '.' character, and a space is returned
// otherwise.
//
// When a keyword precedes a NUMBERFLOAT, -1 is returned if the bytes of the
// number token starts with a '.' character, and a space is returned
// otherwise.
func AdjoinSeparator(left, right Token) rune {
	c := left.Type.AdjoinSeparator(right.Type)
	if c == -2 {
		switch {
		case left.Type == token.CONCAT && right.Type == token.NUMBERFLOAT:
			// Adjacency is allowed only if number does not begin with a '.'.
			if len(right.Bytes) == 0 || right.Bytes[0] != '.' {
				return -1
			}
		case left.Type.IsKeyword() && right.Type == token.NUMBERFLOAT:
			// Adjacency is allowed only if number begins with a '.'.
			if len(right.Bytes) > 0 && right.Bytes[0] == '.' {
				return -1
			}
		}
		// Insert space otherwise.
		c = ' '
	}
	return c
}

// adjoinFixer keeps track of the previous token while processing adjoined
// tokens.
type adjoinFixer struct {
	prevToken *Token
}

// Visit implements the Visitor interface.
func (v *adjoinFixer) Visit(Node) Visitor {
	return v
}

// VisitToken implements the TokenVisitor interface.
func (v *adjoinFixer) VisitToken(_ Node, _ int, tok *Token) {
	// Skip invalid tokens.
	if !tok.Type.IsValid() {
		return
	}

	if v.prevToken == nil {
		v.prevToken = tok
		return
	}

	left := *v.prevToken
	var right Token

	// Walk through prefixes as though they were tokens.
	for i := 0; i < len(tok.Prefix); i++ {
		right.Type = tok.Prefix[i].Type
		right.Bytes = tok.Prefix[i].Bytes
		if c := AdjoinSeparator(left, right); c >= 0 {
			if tok.Prefix[i].Type == token.SPACE {
				// Prepend directly to bytes.
				tok.Prefix[i].Bytes = append(tok.Prefix[i].Bytes, 0)
				copy(tok.Prefix[i].Bytes[1:], tok.Prefix[i].Bytes[0:])
				tok.Prefix[i].Bytes[0] = byte(c)
			} else {
				// Insert as SPACE between tokens.
				tok.Prefix = append(tok.Prefix, Prefix{})
				copy(tok.Prefix[i+1:], tok.Prefix[i:])
				tok.Prefix[i].Type = token.SPACE
				tok.Prefix[i].Bytes = []byte{byte(c)}
				i++
			}
		}
		left.Type = tok.Prefix[i].Type
		left.Bytes = tok.Prefix[i].Bytes
	}

	// Handle actual token, which appears after either the token's last
	// prefix, or the previous token.
	if c := AdjoinSeparator(left, *tok); c >= 0 {
		if n := len(tok.Prefix); n > 0 && tok.Prefix[n-1].Type == token.SPACE {
			// Append after last SPACE prefix so that it appears directly
			// before token. This will happen if the SPACE prefix is empty.
			tok.Prefix[n-1].Bytes = append(tok.Prefix[n-1].Bytes, byte(c))
		} else {
			// Insert a SPACE between the last prefix and the token. If there
			// were no prefixes, then the SPACE will be between the previous
			// token and this one.
			tok.Prefix = append(tok.Prefix, Prefix{
				Type:  token.SPACE,
				Bytes: []byte{byte(c)},
			})
		}
	}

	v.prevToken = tok
}

// FixAdjoinedTokens walks through a parse tree and ensures that adjacent tokens
// have the minimum amount of spacing required to prevent them from being parsed
// incorrectly.
func FixAdjoinedTokens(node Node) {
	var v adjoinFixer
	Walk(&v, node)
}
