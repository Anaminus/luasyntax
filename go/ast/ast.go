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
	// Implements the WriteTo interface by writing the source-code equivalent
	// of the node. Assumes that the node is valid.
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
	// Info contains information about the file, such as the name, and line
	// offsets.
	Info *token.File
	// Body is the top-level block of the file.
	Body Block
	// EOFToken is the EOF token at the end of the file.
	EOFToken Token
}

// Block represents a Lua block.
type Block struct {
	// Items is a list of 0 or more Lua statements.
	Items []Stmt
	// Seps indicates whether a statement is followed by a semi-colon. The
	// length of Seps is the same as Stmts. Tokens will either be a SEMICOLON,
	// or INVALID when no semicolon is present.
	Seps []Token
}

// Len returns the combined length of Stmts and Seps.
func (b *Block) Len() int {
	return len(b.Items) + len(b.Seps)
}

// Expr is the interface that all Lua expressions implement.
type Expr interface {
	Node
	exprNode()
}

// ExprList represents a list of one or more Lua expressions.
type ExprList struct {
	// Items contains each expression in the list.
	Items []Expr
	// Seps contains each COMMA between expressions. The length of Seps is one
	// less than the length of Exprs.
	Seps []Token
}

// Len returns the combined length of Exprs and Seps.
func (l *ExprList) Len() int {
	return len(l.Items) + len(l.Seps)
}

// Name represents a Lua name expression.
type Name struct {
	// Token is the underlying NAME token.
	Token
	// Value is the evaluated form of the token. It is set only if the parser
	// has the EvalConst mode set, and is not used when printing.
	Value string
}

func (Name) exprNode() {}

// NameList represents a list of one or more name expressions.
type NameList struct {
	// Items contains each name in the list.
	Items []Name
	// Seps contains each COMMA between names. The length of Seps is one less
	// then the length of Names.
	Seps []Token
}

// Len returns the combined length of Names and Seps.
func (l *NameList) Len() int {
	return len(l.Items) + len(l.Seps)
}

// NumberExpr represents a Lua number expression.
type NumberExpr struct {
	// NumberToken is the number token holding the content of the expression.
	NumberToken Token
}

func (NumberExpr) exprNode() {}

// StringExpr represents a Lua string expression.
type StringExpr struct {
	// StringToken is the string token holding the content the expression.
	StringToken Token
}

func (StringExpr) exprNode() {}

// NilExpr represents a Lua nil expression.
type NilExpr struct {
	// NilToken is the NIL token holding the content the expression.
	NilToken Token
}

func (NilExpr) exprNode() {}

// BoolExpr represents a Lua boolean expression.
type BoolExpr struct {
	// BoolToken is the bool token holding the content the expression.
	BoolToken Token
}

func (BoolExpr) exprNode() {}

// VarArgExpr represents a Lua variable argument expression.
type VarArgExpr struct {
	// VarArgToken is the VARARG token holding the content the expression.
	VarArgToken Token
}

func (VarArgExpr) exprNode() {}

// UnopExpr represents a unary operation.
type UnopExpr struct {
	// UnopToken is the unary operator token.
	UnopToken Token
	// Operand is the expression being operated on.
	Operand Expr
}

func (UnopExpr) exprNode() {}

// BinopExpr represents a binary operation.
type BinopExpr struct {
	// Left is the left side of the operation.
	Left Expr
	// BinopToken is the binary operator token.
	BinopToken Token
	// Right is the right side of the binary operation.
	Right Expr
}

func (BinopExpr) exprNode() {}

// ParenExpr represents an expression enclosed in parentheses.
type ParenExpr struct {
	// LParenToken is the LPAREN token that opens the expression.
	LParenToken Token
	// Value is the enclosed expression.
	Value Expr
	// RParenToken is the RPAREN token that closes the expression.
	RParenToken Token
}

func (ParenExpr) exprNode() {}

// VariableExpr represents a variable name used as an expression.
type VariableExpr struct {
	// Name is the token indicating the name of the variable.
	Name Name
}

func (VariableExpr) exprNode() {}

// TableCtor represents a table constructor expression.
type TableCtor struct {
	// LBraceToken is the LBRACE token that opens the table.
	LBraceToken Token
	// Entries a list of entries in the table.
	Entries EntryList
	// RBraceToken is the RBRACE token that closes the table.
	RBraceToken Token
}

