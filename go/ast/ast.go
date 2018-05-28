// The ast package declares the types used to represent Lua syntax trees.
package ast

import (
	"github.com/anaminus/luasyntax/go/token"
	"io"
)

// Notes
//
// A pointer is used when
//     - the value is within an interface.
//     - the field is optional.
//
// Methods assume that *<node> is not nil

// Node is the interface that all AST nodes implement.
type Node interface {
	// IsValid returns whether the node is well-formed (child nodes excluded).
	IsValid() bool
	// FirstToken returns the first Token in the node. Assumes that the node
	// is valid.
	FirstToken() *Token
	// LastToken returns the last Token in the node. Assumes that the node is
	// valid.
	LastToken() *Token
	// Implements the WriteTo interface. Assumes that the node is valid.
	io.WriterTo
}

// A Token represents a token within a file.
type Token struct {
	// Type indicates the token's type.
	Type token.Type
	// Prefix holds a list of Prefix tokens that precede this token, ordered
	// left to right.
	Prefix []Prefix
	// Offset is the location of the token within a file.
	Offset int
	// Bytes contains the actual bytes of a file that the token represents.
	Bytes []byte
}

// Prefix represents a token that precedes a Token. A token.Type is a prefix
// when Type.IsPrefix returns true.
type Prefix struct {
	Type  token.Type
	Bytes []byte
}

// StartOffset returns the offset at the start of the token, which includes
// the prefix.
func (t *Token) StartOffset() int {
	if t == nil {
		return 0
	}
	if !t.Type.IsValid() {
		return 0
	}
	n := t.Offset
	for _, p := range t.Prefix {
		n -= len(p.Bytes)
	}
	return n
}

// EndOffset returns the offset following the end of the token.
func (t *Token) EndOffset() int {
	if t == nil {
		return 0
	}
	if !t.Type.IsValid() {
		return 0
	}
	return t.Offset + len(t.Bytes)
}

// File is a Node representing an entire file.
type File struct {
	// Name is the name of the file. It is not used when printing.
	Name string
	// Block is the top-level block of the file.
	Block Block
	// EOFToken is the EOF token at the end of the file.
	EOFToken Token
}

// Block represents a Lua block.
type Block struct {
	// Stats is a list of 0 or more Lua statements.
	Stats []Stat
	// Seps indicates whether a statement is followed by a semi-colon. The
	// length of Seps is the same as Stats. Tokens will either be a SEMICOLON,
	// or INVALID when no semicolon is present.
	Seps []Token
}

// Len returns the combined length of Stats and Seps.
func (b *Block) Len() int {
	return len(b.Stats) + len(b.Seps)
}

// Exp is the interface that all Lua expressions implement.
type Exp interface {
	Node
	expNode()
}

// ExpList represents a list of one or more Lua expressions.
type ExpList struct {
	// Exps contains each expression in the list.
	Exps []Exp
	// Seps contains each COMMA between expressions. The length of Seps is one
	// less than the length of Exps.
	Seps []Token
}

// Len returns the combined length of Exps and Seps.
func (l *ExpList) Len() int {
	return len(l.Exps) + len(l.Seps)
}

// Name represents a Lua name expression.
type Name struct {
	// Token is the underlying NAME token.
	Token
	// Value is the parsed form of the token. It is not used when printing.
	Value string
}

func (Name) expNode() {}

// NameList represents a list of one or more name expressions.
type NameList struct {
	// Names contains each name in the list.
	Names []Name
	// Seps contains each COMMA between names. The length of Seps is one less
	// then the length of Names.
	Seps []Token
}

// Len returns the combined length of Names and Seps.
func (l *NameList) Len() int {
	return len(l.Names) + len(l.Seps)
}

// Number represents a Lua number expression.
type Number struct {
	// Token is the underlying number token.
	Token
	// Value is the parsed form of the token. It is not used when printing.
	Value float64
}

func (Number) expNode() {}

// String represents a Lua string expression.
type String struct {
	// Token is the underlying string token.
	Token
	// Value is the parsed form of the token. It is not used when printing.
	Value string
}

func (String) expNode() {}

// Nil represents a Lua nil expression.
type Nil struct {
	// Token is the underlying NIL token.
	Token
}

func (Nil) expNode() {}

// Bool represents a Lua boolean expression.
type Bool struct {
	// Token is the underlying boolean token.
	Token
	// Value is the parsed form of the token. It is not used when printing.
	Value bool
}

func (Bool) expNode() {}

// VarArg represents a Lua variable argument expression.
type VarArg struct {
	// Token is the underlying VARARG token.
	Token
}

func (VarArg) expNode() {}

