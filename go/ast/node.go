package ast

import (
	"github.com/anaminus/luasyntax/go/token"
)

type Node interface {
	Start() int
	End() int
}

// Represents a token within a file. Leading whitespace and comments are
// merged into the token, represented by the Space field.
type Token struct {
	Space  int
	Offset int
	Type   token.Token
}

func (t Token) IsValid() bool {
	return t.Type != token.INVALID
}

func (t Token) HasSpace() bool {
	return t.IsValid() && t.Space > 0
}

func (t Token) Start() int {
	if !t.IsValid() {
		return 0
	}
	return t.Space
}
func (t Token) End() int {
	if !t.IsValid() {
		return 0
	}
	return t.Offset + len(t.Type.String())
}

type Exp interface {
	Node
	expNode()
}

type ExpList struct {
	Exps []Exp
	Seps []Token // Commas between expressions.
}

func (l *ExpList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.Exps) + len(l.Seps)
}
func (l *ExpList) Start() int {
	if l == nil {
		return 0
	}
	if len(l.Exps) == 0 {
		if len(l.Seps) == 0 {
			return 0
		}
		return l.Seps[0].Start()
	}
	return l.Exps[0].Start()
}
func (l *ExpList) End() int {
	if l == nil {
		return 0
	}
	j := len(l.Seps)
	i := len(l.Exps)
	if j == 0 && i == 0 {
		return 0
	}
	if j >= i {
		return l.Seps[j-1].End()
	}
	return l.Exps[i-1].End()
}

type Name struct {
	Token
	Value []byte
}

func (Name) expNode()     {}
func (e Name) Start() int { return e.Space }
func (e Name) End() int   { return e.Offset + len(e.Value) }

type NameList struct {
	Names []Name
	Seps  []Token // Separators between names (COMMA, DOT, COLON).
}

func (l *NameList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.Names) + len(l.Seps)
}
func (l *NameList) Start() int {
	if l == nil {
		return 0
	}
	if len(l.Names) == 0 {
		if len(l.Seps) == 0 {
			return 0
		}
		return l.Seps[0].Start()
	}
	return l.Names[0].Start()
}
func (l *NameList) End() int {
	if l == nil {
		return 0
	}
	j := len(l.Seps)
	i := len(l.Names)
	if j == 0 && i == 0 {
		return 0
	}
	if j >= i {
		return l.Seps[j-1].End()
	}
	return l.Names[i-1].End()
}

type Number struct {
	Token
	Value []byte
}

func (Number) expNode()     {}
func (e Number) Start() int { return e.Space }
func (e Number) End() int   { return e.Offset + len(e.Value) }

type String struct {
	Token
	Value []byte
}

func (String) expNode()     {}
func (e String) Start() int { return e.Space }
func (e String) End() int   { return e.Offset + len(e.Value) }

type Nil struct {
	Token
}

func (Nil) expNode()     {}
func (e Nil) Start() int { return e.Space }
func (e Nil) End() int   { return e.Offset + len(e.Type.String()) }

type Bool struct {
	Token
	Value bool
}

func (Bool) expNode()     {}
func (e Bool) Start() int { return e.Space }
func (e Bool) End() int   { return e.Offset + len(e.Type.String()) }

type VarArg struct {
	Token
}

func (VarArg) expNode()     {}
func (e VarArg) Start() int { return e.Space }
func (e VarArg) End() int   { return e.Offset + len(e.Type.String()) }

type UnopExp struct {
	UnopToken Token
	Exp       Exp
}

func (UnopExp) expNode() {}
func (e *UnopExp) Start() int {
	if e == nil {
		return 0
	}
	return e.UnopToken.Start()
}
func (e *UnopExp) End() int {
	if e == nil {
		return 0
	}
	if e.Exp == nil {
		return e.UnopToken.End()
	}
	return e.Exp.End()
}

type BinopExp struct {
	Left       Exp
	BinopToken Token
	Right      Exp
}

func (BinopExp) expNode() {}
func (e *BinopExp) Start() int {
	if e == nil {
		return 0
	}
	if e.Left == nil {
		return e.BinopToken.Start()
	}
	return e.Left.Start()
}
func (e *BinopExp) End() int {
	if e == nil {
		return 0
	}
	if e.Right == nil {
		return e.BinopToken.End()
	}
	return e.Right.End()
}

type ParenExp struct {
	LParenToken Token
	Exp         Exp
	RParenToken Token
}

