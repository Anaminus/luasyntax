package ast

func (f *File) FirstToken() *Token { return f.Block.FirstToken() }
func (f *File) LastToken() *Token  { return &f.EOFToken }

func (b *Block) FirstToken() *Token {
	if len(b.Stats) == 0 {
		return nil
	}
	return b.Stats[0].FirstToken()
}
func (b *Block) LastToken() *Token {
	if len(b.Stats) == 0 {
		return nil
	}
	if len(b.Seps) == len(b.Stats) {
		return &b.Seps[len(b.Seps)-1]
	}
	return b.Stats[len(b.Stats)-1].LastToken()
}

func (l *ExpList) FirstToken() *Token { return l.Exps[0].FirstToken() }
func (l *ExpList) LastToken() *Token  { return l.Exps[len(l.Exps)-1].LastToken() }

func (e *Name) FirstToken() *Token { return &e.Token }
func (e *Name) LastToken() *Token  { return &e.Token }

func (l *NameList) FirstToken() *Token { return &l.Names[0].Token }
func (l *NameList) LastToken() *Token  { return &l.Names[len(l.Names)-1].Token }

func (e *Number) FirstToken() *Token { return &e.Token }
func (e *Number) LastToken() *Token  { return &e.Token }

func (e *String) FirstToken() *Token { return &e.Token }
func (e *String) LastToken() *Token  { return &e.Token }

func (e *Nil) FirstToken() *Token { return &e.Token }
func (e *Nil) LastToken() *Token  { return &e.Token }

func (e *Bool) FirstToken() *Token { return &e.Token }
func (e *Bool) LastToken() *Token  { return &e.Token }

func (e *VarArg) FirstToken() *Token { return &e.Token }
func (e *VarArg) LastToken() *Token  { return &e.Token }

func (e *UnopExp) FirstToken() *Token { return &e.UnopToken }
func (e *UnopExp) LastToken() *Token  { return e.Exp.LastToken() }

func (e *BinopExp) FirstToken() *Token { return e.Left.FirstToken() }
func (e *BinopExp) LastToken() *Token  { return e.Right.LastToken() }

func (e *ParenExp) FirstToken() *Token { return &e.LParenToken }
func (e *ParenExp) LastToken() *Token  { return &e.RParenToken }

func (e *VariableExp) FirstToken() *Token { return &e.NameToken.Token }
func (e *VariableExp) LastToken() *Token  { return &e.NameToken.Token }

func (e *TableCtor) FirstToken() *Token { return &e.LBraceToken }
func (e *TableCtor) LastToken() *Token  { return &e.RBraceToken }

func (l *EntryList) FirstToken() *Token {
	if len(l.Entries) == 0 {
		return nil
	}
	return l.Entries[0].FirstToken()
}
func (l *EntryList) LastToken() *Token {
	if len(l.Seps) == len(l.Entries) {
		return &l.Seps[len(l.Seps)-1]
	}
	return l.Entries[len(l.Entries)-1].LastToken()
}

func (e *IndexEntry) FirstToken() *Token { return &e.LBrackToken }
func (e *IndexEntry) LastToken() *Token  { return e.ValueExp.LastToken() }

func (e *FieldEntry) FirstToken() *Token { return &e.Name.Token }
func (e *FieldEntry) LastToken() *Token  { return e.Value.LastToken() }

func (e *ValueEntry) FirstToken() *Token { return e.Value.FirstToken() }
func (e *ValueEntry) LastToken() *Token  { return e.Value.LastToken() }

func (s *FunctionExp) FirstToken() *Token { return &s.FuncToken }
func (s *FunctionExp) LastToken() *Token  { return &s.EndToken }

func (e *FieldExp) FirstToken() *Token { return e.Exp.FirstToken() }
func (e *FieldExp) LastToken() *Token  { return &e.Field.Token }

func (e *IndexExp) FirstToken() *Token { return e.Exp.FirstToken() }
func (e *IndexExp) LastToken() *Token  { return &e.RBrackToken }

func (e *MethodExp) FirstToken() *Token { return e.Exp.FirstToken() }
func (e *MethodExp) LastToken() *Token  { return e.Args.LastToken() }

func (e *CallExp) FirstToken() *Token { return e.Exp.FirstToken() }
func (e *CallExp) LastToken() *Token  { return e.Args.LastToken() }

func (c *ArgsCall) FirstToken() *Token { return &c.LParenToken }
func (c *ArgsCall) LastToken() *Token  { return &c.RParenToken }

func (c *TableCall) FirstToken() *Token { return c.TableExp.FirstToken() }
func (c *TableCall) LastToken() *Token  { return c.TableExp.LastToken() }

func (c *StringCall) FirstToken() *Token { return &c.StringExp.Token }
func (c *StringCall) LastToken() *Token  { return &c.StringExp.Token }

func (s *DoStat) FirstToken() *Token { return &s.DoToken }
func (s *DoStat) LastToken() *Token  { return &s.EndToken }

func (s *AssignStat) FirstToken() *Token { return s.Left.FirstToken() }
func (s *AssignStat) LastToken() *Token  { return s.Right.LastToken() }

func (s *CallExprStat) FirstToken() *Token { return s.Exp.FirstToken() }
func (s *CallExprStat) LastToken() *Token  { return s.Exp.LastToken() }

func (s *IfStat) FirstToken() *Token { return &s.IfToken }
func (s *IfStat) LastToken() *Token  { return &s.EndToken }

func (c *ElseIfClause) FirstToken() *Token { return &c.ElseIfToken }
func (c *ElseIfClause) LastToken() *Token  { return c.Block.LastToken() }

func (c *ElseClause) FirstToken() *Token { return &c.ElseToken }
func (c *ElseClause) LastToken() *Token  { return c.Block.LastToken() }

func (s *NumericForStat) FirstToken() *Token { return &s.ForToken }
func (s *NumericForStat) LastToken() *Token  { return &s.EndToken }

func (s *GenericForStat) FirstToken() *Token { return &s.ForToken }
func (s *GenericForStat) LastToken() *Token  { return &s.EndToken }

func (s *WhileStat) FirstToken() *Token { return &s.WhileToken }
func (s *WhileStat) LastToken() *Token  { return &s.EndToken }

func (s *RepeatStat) FirstToken() *Token { return &s.RepeatToken }
func (s *RepeatStat) LastToken() *Token  { return s.Exp.LastToken() }

func (s *LocalVarStat) FirstToken() *Token { return &s.LocalToken }
func (s *LocalVarStat) LastToken() *Token {
	if !s.AssignToken.Type.IsValid() {
		return s.NameList.LastToken()
	}
	return s.ExpList.LastToken()
}

func (s *LocalFunctionStat) FirstToken() *Token { return &s.LocalToken }
func (s *LocalFunctionStat) LastToken() *Token  { return s.Exp.LastToken() }

func (s *FunctionStat) FirstToken() *Token { return s.Exp.FirstToken() }
func (s *FunctionStat) LastToken() *Token  { return s.Exp.LastToken() }

func (s *BreakStat) FirstToken() *Token { return &s.BreakToken }
func (s *BreakStat) LastToken() *Token  { return &s.BreakToken }

func (s *ReturnStat) FirstToken() *Token { return &s.ReturnToken }
func (s *ReturnStat) LastToken() *Token {
	if s.ExpList.Len() == 0 {
		return &s.ReturnToken
	}
	return s.ExpList.LastToken()
}
