package ast

import (
	"github.com/anaminus/luasyntax/go/token"
)

// Note: A node is valid if
//     - interface fields are non-nil.
//     - token fields have types acceptable for the node.

func isv(i interface{}) bool               { return i != nil }
func ist(tok Token, typ token.Type) bool   { return tok.Type == typ }
func ist2(tok Token, a, b token.Type) bool { return tok.Type == a || tok.Type == b }

func (f *File) IsValid() bool {
	return ist(f.EOFToken, token.EOF)
}

func (b *Block) IsValid() bool {
	if len(b.Seps) != len(b.Stats) {
		return false
	}
	for _, stat := range b.Stats {
		if !isv(stat) {
			return false
		}
	}
	for _, sep := range b.Seps {
		if !ist2(sep, token.SEMICOLON, token.INVALID) {
			return false
		}
	}
	return true
}

func (l *ExpList) IsValid() bool {
	if len(l.Exps) == 0 || len(l.Seps) != len(l.Exps)-1 {
		return false
	}
	for _, exp := range l.Exps {
		if !isv(exp) {
			return false
		}
	}
	for _, sep := range l.Seps {
		if !ist(sep, token.COMMA) {
			return false
		}
	}
	return true
}

func (e *Name) IsValid() bool {
	return ist(e.Token, token.NAME)
}

func (l *NameList) IsValid() bool {
	if len(l.Names) == 0 || len(l.Seps) != len(l.Names)-1 {
		return false
	}
	for _, name := range l.Seps {
		if !ist(name, token.NAME) {
			return false
		}
	}
	for _, sep := range l.Seps {
		if !ist(sep, token.COMMA) {
			return false
		}
	}
	return true
}

func (e *Number) IsValid() bool {
	return e.Token.Type.IsNumber()
}

func (e *String) IsValid() bool {
	return e.Token.Type.IsString()
}

func (e *Nil) IsValid() bool {
	return ist(e.Token, token.NIL)
}

func (e *Bool) IsValid() bool {
	return e.Token.Type.IsBool()
}

func (e *VarArg) IsValid() bool {
	return ist(e.Token, token.VARARG)
}

func (e *UnopExp) IsValid() bool {
	return e.UnopToken.Type.IsUnary() &&
		isv(e.Exp)
}

func (e *BinopExp) IsValid() bool {
	return isv(e.Left) &&
		e.BinopToken.Type.IsBinary() &&
		isv(e.Right)
}

func (e *ParenExp) IsValid() bool {
	return ist(e.LParenToken, token.LPAREN) &&
		isv(e.Exp) &&
		ist(e.RParenToken, token.RPAREN)
}

func (e *VariableExp) IsValid() bool {
	return ist(e.NameToken.Token, token.NAME)
}

func (e *TableCtor) IsValid() bool {
	return ist(e.LBraceToken, token.LBRACE) &&
		ist(e.RBraceToken, token.RBRACE)
}

func (l *EntryList) IsValid() bool {
	if len(l.Seps) != len(l.Entries) && len(l.Seps) != len(l.Entries)-1 {
		return false
	}
	for _, entry := range l.Entries {
		if entry == nil {
			return false
		}
	}
	for _, sep := range l.Seps {
		if !ist2(sep, token.COMMA, token.SEMICOLON) {
			return false
		}
	}
	return true
}

func (e *IndexEntry) IsValid() bool {
	return ist(e.LBrackToken, token.LBRACK) &&
		isv(e.KeyExp) &&
		ist(e.RBrackToken, token.RBRACK) &&
		ist(e.AssignToken, token.ASSIGN) &&
		isv(e.ValueExp)
}

func (e *FieldEntry) IsValid() bool {
	return ist(e.Name.Token, token.NAME) &&
		ist(e.AssignToken, token.ASSIGN) &&
		isv(e.Value)
}

func (e *ValueEntry) IsValid() bool {
	return isv(e.Value)
}

func (e *FunctionExp) IsValid() bool {
	if !(ist(e.FuncToken, token.FUNCTION) &&
		ist(e.LParenToken, token.LPAREN) &&
		ist(e.RParenToken, token.RPAREN) &&
		ist(e.EndToken, token.END)) {
		return false
	}
	if ist(e.VarArgToken, token.VARARG) {
		if e.ParList != nil {
			return ist(e.VarArgSepToken, token.COMMA)
		}
		return ist(e.VarArgSepToken, token.INVALID)
	} else if ist(e.VarArgToken, token.INVALID) {
		return ist(e.VarArgSepToken, token.INVALID)
	}
	return false
}

