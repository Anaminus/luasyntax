package ast

import (
	"io"
)

// copier handles multiple writes in sequence, then allows the accumulated
// result to be returned.
type copier struct {
	n   int64
	err error
}

// writeTo calls wt.WriteTo(w), and acculumates the results. Returns whether
// an error occurred.
func (c *copier) writeTo(w io.Writer, wt io.WriterTo) bool {
	if c.err != nil {
		return false
	}
	if wt == nil {
		return true
	}
	var n int64
	n, c.err = wt.WriteTo(w)
	c.n += n
	return c.err == nil
}

// write writes p to the Writer w, and accumulates the results. Returns
// whether can error occurred.
func (c *copier) write(w io.Writer, p []byte) bool {
	if c.err != nil {
		return false
	}
	if p == nil {
		return true
	}
	var n int
	n, c.err = w.Write(p)
	c.n += int64(n)
	return c.err == nil
}

// finish returns the accumulated results of every write.
func (c *copier) finish() (n int64, err error) {
	return c.n, c.err
}

func (t Token) WriteTo(w io.Writer) (n int64, err error) {
	// TODO: Make this a pointer receiver?
	var c copier
	for _, p := range t.Prefix {
		if !c.write(w, p.Bytes) {
			break
		}
	}
	c.write(w, t.Bytes)
	return c.finish()
}

func (f *File) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, &f.Body)
	c.writeTo(w, f.EOFToken)
	return c.finish()
}

func (l *ExprList) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, expr := range l.Items {
		if !c.writeTo(w, expr) {
			break
		}
		if i < len(l.Seps) && l.Seps[i].Type.IsValid() {
			if !c.writeTo(w, l.Seps[i]) {
				break
			}
		}
	}
	return c.finish()
}

func (l *NameList) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, name := range l.Items {
		if !c.writeTo(w, name) {
			break
		}
		if i < len(l.Seps) && l.Seps[i].Type.IsValid() {
			if !c.writeTo(w, l.Seps[i]) {
				break
			}
		}
	}
	return c.finish()
}

func (e *UnopExpr) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.UnopToken)
	c.writeTo(w, e.Operand)
	return c.finish()
}

func (e *BinopExpr) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Left)
	c.writeTo(w, e.BinopToken)
	c.writeTo(w, e.Right)
	return c.finish()
}

func (e *ParenExpr) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.LParenToken)
	c.writeTo(w, e.Value)
	c.writeTo(w, e.RParenToken)
	return c.finish()
}

func (e *VariableExpr) WriteTo(w io.Writer) (n int64, err error) {
	return e.Name.WriteTo(w)
}

func (e *TableCtor) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.LBraceToken)
	c.writeTo(w, &e.Entries)
	c.writeTo(w, e.RBraceToken)
	return c.finish()
}

func (l *EntryList) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, entry := range l.Items {
		if !c.writeTo(w, entry) {
			break
		}
		if i < len(l.Seps) && l.Seps[i].Type.IsValid() {
			if !c.writeTo(w, l.Seps[i]) {
				break
			}
		}
	}
	return c.finish()
}

func (e *IndexEntry) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.LBrackToken)
	c.writeTo(w, e.Key)
	c.writeTo(w, e.RBrackToken)
	c.writeTo(w, e.AssignToken)
	c.writeTo(w, e.Value)
	return c.finish()
}

func (e *FieldEntry) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Name)
	c.writeTo(w, e.AssignToken)
	c.writeTo(w, e.Value)
	return c.finish()
}

func (e *ValueEntry) WriteTo(w io.Writer) (n int64, err error) {
	return e.Value.WriteTo(w)
}

func (e *FunctionExpr) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.FuncToken)
	c.writeTo(w, e.LParenToken)
	if e.Params != nil {
		c.writeTo(w, e.Params)
		if e.VarArgSepToken.Type.IsValid() {
			c.writeTo(w, e.VarArgSepToken)
			c.writeTo(w, e.VarArgToken)
		}
	} else if e.VarArgToken.Type.IsValid() {
		c.writeTo(w, e.VarArgToken)
	}
	c.writeTo(w, e.RParenToken)
	c.writeTo(w, &e.Body)
	c.writeTo(w, e.EndToken)
	return c.finish()
}

func (e *FieldExpr) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Value)
	c.writeTo(w, e.DotToken)
	c.writeTo(w, e.Name)
	return c.finish()
}

func (e *IndexExpr) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Value)
	c.writeTo(w, e.LBrackToken)
	c.writeTo(w, e.Index)
	c.writeTo(w, e.RBrackToken)
	return c.finish()
}

func (e *MethodExpr) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Value)
	c.writeTo(w, e.ColonToken)
	c.writeTo(w, e.Name)
	c.writeTo(w, e.Args)
	return c.finish()
}

func (e *CallExpr) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Value)
	c.writeTo(w, e.Args)
	return c.finish()
}

func (ac *ListArgs) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, ac.LParenToken)
	if ac.Values != nil {
		c.writeTo(w, ac.Values)
	}
	c.writeTo(w, ac.RParenToken)
	return c.finish()
}