// UnopExp represents a unary operation.
type UnopExp struct {
	// UnopToken is the unary operator token.
	UnopToken Token
	// Exp is the expression being operated on.
	Exp Exp
}

func (UnopExp) expNode() {}

// BinopExp represents a binary operation.
type BinopExp struct {
	// Left is the left side of the operation.
	Left Exp
	// BinopToken is the binary operator token.
	BinopToken Token
	// Right is the right side of the binary operation.
	Right Exp
}

func (BinopExp) expNode() {}

// ParenExp represents an expression enclosed in parentheses.
type ParenExp struct {
	// LParenToken is the LPAREN token that opens the expression.
	LParenToken Token
	// Exp is the enclosed expression.
	Exp Exp
	// RParenToken is the RPAREN token that closes the expression.
	RParenToken Token
}

func (ParenExp) expNode() {}

// VariableExp represents a variable name used as an expression.
type VariableExp struct {
	// NameToken is the token indicating the name of the variable.
	NameToken Name
}

func (VariableExp) expNode() {}

// TableCtor represents a table constructor expression.
type TableCtor struct {
	// LBraceToken is the LBRACE token that opens the table.
	LBraceToken Token
	// Fields a list of entries in the table.
	Fields FieldList
	// RBraceToken is the RBRACE token that closes the table.
	RBraceToken Token
}

func (TableCtor) expNode() {}

// FieldList represents a list of entries in a table.
type FieldList struct {
	// Entries contains each entry in the list.
	Entries []Entry
	// Seps contains each separator between entries, which will be either a
	// COMMA or SEMICOLON. The length of Seps is equal to or one less than the
	// length of Entries.
	Seps []Token
}

// Len returns the combined length of Entries and Seps.
func (l *FieldList) Len() int {
	return len(l.Entries) + len(l.Seps)
}

// Entry is the interface that all table entry nodes implement.
type Entry interface {
	Node
	entryNode()
}

// IndexEntry represents a table entry defining a key-value pair.
type IndexEntry struct {
	// LBrackToken is the LBRACK token that begins the entry key.
	LBrackToken Token
	// KeyExp is the expression evaluating to the key of the entry.
	KeyExp Exp
	// RBrackToken is the RBRACK token that ends the entry key.
	RBrackToken Token
	// AssignToken is the ASSIGN token that begins the entry value.
	AssignToken Token
	// ValueExp is the expression evaluating to the value of the entry.
	ValueExp Exp
}

func (IndexEntry) entryNode() {}

// FieldEntry represents a table entry defining a value with a field as the key.
type FieldEntry struct {
	// Name is the name of the field, evaluating to the key of the entry.
	Name Name
	// AssignToken is the ASSIGN token that begins the entry value.
	AssignToken Token
	// Value is the expression evaluating to the value of the entry.
	Value Exp
}

func (FieldEntry) entryNode() {}

// ValueEntry represents a table entry defining a value with a numerical key.
type ValueEntry struct {
	// Value is the expression evaluating to the value of the entry.
	Value Exp
}

func (ValueEntry) entryNode() {}

// FunctionExp represents a Lua function, primarily an anonymous function
// expression. It is also used by other nodes for other representations of
// functions.
type FunctionExp struct {
	// FuncToken is the FUNCTION token that begins the function.
	FuncToken Token
	// LParenToken is the LPAREN token that opens the function's parameters.
	LParenToken Token
	// ParList is a list of named parameters of the function. It will be nil
	// if the function has no named parameters.
	ParList *NameList
	// VarArgSepToken is the token preceding a variable-argument token. This
	// will be a COMMA when a vararg parameter follows a named parameter, and
	// INVALID otherwise.
	VarArgSepToken Token
	// VarArgToken is the VARARG token following the named parameters of the
	// function. It will be INVALID if the vararg parameter is not present.
	VarArgToken Token
	// RParenToken is the RPAREN token that closes the function's parameters.
	RParenToken Token
	// Block is the body of the function.
	Block Block
	// EndToken is the END token that ends the function.
	EndToken Token
}

func (FunctionExp) expNode() {}

// FieldExp represents an expression that indexes a value from another
// expression, by field.
type FieldExp struct {
	// Exp is the expression being operated on.
	Exp Exp
	// DotToken is the DOT token that separates the field.
	DotToken Token
	// Field is the name of the field.
	Field Name
}

func (FieldExp) expNode() {}

// IndexExp represents an expression that indexes a value from another
// expression, by key.
type IndexExp struct {
	// Exp is the expression being operated on.
	Exp Exp
	// LBrackToken is the LBRACK token opening the key expression.
	LBrackToken Token
	// Index is the expression evaluating to the key.
	Index Exp
	// RBrackToken is the RBRACK token closing the key expression.
	RBrackToken Token
}

