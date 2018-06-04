package ast

import (
	"github.com/anaminus/luasyntax/go/token"
)

// adjoinFixer keeps track of the previous token while processing adjoined
// tokens.
type adjoinFixer struct {
	prevToken *Token
}

// Visit implements the Visitor interface.
func (v *adjoinFixer) Visit(Node) Visitor {
	return v
}

// getSep returns the character that separates the two adjoined token types,
// handing the cases where the result depends on the content of the tokens.
func (v *adjoinFixer) getSep(lt, rt token.Type, rb []byte) rune {
	c := lt.AdjoinSeparator(rt)
	if c == -2 {
		switch {
		case lt == token.CONCAT && rt == token.NUMBERFLOAT:
			// Adjacency is allowed only if number does not begin with a '.'.
			if len(rb) == 0 || rb[0] != '.' {
				return -1
			}
		case lt.IsKeyword() && rt == token.NUMBERFLOAT:
			// Adjacency is allowed only if number begins with a '.'.
			if len(rb) > 0 && rb[0] == '.' {
				return -1
			}
		}
		// Insert space otherwise.
		c = ' '
	}
	return c
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

	prevType := v.prevToken.Type

	// Walk through prefixes as though they were tokens.
	for i := 0; i < len(tok.Prefix); i++ {
		if c := v.getSep(prevType, tok.Prefix[i].Type, tok.Prefix[i].Bytes); c >= 0 {
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
		prevType = tok.Prefix[i].Type
	}

	// Handle actual token, which appears after either the token's last
	// prefix, or the previous token.
	if c := v.getSep(prevType, tok.Type, tok.Bytes); c >= 0 {
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

// FixAdjoinedTokens walks through a syntax tree and ensures that adjacent
// tokens have the minimum amount of spacing required to prevent them from
// being parsed incorrectly.
func FixAdjoinedTokens(node Node) {
	var v adjoinFixer
	Walk(&v, node)
}
