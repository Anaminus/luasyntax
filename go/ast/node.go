package ast

func (f *File) FirstToken() *Token { return f.Body.FirstToken() }
func (f *File) LastToken() *Token  { return &f.EOFToken }

func (b *Block) FirstToken() *Token {
	if len(b.Items) == 0 {
		return nil
	}
	return b.Items[0].FirstToken()
}
func (b *Block) LastToken() *Token {
	if len(b.Items) == 0 {
		return nil
	}
	if len(b.Seps) == len(b.Items) {
		return &b.Seps[len(b.Seps)-1]
	}
	return b.Items[len(b.Items)-1].LastToken()
}

func (l *ExprList) FirstToken() *Token { return l.Items[0].FirstToken() }
func (l *ExprList) LastToken() *Token  { return l.Items[len(l.Items)-1].LastToken() }

func (e *Name) FirstToken() *Token { return &e.Token }
func (e *Name) LastToken() *Token  { return &e.Token }

func (l *NameList) FirstToken() *Token { return &l.Items[0].Token }
func (l *NameList) LastToken() *Token  { return &l.Items[len(l.Items)-1].Token }

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

func (e *UnopExpr) FirstToken() *Token { return &e.UnopToken }
func (e *UnopExpr) LastToken() *Token  { return e.Operand.LastToken() }

func (e *BinopExpr) FirstToken() *Token { return e.Left.FirstToken() }
func (e *BinopExpr) LastToken() *Token  { return e.Right.LastToken() }

func (e *ParenExpr) FirstToken() *Token { return &e.LParenToken }
func (e *ParenExpr) LastToken() *Token  { return &e.RParenToken }

func (e *VariableExpr) FirstToken() *Token { return &e.Name.Token }
func (e *VariableExpr) LastToken() *Token  { return &e.Name.Token }

func (e *TableCtor) FirstToken() *Token { return &e.LBraceToken }
func (e *TableCtor) LastToken() *Token  { return &e.RBraceToken }

func (l *EntryList) FirstToken() *Token {
	if len(l.Items) == 0 {
		return nil
	}
	return l.Items[0].FirstToken()
}
func (l *EntryList) LastToken() *Token {
	if len(l.Seps) == len(l.Items) {
		return &l.Seps[len(l.Seps)-1]
	}
	return l.Items[len(l.Items)-1].LastToken()
}

func (e *IndexEntry) FirstToken() *Token { return &e.LBrackToken }
func (e *IndexEntry) LastToken() *Token  { return e.Value.LastToken() }

func (e *FieldEntry) FirstToken() *Token { return &e.Name.Token }
func (e *FieldEntry) LastToken() *Token  { return e.Value.LastToken() }

func (e *ValueEntry) FirstToken() *Token { return e.Value.FirstToken() }
func (e *ValueEntry) LastToken() *Token  { return e.Value.LastToken() }

func (s *FunctionExpr) FirstToken() *Token { return &s.FuncToken }
func (s *FunctionExpr) LastToken() *Token  { return &s.EndToken }

func (e *FieldExpr) FirstToken() *Token { return e.Value.FirstToken() }
func (e *FieldExpr) LastToken() *Token  { return &e.Name.Token }

func (e *IndexExpr) FirstToken() *Token { return e.Value.FirstToken() }
func (e *IndexExpr) LastToken() *Token  { return &e.RBrackToken }

func (e *MethodExpr) FirstToken() *Token { return e.Value.FirstToken() }
func (e *MethodExpr) LastToken() *Token  { return e.Args.LastToken() }

func (e *CallExpr) FirstToken() *Token { return e.Value.FirstToken() }
func (e *CallExpr) LastToken() *Token  { return e.Args.LastToken() }

func (c *ListArgs) FirstToken() *Token { return &c.LParenToken }
func (c *ListArgs) LastToken() *Token  { return &c.RParenToken }

func (c *TableArg) FirstToken() *Token { return c.Value.FirstToken() }
func (c *TableArg) LastToken() *Token  { return c.Value.LastToken() }

func (c *StringArg) FirstToken() *Token { return &c.Value.Token }
func (c *StringArg) LastToken() *Token  { return &c.Value.Token }

func (s *DoStmt) FirstToken() *Token { return &s.DoToken }
func (s *DoStmt) LastToken() *Token  { return &s.EndToken }

func (s *AssignStmt) FirstToken() *Token { return s.Left.FirstToken() }
func (s *AssignStmt) LastToken() *Token  { return s.Right.LastToken() }

func (s *CallExprStmt) FirstToken() *Token { return s.Call.FirstToken() }
func (s *CallExprStmt) LastToken() *Token  { return s.Call.LastToken() }

func (s *IfStmt) FirstToken() *Token { return &s.IfToken }
func (s *IfStmt) LastToken() *Token  { return &s.EndToken }

func (c *ElseIfClause) FirstToken() *Token { return &c.ElseIfToken }
func (c *ElseIfClause) LastToken() *Token  { return c.Body.LastToken() }

func (c *ElseClause) FirstToken() *Token { return &c.ElseToken }
func (c *ElseClause) LastToken() *Token  { return c.Body.LastToken() }

func (s *NumericForStmt) FirstToken() *Token { return &s.ForToken }
func (s *NumericForStmt) LastToken() *Token  { return &s.EndToken }

func (s *GenericForStmt) FirstToken() *Token { return &s.ForToken }
func (s *GenericForStmt) LastToken() *Token  { return &s.EndToken }

func (s *WhileStmt) FirstToken() *Token { return &s.WhileToken }
func (s *WhileStmt) LastToken() *Token  { return &s.EndToken }

func (s *RepeatStmt) FirstToken() *Token { return &s.RepeatToken }
func (s *RepeatStmt) LastToken() *Token  { return s.Cond.LastToken() }

func (s *LocalVarStmt) FirstToken() *Token { return &s.LocalToken }
func (s *LocalVarStmt) LastToken() *Token {
	if !s.AssignToken.Type.IsValid() {
		return s.Names.LastToken()
	}
	return s.Values.LastToken()
}

func (s *LocalFunctionStmt) FirstToken() *Token { return &s.LocalToken }
func (s *LocalFunctionStmt) LastToken() *Token  { return s.Func.LastToken() }

func (s *FunctionStmt) FirstToken() *Token { return s.Func.FirstToken() }
func (s *FunctionStmt) LastToken() *Token  { return s.Func.LastToken() }

func (s *BreakStmt) FirstToken() *Token { return &s.BreakToken }
func (s *BreakStmt) LastToken() *Token  { return &s.BreakToken }

func (s *ReturnStmt) FirstToken() *Token { return &s.ReturnToken }
func (s *ReturnStmt) LastToken() *Token {
	if s.Values.Len() == 0 {
		return &s.ReturnToken
	}
	return s.Values.LastToken()
}
