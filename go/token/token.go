// The token package defines constants representing the lexical tokens of the
// Lua programming language.
package token

import (
	"strconv"
)

// Type indicates the type of a token.
type Type int

// Note: The list contains markers used only to indicate various classes of
// tokens.

const (
	INVALID      Type = iota
	valid_start       // [ VALID
	EOF               // End of file
	pre_start         // [ PREFIXES
	SPACE             // All whitespace
	comm_start        // [ COMMENTS
	COMMENT           // Line-style comment
	LONGCOMMENT       // Block-style comment
	comm_end          // COMMENTS ]
	pre_end           // PREFIXES ]
	NAME              // Identifier
	num_start         // [ NUMBER
	NUMBERFLOAT       // Float
	NUMBERHEX         // Hexadecimal
	num_end           // NUMBER ]
	str_start         // [ STRINGS
	STRING            // Quote-style string
	LONGSTRING        // Block-style string
	str_end           // STRINGS ]
	op_start          // [ OPERATORS
	SEMICOLON         // `;`
	ASSIGN            // `=`
	COMMA             // `,`
	DOT               // `.`
	COLON             // `:`
	LBRACK            // `[`
	RBRACK            // `]`
	VARARG            // `...`
	LPAREN            // `(`
	RPAREN            // `)`
	LBRACE            // `{`
	RBRACE            // `}`
	binop_start       // [ BINARY OPERATORS
	ADD               // `+`
	MUL               // `*`
	DIV               // `/`
	MOD               // `%`
	EXP               // `^`
	CONCAT            // `..`
	LT                // `<`
	LEQ               // `<=`
	GT                // `>`
	GEQ               // `>=`
	EQ                // `==`
	NEQ               // `~=`
	unop_start        // [ UNARY
	SUB               // `-`
	binop_end         // BINARY OPERATORS ]
	LENGTH            // `#`
	op_end            /// OPERATORS ]
	key_start         // [ KEYWORDS
	NOT               // `not`
	unop_end          // UNARY ]
	DO                // `do`
	END               // `end`
	WHILE             // `while`
	REPEAT            // `repeat`
	UNTIL             // `until`
	IF                // `if`
	THEN              // `then`
	ELSEIF            // `elseif`
	ELSE              // `else`
	FOR               // `for`
	IN                // `in`
	LOCAL             // `local`
	FUNCTION          // `function`
	RETURN            // `return`
	BREAK             // `break`
	NIL               // `nil`
	bool_start        // [ BOOLEANS
	FALSE             // `false`
	TRUE              // `true`
	bool_end          // BOOLEANS ]
	binkey_start      // [ BINARY KEYWORDS
	AND               // `and`
	OR                // `or`
	binkey_end        // BINARY KEYWORDS ]
	key_end           // KEYWORDS ]
	valid_end         // VALID ]
)

var tokens = [...]string{
	INVALID:     "<invalid>",
	EOF:         "<eof>",
	SPACE:       "<space>",
	COMMENT:     "<comment>",
	LONGCOMMENT: "<comment>",
	NAME:        "<name>",
	NUMBERFLOAT: "<number>",
	NUMBERHEX:   "<number>",
	STRING:      "<string>",
	LONGSTRING:  "<string>",
	ADD:         "+",
	SUB:         "-",
	MUL:         "*",
	DIV:         "/",
	EXP:         "^",
	MOD:         "%",
	CONCAT:      "..",
	LT:          "<",
	LEQ:         "<=",
	GT:          ">",
	GEQ:         ">=",
	EQ:          "==",
	NEQ:         "~=",
	SEMICOLON:   ";",
	ASSIGN:      "=",
	COMMA:       ",",
	DOT:         ".",
	COLON:       ":",
	LBRACK:      "[",
	RBRACK:      "]",
	VARARG:      "...",
	LPAREN:      "(",
	RPAREN:      ")",
	LBRACE:      "{",
	RBRACE:      "}",
	LENGTH:      "#",
	DO:          "do",
	END:         "end",
	WHILE:       "while",
	REPEAT:      "repeat",
	UNTIL:       "until",
	IF:          "if",
	THEN:        "then",
	ELSEIF:      "elseif",
	ELSE:        "else",
	FOR:         "for",
	IN:          "in",
	LOCAL:       "local",
	FUNCTION:    "function",
	RETURN:      "return",
	BREAK:       "break",
	NIL:         "nil",
	FALSE:       "false",
	TRUE:        "true",
	AND:         "and",
	OR:          "or",
	NOT:         "not",
	// So that we don't get errors when iterating keywords.
	unop_end:     "",
	bool_start:   "",
	bool_end:     "",
	binkey_start: "",
	binkey_end:   "",
}

// String returns a string representation of the token type, when possible.
func (t Type) String() (s string) {
	if 0 <= t && t < Type(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return
}

// IsValid returns whether the type is valid.
func (t Type) IsValid() bool {
	return valid_start < t && t < valid_end
}

// IsPrefix returns whether the type is considered non-essential data. When
// parsing, prefixes are held within a token structure as extra data that
// precedes the main content.
//
// Spaces and comments are treated as prefixes.
func (t Type) IsPrefix() bool {
	return pre_start < t && t < pre_end
}

// IsComment returns whether the type indicates a comment.
func (t Type) IsComment() bool {
	return comm_start < t && t < comm_end
}

// IsNumber returns whether the type indicates a number value.
func (t Type) IsNumber() bool {
	return num_start < t && t < num_end
}

// IsString returns whether the type indicates a string value.
func (t Type) IsString() bool {
	return str_start < t && t < str_end
}

// IsOperator returns whether the type indicates an operator.
func (t Type) IsOperator() bool {
	return key_start < t && t < key_end
}

// IsUnary returns whether the type indicates a unary operator.
func (t Type) IsUnary() bool {
	return unop_start < t && t < unop_end
}

// IsKeyword returns whether the type indicates a keyword.
func (t Type) IsKeyword() bool {
	return key_start < t && t < key_end
}

// IsBool returns whether the type indicates a boolean value.
func (t Type) IsBool() bool {
	return bool_start < t && t < bool_end
}

// IsBinary returns whether the type indicates a binary operator.
func (t Type) IsBinary() bool {
	return (binop_start < t && t < binop_end) || (binkey_start < t && t < binkey_end)
}

// Precedence returns the priority of binary operators. The returned numbers
// indicate the left and right priorities, respectively.
func (t Type) Precedence() [2]int {
	switch t {
	case EXP:
		return [2]int{10, 9}
	case MUL, DIV, MOD:
		return [2]int{7, 7}
	case ADD, SUB:
		return [2]int{6, 6}
	case CONCAT:
		return [2]int{5, 4}
	case LT, GT, LEQ, GEQ, NEQ, EQ:
		return [2]int{3, 3}
	case AND:
		return [2]int{2, 2}
	case OR:
		return [2]int{1, 1}
	}
	return [2]int{0, 0}
}

// UnaryPrecedence indicates the priority of unary operators.
const UnaryPrecedence = 8

var keywords map[string]Type

func init() {
	keywords = make(map[string]Type)
	for i := key_start + 1; i < key_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup maps a name to its keyword, or NAME if it is not a keyword.
func Lookup(name string) Type {
	if t, ok := keywords[name]; ok {
		return t
	}
	return NAME
}
