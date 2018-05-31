package ast

import (
	"errors"
	"github.com/anaminus/luasyntax/go/token"
	"math"
	"strconv"
)

// ParseValue parses the content of the number token and returns the resulting
// value, or an error explaining why the value could not be parsed.
func (e *NumberExpr) ParseValue() (v float64, err error) {
	switch e.NumberToken.Type {
	case token.NUMBERFLOAT:
		// Actual parsing of the number depends on the compiler (strtod), so
		// technically it's correct to just use Go's parser.
		v, err = strconv.ParseFloat(string(e.NumberToken.Bytes), 64)
	case token.NUMBERHEX:
		var i uint64
		// Trim leading `0x`.
		i, err = strconv.ParseUint(string(e.NumberToken.Bytes[2:]), 16, 32)
		v = float64(i)
	default:
		err = errors.New("'" + token.NUMBERFLOAT.String() + "' expected")
	}
	return
}

// FormatValue formats the absolute value of a given number, setting the
// result to the bytes of the token.
//
// When fmt is 'e', 'E', 'f', 'g', or 'G', the number is formatted as a float
// with the NUMBERFLOAT type, and fmt and prec follow the same rules as in
// strconv.ParseFloat.
//
// When fmt is 'd', 'i', or 'u', the number is formatted as a base-10 integer
// with the NUMBERFLOAT type. The prec argument is unused.
//
// When fmt is 'x' or 'X', the number is formatted as a base-16 number with
// the NUMBERHEX type. The prec argument is unused.
//
// When fmt is 0, the format is determined by the current token type, and uses
// the shortest representation of the number. The prec argument is unused.
func (e *NumberExpr) FormatValue(v float64, fmt byte, prec int) {
	if math.Signbit(v) {
		v = -v
	}
	switch fmt {
	case 'e', 'E', 'f', 'g', 'G':
		e.NumberToken.Type = token.NUMBERFLOAT
		e.NumberToken.Bytes = []byte(strconv.FormatFloat(v, fmt, prec, 32))
	case 'd', 'i', 'u':
		e.NumberToken.Type = token.NUMBERFLOAT
		e.NumberToken.Bytes = []byte(strconv.FormatUint(uint64(v), 10))
	case 'x', 'X':
		e.NumberToken.Type = token.NUMBERHEX
		e.NumberToken.Bytes = []byte("0x" + strconv.FormatUint(uint64(v), 16))
	case 0:
		switch e.NumberToken.Type {
		case token.NUMBERFLOAT:
			e.NumberToken.Bytes = []byte(strconv.FormatFloat(v, 'g', -1, 32))
		case token.NUMBERHEX:
			e.NumberToken.Bytes = []byte("0x" + strconv.FormatUint(uint64(uint32(v)), 16))
		default:
			panic("expected number token type")
		}
	default:
		panic("unexpected format")
	}
}

// parseQuotedString parses literal quoted string into actual text.
func parseQuotedString(b []byte) string {
	b = b[1 : len(b)-1]          // Trim quotes.
	c := make([]byte, 0, len(b)) // Result will never be larger than source.
	for i := 0; i < len(b); i++ {
		ch := b[i]
		if ch == '\\' {
			i++
			ch = b[i]
			switch ch {
			case 'a':
				ch = '\a'
			case 'b':
				ch = '\b'
			case 'f':
				ch = '\f'
			case 'n':
				ch = '\n'
			case 'r':
				ch = '\r'
			case 't':
				ch = '\t'
			case 'v':
				ch = '\v'
			default:
				if '0' <= ch && ch <= '9' {
					var n byte
					for j := 0; j < 3 && '0' <= b[i] && b[i] <= '9'; j++ {
						n = n*10 + (b[i] - '0')
						i++
					}
					// Size of number was already checked by scanner.
					ch = n
				}
			}
		}
		c = append(c, ch)
	}
	return string(c)
}

// parseBlockString parses a literal long string into actual text.
func parseBlockString(b []byte) string {
	// Assumes string is wrapped in a `[==[]==]`-like block.
	b = b[1:] // Trim first `[`
	for i, c := range b {
		if c == '[' {
			// Trim to second '[', as well as trailing block.
			b = b[i+1 : len(b)-i-2]
		}
	}
	// Skip first newline.
	if len(b) > 0 && (b[0] == '\n' || b[0] == '\r') {
		if len(b) > 1 && (b[1] == '\n' || b[1] == '\r') && b[1] != b[0] {
			b = b[2:]
		} else {
			b = b[1:]
		}
	}
	return string(b)
}

// ParseValue parses the content of the string token and returns the resulting
// value, or an error explaining why the value could not be parsed.
func (e *StringExpr) ParseValue() (v string, err error) {
	switch e.StringToken.Type {
	case token.STRING:
		v = parseQuotedString(e.StringToken.Bytes)
	case token.LONGSTRING:
		v = parseBlockString(e.StringToken.Bytes)
	default:
		err = errors.New("'" + token.STRING.String() + "' expected")
	}
	return
}