func (TableCtor) exprNode() {}

// EntryList represents a list of entries in a table.
type EntryList struct {
	// Items contains each entry in the list.
	Items []Entry
	// Seps contains each separator between entries, which will be either a
	// COMMA or SEMICOLON. The length of Seps is equal to or one less than the
	// length of Entries.
	Seps []Token
}

// Len returns the combined length of Entries and Seps.
func (l *EntryList) Len() int {
	return len(l.Items) + len(l.Seps)
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
	// Key is the expression evaluating to the key of the entry.
	Key Expr
	// RBrackToken is the RBRACK token that ends the entry key.
	RBrackToken Token
	// AssignToken is the ASSIGN token that begins the entry value.
	AssignToken Token
	// Value is the expression evaluating to the value of the entry.
	Value Expr
}

func (IndexEntry) entryNode() {}

// FieldEntry represents a table entry defining a value with a field as the
// key.
type FieldEntry struct {
	// Name is the name of the field, evaluating to the key of the entry.
	Name Name
	// AssignToken is the ASSIGN token that begins the entry value.
	AssignToken Token
	// Value is the expression evaluating to the value of the entry.
	Value Expr
}

func (FieldEntry) entryNode() {}

// ValueEntry represents a table entry defining a value with a numerical key.
type ValueEntry struct {
	// Value is the expression evaluating to the value of the entry.
	Value Expr
}

func (ValueEntry) entryNode() {}

// FunctionExpr represents a Lua function, primarily an anonymous function
// expression. It is also used by other nodes for other representations of
// functions.
type FunctionExpr struct {
	// FuncToken is the FUNCTION token that begins the function.
	FuncToken Token
	// LParenToken is the LPAREN token that opens the function's parameters.
	LParenToken Token
	// Params is a list of named parameters of the function. It will be nil if
	// the function has no named parameters.
	Params *NameList
	// VarArgSepToken is the token preceding a variable-argument token. This
	// will be a COMMA when a vararg parameter follows a named parameter, and
	// INVALID otherwise.
	VarArgSepToken Token
	// VarArgToken is the VARARG token following the named parameters of the
	// function. It will be INVALID if the vararg parameter is not present.
	VarArgToken Token
	// RParenToken is the RPAREN token that closes the function's parameters.
	RParenToken Token
	// Body is the body of the function.
	Body Block
	// EndToken is the END token that ends the function.
	EndToken Token
}

func (FunctionExpr) exprNode() {}

// FieldExpr represents an expression that indexes a value from another
// expression, by field.
type FieldExpr struct {
	// Value is the expression being operated on.
	Value Expr
	// DotToken is the DOT token that separates the field.
	DotToken Token
	// Name is the name of the field.
	Name Name
}

func (FieldExpr) exprNode() {}

// IndexExpr represents an expression that indexes a value from another
// expression, by key.
type IndexExpr struct {
	// Value is the expression being operated on.
	Value Expr
	// LBrackToken is the LBRACK token opening the key expression.
	LBrackToken Token
	// Index is the expression evaluating to the key.
	Index Expr
	// RBrackToken is the RBRACK token closing the key expression.
	RBrackToken Token
}

func (IndexExpr) exprNode() {}

// Call is the interface that all function call expressions implement.
type Call interface {
	Node
	callNode()
}

// MethodExpr represents an expression that gets and calls a method on another
// expression.
type MethodExpr struct {
	// Value is the expression being operated on.
	Value Expr
	// ColonToken is the COLON token that separates the method name.
	ColonToken Token
	// Name is the name of the method.
	Name Name
	// Args holds the arguments of the method call.
	Args Args
}

func (MethodExpr) exprNode() {}
func (MethodExpr) callNode() {}

// CallExpr represents an expression that calls another expression as a
// function.
type CallExpr struct {
	// Value is the expression being operated on.
	Value Expr
	// Args holds the arguments of the function call.
	Args Args
}

func (CallExpr) exprNode() {}
func (CallExpr) callNode() {}

// Args is the interface that all function call argument nodes implement.
type Args interface {
	Node
	argsNode()
}