func (tc *TableArg) WriteTo(w io.Writer) (n int64, err error) {
	return tc.Value.WriteTo(w)
}

func (sc *StringArg) WriteTo(w io.Writer) (n int64, err error) {
	return sc.Value.WriteTo(w)
}

func (b *Block) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, stmt := range b.Items {
		if !c.writeTo(w, stmt) {
			break
		}
		if i < len(b.Seps) && b.Seps[i].Type.IsValid() {
			if !c.writeTo(w, b.Seps[i]) {
				break
			}
		}
	}
	return c.finish()
}

func (s *DoStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.DoToken)
	c.writeTo(w, &s.Body)
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (s *AssignStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, &s.Left)
	c.writeTo(w, s.AssignToken)
	c.writeTo(w, &s.Right)
	return c.finish()
}

func (s *CallExprStmt) WriteTo(w io.Writer) (n int64, err error) {
	return s.Call.WriteTo(w)
}

func (s *IfStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.IfToken)
	c.writeTo(w, s.Cond)
	c.writeTo(w, s.ThenToken)
	c.writeTo(w, &s.Body)
	for _, elif := range s.ElseIf {
		if !c.writeTo(w, &elif) {
			break
		}
	}
	if s.Else != nil {
		c.writeTo(w, s.Else)
	}
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (cl *ElseIfClause) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, cl.ElseIfToken)
	c.writeTo(w, cl.Cond)
	c.writeTo(w, cl.ThenToken)
	c.writeTo(w, &cl.Body)
	return c.finish()
}

func (cl *ElseClause) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, cl.ElseToken)
	c.writeTo(w, &cl.Body)
	return c.finish()
}

func (s *NumericForStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.ForToken)
	c.writeTo(w, s.Name)
	c.writeTo(w, s.AssignToken)
	c.writeTo(w, s.Min)
	c.writeTo(w, s.MaxSepToken)
	c.writeTo(w, s.Max)
	if s.StepSepToken.Type.IsValid() {
		c.writeTo(w, s.StepSepToken)
		c.writeTo(w, s.Step)
	}
	c.writeTo(w, s.DoToken)
	c.writeTo(w, &s.Body)
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (s *GenericForStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.ForToken)
	c.writeTo(w, &s.Names)
	c.writeTo(w, s.InToken)
	c.writeTo(w, &s.Iterator)
	c.writeTo(w, s.DoToken)
	c.writeTo(w, &s.Body)
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (s *WhileStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.WhileToken)
	c.writeTo(w, s.Cond)
	c.writeTo(w, s.DoToken)
	c.writeTo(w, &s.Body)
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (s *RepeatStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.RepeatToken)
	c.writeTo(w, &s.Body)
	c.writeTo(w, s.UntilToken)
	c.writeTo(w, s.Cond)
	return c.finish()
}

func (s *LocalVarStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.LocalToken)
	c.writeTo(w, &s.Names)
	if s.AssignToken.Type.IsValid() {
		c.writeTo(w, s.AssignToken)
		c.writeTo(w, s.Values)
	}
	return c.finish()
}

func (s *LocalFunctionStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.LocalToken)
	c.writeTo(w, s.Func.FuncToken)
	c.writeTo(w, s.Name)
	c.writeTo(w, s.Func.LParenToken)
	if s.Func.Params != nil {
		c.writeTo(w, s.Func.Params)
		if s.Func.VarArgSepToken.Type.IsValid() {
			c.writeTo(w, s.Func.VarArgSepToken)
			c.writeTo(w, s.Func.VarArgToken)
		}
	} else if s.Func.VarArgToken.Type.IsValid() {
		c.writeTo(w, s.Func.VarArgToken)
	}
	c.writeTo(w, s.Func.RParenToken)
	c.writeTo(w, &s.Func.Body)
	c.writeTo(w, s.Func.EndToken)
	return c.finish()
}

func (s *FunctionStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.Func.FuncToken)
	c.writeTo(w, &s.Name)
	c.writeTo(w, s.Func.LParenToken)
	if s.Func.Params != nil {
		c.writeTo(w, s.Func.Params)
		if s.Func.VarArgSepToken.Type.IsValid() {
			c.writeTo(w, s.Func.VarArgSepToken)
			c.writeTo(w, s.Func.VarArgToken)
		}
	} else if s.Func.VarArgToken.Type.IsValid() {
		c.writeTo(w, s.Func.VarArgToken)
	}
	c.writeTo(w, s.Func.RParenToken)
	c.writeTo(w, &s.Func.Body)
	c.writeTo(w, s.Func.EndToken)
	return c.finish()
}

func (l *FuncNameList) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, name := range l.Items {
		if !c.writeTo(w, name) {
			break
		}
		if i < len(l.Seps) && l.Seps[i].Type.IsValid() {
			if !c.writeTo(w, l.Seps[i]) {
				break
			}
		}
	}
	return c.finish()
}

func (s *BreakStmt) WriteTo(w io.Writer) (n int64, err error) {
	return s.BreakToken.WriteTo(w)
}

func (s *ReturnStmt) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.ReturnToken)
	if s.Values != nil {
		c.writeTo(w, s.Values)
	}
	return c.finish()
}
