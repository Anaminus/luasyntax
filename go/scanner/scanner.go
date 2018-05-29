// The scanner package implements a tokenizer for Lua source text.
package scanner

import (
	"bytes"
	"github.com/anaminus/luasyntax/go/token"
)

const eof = -1

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\v' || ch == '\f'
}

// An ErrorHandler may be provided to Scanner.Init. If an error occurs while
// there is a handler, the handler is called with the position of the
// offending token and an error message.
type ErrorHandler func(pos token.Position, msg string)

// Scanner holds the scanner's state while processing a source file. It must
// be initialized with Init before using.
type Scanner struct {
	file *token.File
	src  []byte
	err  ErrorHandler

	ch         rune // Current character.
	offset     int  // Offset of current character.
	rdOffset   int  // Offset of next character to read.
	lineOffset int  // Offset of the current line.

	// ErrorCount is the number of errors encountered by the scanner.
	ErrorCount int
}

// next scans the next character, updating ch and offset, and tracking any new
// lines encountered.
func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = rune(s.src[s.rdOffset])
		s.rdOffset += 1
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = eof
	}
}

// Init prepares the scanner to tokenize a given source. The file argument
// sets the file to use for position information, and src sets the source to
// tokenize. The option err argument is used to handle errors.
func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler) {
	s.file = file
	s.src = src
	s.err = err

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0

	s.next()
}

// error calls the error handler, if available, and increments ErrorCount.
func (s *Scanner) error(off int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(off), msg)
	}
	s.ErrorCount++
}

// scanSpace scans for a sequence of space characters.
func (s *Scanner) scanSpace() {
	for isSpace(s.ch) {
		s.next()
	}
}

// scanName scans for a name of the form `[A-Za-z_][0-9A-Za-z_]*`.
func (s *Scanner) scanName() []byte {
	off := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return s.src[off:s.offset]
}

// scanNumber scans for a number. The result may not evaluate to a valid
// number.
func (s *Scanner) scanNumber() token.Type {
	off := s.offset
	for isDigit(s.ch) || s.ch == '.' {
		s.next()
	}
	if s.ch == 'e' || s.ch == 'E' {
		s.next()
		if s.ch == '+' || s.ch == '-' {
			s.next()
		}
	}
	s.scanName()
	if bytes.HasPrefix(s.src[off:s.offset], []byte{'0', 'x'}) {
		return token.NUMBERHEX
	}
	return token.NUMBERFLOAT
}

// scanString scans for a quoted string.
func (s *Scanner) scanString(off int) {
	quote := s.ch
	s.next()
	for s.ch != quote {
		switch s.ch {
		case eof:
			s.error(off, "unfinished string (EOF)")
			return
		case '\n', '\r':
			s.error(off, "unfinished string (EOL)")
			return
		case '\\':
			s.next()
			switch s.ch {
			case '\n', '\r':
				c := s.ch
				s.next()
				if (s.ch == '\n' || s.ch == '\r') && s.ch != c {
					s.next()
				}
				continue
			case eof:
				// handled in next loop
				continue
			default:
				if isDigit(s.ch) {
					var c rune
					for i := 0; ; {
						c = 10*c + (s.ch - '0')
						s.next()
						i++
						if i >= 3 || !isDigit(s.ch) {
							break
						}
						if c > 255 { // UCHAR_MAX
							s.error(off, "escape sequence too large")
						}
					}
					continue
				}
			}
		}
		s.next()
	}
	s.next()
}

// scanLongString scans for a long, bracket-enclosed string.
func (s *Scanner) scanLongString(off int, t token.Type) {
	eq := 0
	for s.ch == '=' {
		eq++
		s.next()
	}
	if s.ch != '[' {
		// TODO: syntax error
		s.error(off, "invalid long string delimiter near '<off:s.offset>'")
		return
	}
	s.next()
loop:
	for {
		if s.ch == eof {
			// TODO: EOF error
			if t == token.LONGCOMMENT {
				s.error(off, "unfinished long comment near '<eof>'")
			} else {
				s.error(off, "unfinished long string near '<eof>'")
			}
		}
		if s.ch == ']' {
			s.next()
			for i := 0; i < eq; i++ {
				if s.ch != '=' {
					continue loop
				}
				s.next()
			}
			if s.ch == ']' {
				s.next()
				break
			}
		}
		s.next()
	}
}

// scanComment scans for a short or long comment.
func (s *Scanner) scanComment(off int) token.Type {
	s.next()
	if s.ch == '[' {
		s.next()
		s.scanLongString(off, token.LONGCOMMENT)
		return token.LONGCOMMENT
	}
	for s.ch != '\n' && s.ch != eof {
		s.next()
	}
	return token.COMMENT
}

// Scan scans the next token and returns its offset, type, and the bytes
// represented by the token. The end of the source is indicated by token.EOF
// as the type.
//
// Scan adds line information to the token.File specified by Init.
func (s *Scanner) Scan() (off int, tok token.Type, lit []byte) {
	off = s.offset
	switch ch := s.ch; {
	case isSpace(ch):
		s.scanSpace()
		tok = token.SPACE
	case isLetter(ch):
		tok = token.Lookup(string(s.scanName()))
	case isDigit(ch):
		tok = s.scanNumber()
	case ch == '"', ch == '\'':
		s.scanString(off)
		tok = token.STRING
	default:
		s.next()
		switch ch {
		case '-':
			if s.ch == '-' {
				tok = s.scanComment(off)
			} else {
				tok = token.SUB
			}
		case '+':
			tok = token.ADD
		case '*':
			tok = token.MUL
		case '/':
			tok = token.DIV
		case '%':
			tok = token.MOD
		case '^':
			tok = token.EXP
		case '.':
			if isDigit(s.ch) {
				tok = s.scanNumber()
			} else if s.ch == '.' {
				s.next()
				if s.ch == '.' {
					s.next()
					tok = token.VARARG
				} else {
					tok = token.CONCAT
				}
			} else {
				tok = token.DOT
			}
		case '<':
			if s.ch == '=' {
				s.next()
				tok = token.LEQ
			} else {
				tok = token.LT
			}
		case '>':
			if s.ch == '=' {
				s.next()
				tok = token.GEQ
			} else {
				tok = token.GT
			}
		case '=':
			if s.ch == '=' {
				s.next()
				tok = token.EQ
			} else {
				tok = token.ASSIGN
			}
		case '~':
			if s.ch == '=' {
				s.next()
				tok = token.NEQ
			} else {
				s.error(s.offset, "unexpected symbol")
			}
		case ';':
			tok = token.SEMICOLON
		case ',':
			tok = token.COMMA
		case ':':
			tok = token.COLON
		case '[':
			if s.ch == '[' || s.ch == '=' {
				s.scanLongString(off, token.LONGSTRING)
				tok = token.LONGSTRING
			} else {
				tok = token.LBRACK
			}
		case ']':
			tok = token.RBRACK
		case '(':
			tok = token.LPAREN
		case ')':
			tok = token.RPAREN
		case '{':
			tok = token.LBRACE
		case '}':
			tok = token.RBRACE
		case '#':
			tok = token.LENGTH
		case eof:
			tok = token.EOF
		default:
			tok = token.INVALID
		}
	}
	lit = s.src[off:s.offset]
	return
}