func (e *FieldExp) IsValid() bool {
	return isv(e.Exp) &&
		ist(e.DotToken, token.DOT) &&
		ist(e.Field.Token, token.NAME)
}

func (e *IndexExp) IsValid() bool {
	return isv(e.Exp) &&
		ist(e.LBrackToken, token.LBRACK) &&
		isv(e.Index) &&
		ist(e.RBrackToken, token.RBRACK)
}

func (e *MethodExp) IsValid() bool {
	return isv(e.Exp) &&
		ist(e.ColonToken, token.COLON) &&
		ist(e.Name.Token, token.NAME) &&
		isv(e.Args)
}

func (e *CallExp) IsValid() bool {
	return isv(e.Exp) &&
		isv(e.Args)
}

func (c *ArgsCall) IsValid() bool {
	return ist(c.LParenToken, token.LPAREN) &&
		ist(c.RParenToken, token.RPAREN)
}

func (c *TableCall) IsValid() bool {
	return true
}

func (c *StringCall) IsValid() bool {
	return c.StringExp.Type.IsString()
}

func (s *DoStat) IsValid() bool {
	return ist(s.DoToken, token.DO) &&
		ist(s.EndToken, token.END)
}

func (s *AssignStat) IsValid() bool {
	return ist(s.AssignToken, token.ASSIGN)
}

func (s *CallExprStat) IsValid() bool {
	return isv(s.Exp)
}

func (s *IfStat) IsValid() bool {
	return ist(s.IfToken, token.IF) &&
		isv(s.Exp) &&
		ist(s.ThenToken, token.THEN) &&
		ist(s.EndToken, token.END)
}

func (c *ElseIfClause) IsValid() bool {
	return ist(c.ElseIfToken, token.ELSEIF) &&
		isv(c.Exp) &&
		ist(c.ThenToken, token.THEN)
}

func (c *ElseClause) IsValid() bool {
	return ist(c.ElseToken, token.ELSE)
}

func (s *NumericForStat) IsValid() bool {
	if !(ist(s.ForToken, token.FOR) &&
		ist(s.Name.Token, token.NAME) &&
		ist(s.AssignToken, token.ASSIGN) &&
		isv(s.MinExp) &&
		ist(s.MaxSepToken, token.COMMA) &&
		isv(s.MaxExp) &&
		ist2(s.StepSepToken, token.COMMA, token.INVALID) &&
		ist(s.DoToken, token.DO) &&
		ist(s.EndToken, token.END)) {
		return false
	}
	if ist(s.StepSepToken, token.COMMA) {
		return isv(s.StepExp)
	} else if ist(s.StepSepToken, token.INVALID) {
		return !isv(s.StepExp)
	}
	return false
}

func (s *GenericForStat) IsValid() bool {
	return ist(s.ForToken, token.FOR) &&
		ist(s.InToken, token.IN) &&
		ist(s.DoToken, token.DO) &&
		ist(s.EndToken, token.END)
}

func (s *WhileStat) IsValid() bool {
	return ist(s.WhileToken, token.WHILE) &&
		isv(s.Exp) &&
		ist(s.DoToken, token.DO) &&
		ist(s.EndToken, token.END)
}

func (s *RepeatStat) IsValid() bool {
	return ist(s.RepeatToken, token.REPEAT) &&
		ist(s.UntilToken, token.UNTIL) &&
		isv(s.Exp)
}

func (s *LocalVarStat) IsValid() bool {
	if !ist(s.LocalToken, token.LOCAL) {
		return false
	}
	if ist(s.AssignToken, token.ASSIGN) {
		return s.ExpList != nil
	} else if ist(s.AssignToken, token.INVALID) {
		return s.ExpList == nil
	}
	return false
}

func (s *LocalFunctionStat) IsValid() bool {
	return ist(s.LocalToken, token.LOCAL) &&
		ist(s.Name.Token, token.NAME)
}

func (s *FunctionStat) IsValid() bool {
	return true
}

func (l *FuncNameList) IsValid() bool {
	if len(l.Names) == 0 || len(l.Seps) != len(l.Names)-1 {
		return false
	}
	if len(l.Seps) > 0 {
		for i := 0; i < len(l.Seps)-1; i++ {
			if ist(l.Seps[i], token.DOT) {
				return false
			}
		}
		if ist2(l.Seps[len(l.Seps)-1], token.DOT, token.COLON) {
			return false
		}
	}
	return true
}

func (s *BreakStat) IsValid() bool {
	return ist(s.BreakToken, token.BREAK)
}

func (s *ReturnStat) IsValid() bool {
	return ist(s.ReturnToken, token.RETURN)
}