// ListArgs represents the arguments of a function call, in the form of a list
// of expressions.
type ListArgs struct {
	// LParenToken is the LPAREN token that opens the argument list.
	LParenToken Token
	// Values contains each argument of the call. It is nil if the call has no
	// arguments.
	Values *ExprList
	// RParenToken is the RPAREN token that closes the argument list.
	RParenToken Token
}

func (ListArgs) argsNode() {}

// TableArg represents the arguments of a function call, in the form of a
// single table constructor.
type TableArg struct {
	// Value is the table constructor expression.
	Value TableCtor
}

func (TableArg) argsNode() {}

// StringArg represents the arguments of a function call, in the form of a
// single string expression.
type StringArg struct {
	// Value is the string expression.
	Value StringExpr
}

func (StringArg) argsNode() {}

// Stmt is the interface that all statement nodes implement.
type Stmt interface {
	Node
	stmtNode()
}

// DoStmt represents a `do ... end` Lua statement.
type DoStmt struct {
	// DoToken is the DO token that begins the do statement.
	DoToken Token
	// Body is the body of the do statement.
	Body Block
	// EndToken is the END token that ends the do statement.
	EndToken Token
}

func (DoStmt) stmtNode() {}

// AssignStmt represents the assignment of one or more variables with a number
// of values.
type AssignStmt struct {
	// Left is the left side of the assignment, comprised of one or more
	// expressions.
	Left ExprList
	// AssignToken is the ASSIGN token separating the values.
	AssignToken Token
	// Right is the right side of the assignment, comprised of one or more
	// expressions.
	Right ExprList
}

func (AssignStmt) stmtNode() {}

// CallStmt represents a call expression as a statement.
type CallStmt struct {
	// Call is the call expression.
	Call Call
}

func (CallStmt) stmtNode() {}

// IfStmt represents a `if .. then .. end` statement.
type IfStmt struct {
	// IfToken is the IF token that begins the if statement.
	IfToken Token
	// Cond is the condition of the if statement.
	Cond Expr
	// ThenToken is the THEN token that begins the body of the if statement.
	ThenToken Token
	// Body is the body of the if statement.
	Body Block
	// ElseIf is a list of zero or more elseif clauses of the if statement.
	ElseIf []ElseIfClause
	// Else is the else clause of the if statement. It is nil if not present.
	Else *ElseClause
	// EndToken is the END token that ends the if statement.
	EndToken Token
}

func (IfStmt) stmtNode() {}

// ElseIfClause represents an `elseif .. then` clause within an `if`
// statement.
type ElseIfClause struct {
	// ElseIfToken is th ELSEIF token that begins the elseif clause.
	ElseIfToken Token
	// Cond is the condition of the elseif clause.
	Cond Expr
	// ThenToken is the THEN token that begins the body of the elseif clause.
	ThenToken Token
	// Body is the body of the elseif clause.
	Body Block
}

// ElseClause represents an `else` clause within an `if` statement.
type ElseClause struct {
	// ElseToken is the ELSE token that begins the body of the else clause.
	ElseToken Token
	// Body is the body of the else clause.
	Body Block
}

// NumericForStmt represents a numeric `for` statement.
type NumericForStmt struct {
	// ForToken is the FOR token that begins the for statement.
	ForToken Token
	// Name is the name of the control variable.
	Name Name
	// AssignToken is the ASSIGN token that begins the control expressions.
	AssignToken Token
	// Min is the expression indicating the lower bound of the control
	// variable.
	Min Expr
	// MaxSepToken is the COMMA token that separates the lower and upper
	// bound.
	MaxSepToken Token
	// Max is the expression indicating the upper bound of the control
	// variable.
	Max Expr
	// StepSepToken is the separator token between the upper bound and the
	// step expressions. It is a COMMA if the step is present, and INVALID
	// otherwise.
	StepSepToken Token
	// Step is the expression indicating the step of the control variable. It
	// is nil if not present.
	Step Expr
	// DoToken is the DO token that begins the body of the for statement.
	DoToken Token
	// Body is the body of the for statement.
	Body Block
	// EndToken is the END token that ends the for statement.
	EndToken Token
}

func (NumericForStmt) stmtNode() {}