func (IndexExp) expNode() {}

// MethodExp represents an expression that gets and calls a method on another
// expression.
type MethodExp struct {
	// Exp is the expression being operated on.
	Exp Exp
	// ColonToken is the COLON token that separates the method name.
	ColonToken Token
	// Name is the name of the method.
	Name Name
	// Args holds the arguments of the method call.
	Args CallArgs
}

func (MethodExp) expNode() {}

// CallExp represents an expression that calls another expression as a
// function.
type CallExp struct {
	// Exp is the expression being operated on.
	Exp Exp
	// Args holds the arguments of the function call.
	Args CallArgs
}

func (CallExp) expNode() {}

// CallArgs is the interface that all function call argument nodes implement.
type CallArgs interface {
	Node
	callArgsNode()
}

// ArgsCall represents the arguments of a function call, in the form of a list
// of expressions.
type ArgsCall struct {
	// LParenToken is the LPAREN token that opens the argument list.
	LParenToken Token
	// ExpList contains each argument of the call. It is nil if the call has
	// no arguments.
	ExpList *ExpList
	// RParenToken is the RPAREN token that closes the argument list.
	RParenToken Token
}

func (ArgsCall) callArgsNode() {}

// TableCall represents the arguments of a function call, in the form of a
// single table constructor.
type TableCall struct {
	// TableExp is the table constructor expression.
	TableExp TableCtor
}

func (TableCall) callArgsNode() {}

// StringCall represents the arguments of a function call, in the form of a
// single string expression.
type StringCall struct {
	// StringExp is the string expression.
	StringExp String
}

func (StringCall) callArgsNode() {}

// Stat is the interface that all statement nodes implement.
type Stat interface {
	Node
	statNode()
}

// DoStat represents a `do ... end` Lua statement.
type DoStat struct {
	// DoToken is the DO token that begins the do statement.
	DoToken Token
	// Block is the body of the do statement.
	Block Block
	// EndToken is the END token that ends the do statement.
	EndToken Token
}

func (DoStat) statNode() {}

// AssignStat represents the assignment of one or more variables with a number
// of values.
type AssignStat struct {
	// Left is the left side of the assignment, comprised of one or more
	// expressions.
	Left ExpList
	// AssignToken is the ASSIGN token separating the values.
	AssignToken Token
	// Right is the right side of the assignment, comprised of one or more
	// expressions.
	Right ExpList
}

func (AssignStat) statNode() {}

// CallExprStat represents a call expression as a statement.
type CallExprStat struct {
	// Exp is the call expression.
	Exp Exp
}

func (CallExprStat) statNode() {}

// IfStat represents a `if .. then .. end` statement.
type IfStat struct {
	// IfToken is the IF token that begins the if statement.
	IfToken Token
	// Exp is the condition of the if statement.
	Exp Exp
	// ThenToken is the THEN token that begins the body of the if statement.
	ThenToken Token
	// Block is the body of the if statement.
	Block Block
	// ElseIfClauses is a list of zero or more elseif clauses of the if
	// statement.
	ElseIfClauses []ElseIfClause
	// ElseClause is the else clause of the if statement. It is nil if not
	// present.
	ElseClause *ElseClause // nil if not present
	// EndToken is the END token that ends the if statement.
	EndToken Token
}

func (IfStat) statNode() {}

// ElseIfClause represents an `elseif .. then` clause within an `if` statement.
type ElseIfClause struct {
	// ElseIfToken is th ELSEIF token that begins the elseif clause.
	ElseIfToken Token
	// Exp is the condition of the elseif clause.
	Exp Exp
	// ThenToken is the THEN token that begins the body of the elseif clause.
	ThenToken Token
	// Block is the body of the elseif clause.
	Block Block
}

// ElseClause represents an `else` clause within an `if` statement.
type ElseClause struct {
	// ElseToken is the ELSE token that begins the body of the else clause.
	ElseToken Token
	// Block is the body of the else clause.
	Block Block
}