func (ParenExp) expNode() {}
func (e *ParenExp) Start() int {
	if e == nil {
		return 0
	}
	return e.LParenToken.Start()
}
func (e *ParenExp) End() int {
	if e == nil {
		return 0
	}
	return e.RParenToken.End()
}

type VariableExp struct {
	NameToken Name
}

func (VariableExp) expNode() {}
func (e *VariableExp) Start() int {
	if e == nil {
		return 0
	}
	return e.NameToken.Start()
}
func (e *VariableExp) End() int {
	if e == nil {
		return 0
	}
	return e.NameToken.End()
}

type TableCtor struct {
	LBraceToken Token
	Fields      *FieldList
	RBraceToken Token
}

func (TableCtor) expNode() {}
func (e *TableCtor) Start() int {
	if e == nil {
		return 0
	}
	return e.LBraceToken.Start()
}
func (e *TableCtor) End() int {
	if e == nil {
		return 0
	}
	return e.RBraceToken.End()
}

type FieldList struct {
	Entries []Entry
	Seps    []Token
}

func (l *FieldList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.Entries) + len(l.Seps)
}
func (l *FieldList) Start() int {
	if l == nil {
		return 0
	}
	if len(l.Entries) == 0 {
		if len(l.Seps) == 0 {
			return 0
		}
		return l.Seps[0].Start()
	}
	return l.Entries[0].Start()
}
func (l *FieldList) End() int {
	if l == nil {
		return 0
	}
	j := len(l.Seps)
	i := len(l.Entries)
	if j == 0 && i == 0 {
		return 0
	}
	if j >= i {
		return l.Seps[j-1].End()
	}
	return l.Entries[i-1].End()
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
func (e *IndexEntry) Start() int {
	if e == nil {
		return 0
	}
	return e.LBrackToken.Start()
}
func (e *IndexEntry) End() int {
	if e == nil {
		return 0
	}
	if e.ValueExp == nil {
		return e.AssignToken.End()
	}
	return e.ValueExp.End()
}

type FieldEntry struct {
	Name        Name
	AssignToken Token
	Value       Exp
}

func (FieldEntry) entryNode() {}
func (e *FieldEntry) Start() int {
	if e == nil {
		return 0
	}
	return e.Name.Start()
}
func (e *FieldEntry) End() int {
	if e == nil {
		return 0
	}
	if e.Value == nil {
		return e.AssignToken.End()
	}
	return e.Value.End()
}

type ValueEntry struct {
	Value Exp
}

func (ValueEntry) entryNode() {}
func (e *ValueEntry) Start() int {
	if e == nil || e.Value == nil {
		return 0
	}
	return e.Value.Start()
}
func (e *ValueEntry) End() int {
	if e == nil || e.Value == nil {
		return 0
	}
	return e.Value.End()
}

type FieldExp struct {
	Exp      Exp
	DotToken Token
	Field    Name
}

func (FieldExp) expNode() {}
func (e *FieldExp) Start() int {
	if e == nil {
		return 0
	}
	if e.Exp == nil {
		return e.DotToken.Start()
	}
	return e.Exp.Start()
}
func (e *FieldExp) End() int {
	if e == nil {
		return 0
	}
	return e.Field.End()
}

type IndexExp struct {
	Exp         Exp
	LBrackToken Token
	Index       Exp
	RBrackToken Token
}

func (IndexExp) expNode() {}
func (e *IndexExp) Start() int {
	if e == nil {
		return 0
	}
	if e.Exp == nil {
		return e.LBrackToken.Start()
	}
	return e.Exp.Start()
}
func (e *IndexExp) End() int {
	if e == nil {
		return 0
	}
	return e.RBrackToken.End()
}

type MethodExp struct {
	Exp        Exp
	ColonToken Token
	Name       Name
	Args       CallArgs
}

func (MethodExp) expNode() {}
func (e *MethodExp) Start() int {
	if e == nil {
		return 0
	}
	if e.Exp == nil {
		return e.ColonToken.Start()
	}
	return e.Exp.End()
}
func (e *MethodExp) End() int {
	if e == nil {
		return 0
	}
	if e.Args == nil {
		return e.Name.End()
	}
	return e.Args.End()
}

type CallExp struct {
	Exp  Exp
	Args CallArgs
}

func (CallExp) expNode() {}
func (e *CallExp) Start() int {
	if e == nil {
		return 0
	}
	if e.Exp == nil {
		if e.Args == nil {
			return 0
		}
		return e.Args.Start()
	}
	return e.Exp.Start()
}
func (e *CallExp) End() int {
	if e == nil {
		return 0
	}
	if e.Args == nil {
		if e.Exp == nil {
			return 0
		}
		return e.Exp.End()
	}
	return e.Args.End()
}

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
func (c *ArgsCall) Start() int {
	if c == nil {
		return 0
	}
	return c.LParenToken.Start()
}
func (c *ArgsCall) End() int {
	if c == nil {
		return 0
	}
	return c.RParenToken.End()
}

type TableCall struct {
	TableExp *TableCtor
}

func (TableCall) callArgsNode() {}
func (c *TableCall) Start() int {
	if c == nil || c.TableExp == nil {
		return 0
	}
	return c.TableExp.Start()
}
func (c *TableCall) End() int {
	if c == nil || c.TableExp == nil {
		return 0
	}
	return c.TableExp.End()
}

type StringCall struct {
	StringExp String
}

func (StringCall) callArgsNode() {}
func (c *StringCall) Start() int {
	if c == nil {
		return 0
	}
	return c.StringExp.Start()
}
func (c *StringCall) End() int {
	if c == nil {
		return 0
	}
	return c.StringExp.End()
}

type File struct {
	Name  string
	Block *Block
}

func (f *File) Start() int {
	if f == nil || f.Block == nil {
		return 0
	}
	return f.Block.Start()
}
func (f *File) End() int {
	if f == nil || f.Block == nil {
		return 0
	}
	return f.Block.End()
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
func (b *Block) Start() int {
	if b == nil {
		return 0
	}
	if len(b.Stats) == 0 {
		if len(b.Seps) == 0 {
			return 0
		}
		return b.Seps[0].Start()
	}
	return b.Stats[0].Start()
}
func (b *Block) End() int {
	if b == nil {
		return 0
	}
	j := len(b.Seps)
	i := len(b.Stats)
	if j == 0 && i == 0 {
		return 0
	}
	if j >= i {
		return b.Seps[j-1].End()
	}
	return b.Stats[i-1].End()
}

type Stat interface {
	Node
	statNode()
}

type DoStat struct {
	DoToken  Token
	Block    *Block
	EndToken Token
}

func (DoStat) statNode() {}
func (s *DoStat) Start() int {
	if s == nil {
		return 0
	}
	return s.DoToken.Start()
}
func (s *DoStat) End() int {
	if s == nil {
		return 0
	}
	return s.EndToken.End()
}

type AssignStat struct {
	Left        *ExpList
	AssignToken Token
	Right       *ExpList
}

func (AssignStat) statNode() {}
func (s *AssignStat) Start() int {
	if s == nil {
		return 0
	}
	if s.Left.Len() == 0 {
		return s.AssignToken.Start()
	}
	return s.Left.Start()
}
func (s *AssignStat) End() int {
	if s == nil {
		return 0
	}
	if s.Right.Len() == 0 {
		return s.AssignToken.End()
	}
	return s.Right.End()
}

type CallExprStat struct {
	Exp Exp
}

func (CallExprStat) statNode() {}
func (s *CallExprStat) Start() int {
	if s == nil || s.Exp == nil {
		return 0
	}
	return s.Exp.Start()
}
func (s *CallExprStat) End() int {
	if s == nil || s.Exp == nil {
		return 0
	}
	return s.Exp.End()
}

type IfStat struct {
	IfToken       Token
	Exp           Exp
	ThenToken     Token
	Block         *Block
	ElseIfClauses []*ElseIfClause
	ElseClause    *ElseClause
	EndToken      Token
}

func (IfStat) statNode() {}
func (s *IfStat) Start() int {
	if s == nil {
		return 0
	}
	return s.IfToken.Start()
}
func (s *IfStat) End() int {
	if s == nil {
		return 0
	}
	return s.EndToken.End()
}

type ElseIfClause struct {
	ElseIfToken Token
	Exp         Exp
	ThenToken   Token
	Block       *Block
}

func (c *ElseIfClause) Start() int {
	if c == nil {
		return 0
	}
	return c.ElseIfToken.Start()
}
func (c *ElseIfClause) End() int {
	if c == nil {
		return 0
	}
	if c.Block.Len() == 0 {
		return c.ThenToken.End()
	}
	return c.Block.End()
}

type ElseClause struct {
	ElseToken Token
	Block     *Block
}

func (c *ElseClause) Start() int {
	if c == nil {
		return 0
	}
	return c.ElseToken.Start()
}
func (c *ElseClause) End() int {
	if c == nil {
		return 0
	}
	if c.Block.Len() == 0 {
		return c.ElseToken.End()
	}
	return c.Block.End()
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
	Block        *Block
	EndToken     Token
}

func (NumericForStat) statNode() {}
func (s *NumericForStat) Start() int {
	if s == nil {
		return 0
	}
	return s.ForToken.Start()
}
func (s *NumericForStat) End() int {
	if s == nil {
		return 0
	}
	return s.EndToken.End()
}

type GenericForStat struct {
	ForToken Token
	NameList *NameList
	InToken  Token
	ExpList  *ExpList
	DoToken  Token
	Block    *Block
	EndToken Token
}

func (GenericForStat) statNode() {}
func (s *GenericForStat) Start() int {
	if s == nil {
		return 0
	}
	return s.ForToken.Start()
}
func (s *GenericForStat) End() int {
	if s == nil {
		return 0
	}
	return s.EndToken.End()
}

type WhileStat struct {
	WhileToken Token
	Exp        Exp
	DoToken    Token
	Block      *Block
	EndToken   Token
}

func (WhileStat) statNode() {}
func (s *WhileStat) Start() int {
	if s == nil {
		return 0
	}
	return s.WhileToken.Start()
}
func (s *WhileStat) End() int {
	if s == nil {
		return 0
	}
	return s.EndToken.End()
}

type RepeatStat struct {
	RepeatToken Token
	Block       *Block
	UntilToken  Token
	Exp         Exp
}

func (RepeatStat) statNode() {}
func (s *RepeatStat) Start() int {
	if s == nil {
		return 0
	}
	return s.RepeatToken.Start()
}
func (s *RepeatStat) End() int {
	if s == nil {
		return 0
	}
	if s.Exp == nil {
		return s.UntilToken.End()
	}
	return s.Exp.End()
}

type LocalVarStat struct {
	LocalToken  Token
	NameList    *NameList
	AssignToken Token    // INVALID if not present
	ExpList     *ExpList // nil if not present
}

func (LocalVarStat) statNode() {}
func (s *LocalVarStat) Start() int {
	if s == nil {
		return 0
	}
	return s.LocalToken.Start()
}
func (s *LocalVarStat) End() int {
	if s == nil {
		return 0
	}
	if s.ExpList.Len() == 0 {
		return s.AssignToken.End()
	}
	return s.ExpList.End()
}

type LocalFunctionStat struct {
	LocalToken Token
	Function   Function
}

func (LocalFunctionStat) statNode() {}
func (s *LocalFunctionStat) Start() int {
	if s == nil {
		return 0
	}
	return s.LocalToken.Start()
}
func (s *LocalFunctionStat) End() int {
	if s == nil {
		return 0
	}
	return s.Function.End()
}

type Function struct {
	FuncToken      Token
	FuncName       *NameList // nil if anonymous.
	LParenToken    Token
	ParList        *NameList
	VarArgSepToken Token // INVALID if not present
	VarArgToken    Token // INVALID if not present
	RParenToken    Token
	Block          *Block
	EndToken       Token
}

func (Function) statNode() {}
func (Function) expNode()  {}
func (s *Function) Start() int {
	if s == nil {
		return 0
	}
	return s.FuncToken.Start()
}
func (s *Function) End() int {
	if s == nil {
		return 0
	}
	return s.EndToken.End()
}

type BreakStat struct {
	BreakToken Token
}

func (BreakStat) statNode() {}
func (s *BreakStat) Start() int {
	if s == nil {
		return 0
	}
	return s.BreakToken.Start()
}
func (s *BreakStat) End() int {
	if s == nil {
		return 0
	}
	return s.BreakToken.End()
}

type ReturnStat struct {
	ReturnToken Token
	ExpList     *ExpList // nil if not present
}

func (ReturnStat) statNode() {}
func (s *ReturnStat) Start() int {
	if s == nil {
		return 0
	}
	return s.ReturnToken.Start()
}
func (s *ReturnStat) End() int {
	if s == nil {
		return 0
	}
	if s.ExpList.Len() == 0 {
		return s.ReturnToken.End()
	}
	return s.ExpList.End()
}