// GenericForStmt represents a generic `for` statement.
type GenericForStmt struct {
	// ForToken is the FOR token that begins the for statement.
	ForToken Token
	// Names is the list of names of variables that will be assigned to by the
	// iterator.
	Names NameList
	// InToken is the IN token that separates the variables from the iterator
	// expressions.
	InToken Token
	// Iterator is the list of expressions that evaluate to the iterator of
	// the for statement.
	Iterator ExprList
	// DoToken is the DO token that begins the body of the for statement.
	DoToken Token
	// Body is the body of the for statement.
	Body Block
	// EndToken is the END token that ends the for statement.
	EndToken Token
}

func (GenericForStmt) stmtNode() {}

// WhileStmt represents a `while .. do .. end` statement.
type WhileStmt struct {
	// WhileToken is the WHILE token that begins a while statement.
	WhileToken Token
	// Cond is the condition of the while statement.
	Cond Expr
	// DoToken is the DO token that begins the body of the while statement.
	DoToken Token
	// Body is body of the while statement.
	Body Block
	// EndToken is the END token that ends the while statement.
	EndToken Token
}

func (WhileStmt) stmtNode() {}

// RepeatStmt represents a `repeat .. until ..` statement.
type RepeatStmt struct {
	// RepeatToken is the REPEAT token that begins the repeat statement.
	RepeatToken Token
	// Body is the body of the repeat statement.
	Body Block
	// UntilToken is the UNTIL token that ends the body of the repeat
	// statement.
	UntilToken Token
	// Cond is the condition of the repeat statement.
	Cond Expr
}

func (RepeatStmt) stmtNode() {}

// LocalVarStmt represents the statement that assigns local variables.
type LocalVarStmt struct {
	// LocalToken is the LOCAL token that begins the local statement.
	LocalToken Token
	// Names contains the name of each variable in the local statement.
	Names NameList
	// AssignToken is the ASSIGN token that separates the variables from the
	// values. It is INVALID if not present.
	AssignToken Token
	// Values is the list of expressions that are assigned to each variable.
	// It is nil if not present.
	Values *ExprList
}

func (LocalVarStmt) stmtNode() {}

// LocalFunctionStmt represents the statement that assigns a function to a
// local variable.
type LocalFunctionStmt struct {
	// LocalToken is the LOCAL token that begins the local statement.
	LocalToken Token
	// Name is the name of the function. Note that this token is located after
	// the FuncToken of the FunctionExpr.
	Name Name
	// Func defines the parameters and body of the function.
	Func FunctionExpr
}

func (LocalFunctionStmt) stmtNode() {}

// FunctionStmt represents the statement that assigns a function.
type FunctionStmt struct {
	// Name contains the name of the function. Note that tokens within this
	// are located after the FuncToken of the FunctionExpr.
	Name FuncNameList
	// Func defines the parameters and body of the function.
	Func FunctionExpr
}

func (FunctionStmt) stmtNode() {}

// FuncNameList represents a list of dot-separated names in a function
// statement. The list may optionally be followed by a method name, indicating
// that the function is a method.
type FuncNameList struct {
	// Items contains the chain of one or more names, indicating the name of a
	// function statement. Each successive name is a field of the previous
	// value.
	Items []Name
	// Seps contains each DOT between names. The length of Seps is one less
	// than the length of Exprs.
	Seps []Token
	// ColonToken is the COLON token following the last Name in Items, and
	// preceding Method. It is INVALID if the method is not present.
	ColonToken Token
	// Method indicates that the function name describes a method, as well as
	// the name of the method. It is INVALID if the method is not present.
	Method Name
}

// Len returns the combined length of Names and Seps, and ColonToken and
// Method, if present.
func (l *FuncNameList) Len() int {
	n := len(l.Items) + len(l.Seps)
	if l.ColonToken.Type.IsValid() {
		n += 2
	}
	return n
}

// BreakStmt represents a `break` statement.
type BreakStmt struct {
	// BreakToken is the BREAK token of the break statement.
	BreakToken Token
}

func (BreakStmt) stmtNode() {}

// ReturnStmt represents a `return` statement.
type ReturnStmt struct {
	// ReturnToken is the RETURN token of the return statement.
	ReturnToken Token
	// Values is the list of expressions that evaluate to the values being
	// returned. Will be nil if there are no values.
	Values *ExprList
}

func (ReturnStmt) stmtNode() {}