// NumericForStat represents a numeric `for` statement.
type NumericForStat struct {
	// ForToken is the FOR token that begins the for statement.
	ForToken Token
	// Name is the name of the control variable.
	Name Name
	// AssignToken is the ASSIGN token that begins the control expressions.
	AssignToken Token
	// MinExp is the expression indicating the lower bound of the control variable.
	MinExp Exp
	// MaxSepToken is the COMMA token that separates the lower and upper
	// bound.
	MaxSepToken Token
	// MaxExp is the expression indicating the upper bound of the control variable.
	MaxExp Exp
	// StepSepToken is the separator token between the upper bound and the
	// step expressions. It is a COMMA if the step is present, and INVALID
	// otherwise.
	StepSepToken Token
	// StepExp is the expression indicating the step of the control variable.
	// It is nil if not present.
	StepExp Exp
	// DoToken is the DO token that begins the body of the for statement.
	DoToken Token
	// Block is the body of the for statement.
	Block Block
	// EndToken is the END token that ends the for statement.
	EndToken Token
}

func (NumericForStat) statNode() {}

// GenericForStat represents a generic `for` statement.
type GenericForStat struct {
	// ForToken is the FOR token that begins the for statement.
	ForToken Token
	// NameList is the list of names of variables that will be assigned to by the iterator.
	NameList NameList
	// InToken is the IN token that separates the variables from the iterator
	// expressions.
	InToken Token
	// ExpList is the list of expressions that evaluate to the iterator of the
	// for statement.
	ExpList ExpList
	// DoToken is the DO token that begins the body of the for statement.
	DoToken Token
	// Block is the body of the for statement.
	Block Block
	// EndToken is the END token that ends the for statement.
	EndToken Token
}

func (GenericForStat) statNode() {}

// WhileStat represents a `while .. do .. end` statement.
type WhileStat struct {
	// WhileToken is the WHILE token that begins a while statement.
	WhileToken Token
	// Exp is the condition of the while statement.
	Exp Exp
	// DoToken is the DO token that begins the body of the while statement.
	DoToken Token
	// Block is body of the while statement.
	Block Block
	// EndToken is the END token that ends the while statement.
	EndToken Token
}

func (WhileStat) statNode() {}

// RepeatStat represents a `repeat .. until ..` statement.
type RepeatStat struct {
	// RepeatToken is the REPEAT token that begins the repeat statement.
	RepeatToken Token
	// Block is the body of the repeat statement.
	Block Block
	// UntilToken is the UNTIL token that ends the body of the repeat statement.
	UntilToken Token
	// Exp is the condition of the repeat statement.
	Exp Exp
}

func (RepeatStat) statNode() {}

// LocalVarStat represents the statement that assigns local variables.
type LocalVarStat struct {
	// LocalToken is the LOCAL token that begins the local statement.
	LocalToken Token
	// NameList contains the names of each variable in the local statement.
	NameList NameList
	// AssignToken is the ASSIGN token that separates the variables from the
	// values. It is INVALID if not present.
	AssignToken Token
	// ExpList is the list of expressions that are assigned to each variable.
	// It is nil if not present.
	ExpList *ExpList
}

func (LocalVarStat) statNode() {}

// LocalFunctionStat represents the statement that assigns a function to a
// local variable.
type LocalFunctionStat struct {
	// LocalToken is the LOCAL token that begins the local statement.
	LocalToken Token
	// Name is the name of the function. Note that this token is located after
	// the FuncToken of the FunctionExp.
	Name Name
	// Exp defines the parameters and body of the function.
	Exp FunctionExp
}

func (LocalFunctionStat) statNode() {}

// FunctionStat represents the statement that assigns a function.
type FunctionStat struct {
	// Name contains the name of the function. Note that tokens within this
	// are located after the FuncToken of the FunctionExp.
	Name FuncNameList
	// Exp defines the parameters and body of the function.
	Exp FunctionExp
}

func (FunctionStat) statNode() {}

// FuncNameList represents a list of dot-separated names in a function
// statement.
type FuncNameList struct {
	// Names contains the chain of one or more names, indicating the name of a
	// function statement. Each successive name is a field of the previous
	// value. The last name may or may not indicate a method name, as
	// determined by the Seps field.
	Names []Name
	// Seps contains each DOT between names. If the last token is a COLON, it
	// indicates that the last name is a method. The length of Seps is one
	// less than the length of Exps.
	Seps []Token
	// TODO: colon as separate field
}

// Len returns the combined length of Names and Seps.
func (l *FuncNameList) Len() int {
	return len(l.Names) + len(l.Seps)
}

// BreakStat represents a `break` statement.
type BreakStat struct {
	// BreakToken is the BREAK token of the break statement.
	BreakToken Token
}

func (BreakStat) statNode() {}

// ReturnStat represents a `return` statement.
type ReturnStat struct {
	// ReturnToken is the RETURN token of the return statement.
	ReturnToken Token
	// ExpList is the list of expressions that evaluate to the values being
	// returned. Will be nil if there are no values.
	ExpList *ExpList
}

func (ReturnStat) statNode() {}
