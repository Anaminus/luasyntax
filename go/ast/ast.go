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

// A Token represents a token within a file. The location is indicated by the
// Offset field. Leading whitespace and comments are merged into the token,
// represented by the Prefix field.
type Token struct {
	Type   token.Type
	Prefix []Prefix
	Offset int
	Bytes  []byte
}

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

type File struct {
	Name     string
	Block    Block
	EOFToken Token
}

type Block struct {
	Stats []Stat
	Seps  []Token // Whether statement is followed by a semicolon.
}

func (b *Block) Len() int {
	if b == nil {
		return 0
	}
	return len(b.Stats) + len(b.Seps)
}

type Exp interface {
	Node
	expNode()
}

type ExpList struct {
	Exps []Exp
	Seps []Token // COMMA between expressions.
}

func (l *ExpList) Len() int {
	return len(l.Exps) + len(l.Seps)
}

type Name struct {
	Token
	Value string
}

func (Name) expNode() {}

type NameList struct {
	Names []Name
	Seps  []Token // COMMA between names.
}

func (l *NameList) Len() int {
	return len(l.Names) + len(l.Seps)
}

type Number struct {
	Token
	Value float64
}

func (Number) expNode() {}

type String struct {
	Token
	Value string
}

func (String) expNode() {}

type Nil struct {
	Token
}

func (Nil) expNode() {}

type Bool struct {
	Token
	Value bool
}

func (Bool) expNode() {}

type VarArg struct {
	Token
}

func (VarArg) expNode() {}

type UnopExp struct {
	UnopToken Token
	Exp       Exp
}

func (UnopExp) expNode() {}

type BinopExp struct {
	Left       Exp
	BinopToken Token
	Right      Exp
}

func (BinopExp) expNode() {}

type ParenExp struct {
	LParenToken Token
	Exp         Exp
	RParenToken Token
}

func (ParenExp) expNode() {}

type VariableExp struct {
	NameToken Name
}

func (VariableExp) expNode() {}

type TableCtor struct {
	LBraceToken Token
	Fields      FieldList
	RBraceToken Token
}

func (TableCtor) expNode() {}

type FieldList struct {
	Entries []Entry
	Seps    []Token
}

func (l *FieldList) Len() int {
	return len(l.Entries) + len(l.Seps)
}

type Entry interface {
	Node
	entryNode()
}

type IndexEntry struct {
	LBrackToken Token
	KeyExp      Exp
	RBrackToken Token
	AssignToken Token
	ValueExp    Exp
}

func (IndexEntry) entryNode() {}

type FieldEntry struct {
	Name        Name
	AssignToken Token
	Value       Exp
}

func (FieldEntry) entryNode() {}

type ValueEntry struct {
	Value Exp
}

func (ValueEntry) entryNode() {}

type FunctionExp struct {
	FuncToken      Token
	LParenToken    Token
	ParList        *NameList // nil if not present
	VarArgSepToken Token     // INVALID if not present
	VarArgToken    Token     // INVALID if not present
	RParenToken    Token
	Block          Block
	EndToken       Token
}

func (FunctionExp) expNode() {}

type FieldExp struct {
	Exp      Exp
	DotToken Token
	Field    Name
}

func (FieldExp) expNode() {}

type IndexExp struct {
	Exp         Exp
	LBrackToken Token
	Index       Exp
	RBrackToken Token
}

func (IndexExp) expNode() {}

type MethodExp struct {
	Exp        Exp
	ColonToken Token
	Name       Name
	Args       CallArgs
}

func (MethodExp) expNode() {}

type CallExp struct {
	Exp  Exp
	Args CallArgs
}

func (CallExp) expNode() {}

type CallArgs interface {
	Node
	callArgsNode()
}

type ArgsCall struct {
	LParenToken Token
	ExpList     *ExpList // nil if not present
	RParenToken Token
}

func (ArgsCall) callArgsNode() {}

type TableCall struct {
	TableExp TableCtor
}

func (TableCall) callArgsNode() {}

type StringCall struct {
	StringExp String
}

func (StringCall) callArgsNode() {}

type Stat interface {
	Node
	statNode()
}

type DoStat struct {
	DoToken  Token
	Block    Block
	EndToken Token
}

func (DoStat) statNode() {}

type AssignStat struct {
	Left        ExpList
	AssignToken Token
	Right       ExpList
}

func (AssignStat) statNode() {}

type CallExprStat struct {
	Exp Exp
}

func (CallExprStat) statNode() {}

type IfStat struct {
	IfToken       Token
	Exp           Exp
	ThenToken     Token
	Block         Block
	ElseIfClauses []ElseIfClause
	ElseClause    *ElseClause // nil if not present
	EndToken      Token
}

func (IfStat) statNode() {}

type ElseIfClause struct {
	ElseIfToken Token
	Exp         Exp
	ThenToken   Token
	Block       Block
}

type ElseClause struct {
	ElseToken Token
	Block     Block
}

type NumericForStat struct {
	ForToken     Token
	Name         Name
	AssignToken  Token
	MinExp       Exp
	MaxSepToken  Token
	MaxExp       Exp
	StepSepToken Token // INVALID if not present
	StepExp      Exp   // nil if not present
	DoToken      Token
	Block        Block
	EndToken     Token
}

func (NumericForStat) statNode() {}

type GenericForStat struct {
	ForToken Token
	NameList NameList
	InToken  Token
	ExpList  ExpList
	DoToken  Token
	Block    Block
	EndToken Token
}

func (GenericForStat) statNode() {}

type WhileStat struct {
	WhileToken Token
	Exp        Exp
	DoToken    Token
	Block      Block
	EndToken   Token
}

func (WhileStat) statNode() {}

type RepeatStat struct {
	RepeatToken Token
	Block       Block
	UntilToken  Token
	Exp         Exp
}

func (RepeatStat) statNode() {}

type LocalVarStat struct {
	LocalToken  Token
	NameList    NameList
	AssignToken Token    // INVALID if not present
	ExpList     *ExpList // nil if not present
}

func (LocalVarStat) statNode() {}

type LocalFunctionStat struct {
	LocalToken Token
	Name       Name
	Exp        FunctionExp
}

func (LocalFunctionStat) statNode() {}

type FunctionStat struct {
	Name FuncNameList
	Exp  FunctionExp
}

func (FunctionStat) statNode() {}

type FuncNameList struct {
	Names []Name
	Seps  []Token // Separators between names (DOT, COLON).
	// TODO: colon as separate field
}

func (l *FuncNameList) Len() int {
	return len(l.Names) + len(l.Seps)
}

type BreakStat struct {
	BreakToken Token
}

func (BreakStat) statNode() {}

type ReturnStat struct {
	ReturnToken Token
	ExpList     *ExpList // nil if not present
}

func (ReturnStat) statNode() {}
