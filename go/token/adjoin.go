package token

// AdjoinSeparator returns the character that would allow a token of type
// 'left' to precede a token of type 'right', if placed as a SPACE token
// between the two. Currently, this is '\n' for COMMENTs, and a space for
// everything else.
//
// Returns -1 if the two tokens are allowed to be adjacent. Note that this is
// returned by tokens that will never be adjacent while being syntactically
// correct.
//
// Returns -2 if the allowed adjacency of the two tokens depends on the
// content of the tokens.
func (left Type) AdjoinSeparator(right Type) rune {
	// Tokens must be separated by a space (at minimum).
	const space = ' '
	// Tokens must be separated by a newline.
	const newline = '\n'
	// Tokens are allowed to be adjacent.
	const okay = -1
	// Tokens may or may not be allowed to be adjacent.
	const cond = -2

	switch {
	case left == EOF:
		// While no token would be allowed to appear to the right of an EOF,
		// there is no character that can be inserted to make it correct, so
		// it's allowed.
		return okay
	case !left.IsValid(),
		!right.IsValid():
		// Once again, no character can make these correct, so they're
		// allowed. Undefined types also need to be detected.
		return okay
	case left == COMMENT:
		if right > comm_start {
			// Must return newline so that right token isn't turned into a
			// comment.
			return newline
		}
	case left == NAME:
		switch {
		case right == NAME,
			right.IsNumber(),
			right > key_start:
			return space
		}
	case left.IsNumber():
		switch {
		case right == NAME,
			right.IsNumber(),
			right == DOT,
			right == VARARG,
			right > key_start:
			return space
		}
	case left == ASSIGN:
		switch {
		case right == ASSIGN,
			right == EQ:
			return space
		}
	case left == DOT:
		switch {
		case right.IsNumber(),
			right == DOT,
			right == VARARG,
			right == CONCAT:
			return space
		}
	case left == LBRACK:
		switch {
		case right == LONGSTRING,
			right == LBRACK:
			return space
		}
	case left == CONCAT:
		switch {
		case right == NUMBERFLOAT:
			// Valid only if number does not begin with '.' character.
			return cond
		case right == DOT:
			return space
		}
	case left == LT,
		left == GT:
		switch {
		case right == ASSIGN,
			right == EQ:
			return space
		}
	case left == EQ,
		left == NEQ:
		if right == ASSIGN {
			return space
		}
	case left == MINUS:
		switch {
		case right.IsComment(),
			right == MINUS:
			return space
		}
	case left > key_start:
		switch {
		case right == NAME,
			right > key_start:
			return space
		case right == NUMBERFLOAT:
			if left < ekey_end {
				// Valid only if number does not begin with '.' character.
				return cond
			}
		}
	}
	return okay
}