// formatString formats a string in a form suitable to be safely read by a Lua
// interpreter. The result string is enclosed in double quotes, and all double
// quotes, newlines, embedded zeros, and backslashes are properly escaped.
func formatString(src []byte) (dst []byte) {
	if len(src) == 0 {
		return []byte(`""`)
	}

	// Calculate size of dst and allocate.
	size := len(src) + 2
	for _, c := range src {
		switch c {
		case 0, '\n', '"', '\\':
			size++
		}
	}
	dst = make([]byte, size)

	// Fill in dst.
	dst[0] = '"'
	dst[len(dst)-1] = '"'
	h := 0
	for i, j := 0, 1; i < len(src); i, j = i+1, j+1 {
		switch c := src[i]; c {
		case 0:
			c = '0'
			fallthrough
		case '\n', '"', '\\':
			copy(dst[j-(i-h):j], src[h:i])
			h = i + 1
			dst[j] = '\\'
			j++
			dst[j] = c
		}
	}
	copy(dst[len(dst)-1-(len(src)-h):], src[h:])
	return
}

// findFirstNewline scans for the first newline in a text. Returns a slice of
// the newline (which will have a length of 1 or 2) and whether it is at the
// start of the text. Returns an empty slice if no newline was found.
func findFirstNewline(src []byte) ([]byte, bool) {
	for i := 0; i < len(src); i++ {
		if src[i] == '\n' || src[i] == '\r' {
			if i+1 < len(src) && (src[i+1] == '\n' || src[i+1] == '\r') && src[i+1] != src[i] {
				return src[i : i+2], i == 0
			} else {
				return src[i : i+1], i == 0
			}
		}
	}
	return nil, false
}

// formatBlockString formats a string by enclosing it in long brackets.
func formatBlockString(src []byte, newline bool) (dst []byte) {
	if len(src) == 0 {
		return []byte(`[[]]`)
	}

	// Find shortest closing bracket not in the string.
	eq := 0
loop:
	for i := 0; i < len(src); i++ {
		if src[i] == ']' {
			i++
			count := 0
			for ; src[i] == '='; i++ {
				count++
			}
			if src[i] == ']' && count == eq {
				eq++
				goto loop
			}
		}
	}

	// Decide whether a leading newline must be inserted.
	nl, atStart := findFirstNewline(src)
	if len(nl) > 0 && atStart {
		// Newline at start; insert unconditionally.
	} else if newline {
		// Otherwise, insert only if specified.
		if len(nl) == 0 {
			// Ensure that there is a newline to insert.
			nl = []byte{'\n'}
		}
	} else {
		// Insert nothing.
		nl = []byte{}
	}

	// Allocate result.
	dstSize := 1 + eq + 1 + len(src) + 1 + eq + 1 + len(nl) // [==[src]==]\n
	dst = make([]byte, dstSize)

	// Fill in long brackets.
	openBrack := dst[:eq+2]
	openBrack[0] = '['
	openBrack[len(openBrack)-1] = '['
	closeBrack := dst[len(dst)-eq-2:]
	closeBrack[0] = ']'
	closeBrack[len(closeBrack)-1] = ']'
	for i := 1; i < eq+1; i++ {
		openBrack[i] = '='
		closeBrack[i] = '='
	}

	// Fill in newline.
	copy(dst[len(openBrack):], nl)

	// Fill in enclosed string.
	copy(dst[len(openBrack)+len(nl):], src)

	return
}

// FormatValue receives a string and formats it, setting it to the bytes of
// the token. The format is determined by the current token type.
//
// When the token is a STRING, the result is the same as the "%q" verb in
// Lua's string.format.
//
// When the token is a LONGSTRING, the result is enclosed in the shortest
// possible set of long brackets. If the newline argument is true, then the
// result will be formatted with a newline at the start, which is ignored by
// the parser.
func (e *StringExpr) FormatValue(v string, newline bool) {
	switch e.StringToken.Type {
	case token.STRING:
		e.StringToken.Bytes = formatString(e.StringToken.Bytes)
	case token.LONGSTRING:
		e.StringToken.Bytes = formatBlockString(e.StringToken.Bytes, newline)
	default:
		panic("expected string token type")
	}
}

// ParseValue parses the content of the boolean token and returns the
// resulting value, or an error explaining why the value could not be parsed.
func (e *BoolExpr) ParseValue() (v bool, err error) {
	switch e.BoolToken.Type {
	case token.TRUE:
		v = true
	case token.FALSE:
		v = false
	default:
		err = errors.New("boolean expected")
	}
	return
}

// FormatValue receives a boolean and formats it, setting the type and bytes
// of the token.
func (e *BoolExpr) FormatValue(v bool) {
	if v {
		e.BoolToken.Type = token.TRUE
	} else {
		e.BoolToken.Type = token.FALSE
	}
	e.BoolToken.Bytes = []byte(e.BoolToken.Type.String())
}
