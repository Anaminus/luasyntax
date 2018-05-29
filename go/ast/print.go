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
	c.writeTo(w, &f.Block)
	c.writeTo(w, f.EOFToken)
	return c.finish()
}

func (l *ExpList) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, exp := range l.Exps {
		if !c.writeTo(w, exp) {
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
	for i, name := range l.Names {
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

func (e *UnopExp) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.UnopToken)
	c.writeTo(w, e.Exp)
	return c.finish()
}

func (e *BinopExp) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Left)
	c.writeTo(w, e.BinopToken)
	c.writeTo(w, e.Right)
	return c.finish()
}

func (e *ParenExp) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.LParenToken)
	c.writeTo(w, e.Exp)
	c.writeTo(w, e.RParenToken)
	return c.finish()
}

func (e *VariableExp) WriteTo(w io.Writer) (n int64, err error) {
	return e.NameToken.WriteTo(w)
}

func (e *TableCtor) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.LBraceToken)
	c.writeTo(w, &e.EntryList)
	c.writeTo(w, e.RBraceToken)
	return c.finish()
}

func (l *EntryList) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, entry := range l.Entries {
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
	c.writeTo(w, e.KeyExp)
	c.writeTo(w, e.RBrackToken)
	c.writeTo(w, e.AssignToken)
	c.writeTo(w, e.ValueExp)
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

func (e *FunctionExp) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.FuncToken)
	c.writeTo(w, e.LParenToken)
	if e.ParList != nil {
		c.writeTo(w, e.ParList)
		if e.VarArgSepToken.Type.IsValid() {
			c.writeTo(w, e.VarArgSepToken)
			c.writeTo(w, e.VarArgToken)
		}
	} else if e.VarArgToken.Type.IsValid() {
		c.writeTo(w, e.VarArgToken)
	}
	c.writeTo(w, e.RParenToken)
	c.writeTo(w, &e.Block)
	c.writeTo(w, e.EndToken)
	return c.finish()
}

func (e *FieldExp) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Exp)
	c.writeTo(w, e.DotToken)
	c.writeTo(w, e.Field)
	return c.finish()
}

func (e *IndexExp) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Exp)
	c.writeTo(w, e.LBrackToken)
	c.writeTo(w, e.Index)
	c.writeTo(w, e.RBrackToken)
	return c.finish()
}

func (e *MethodExp) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Exp)
	c.writeTo(w, e.ColonToken)
	c.writeTo(w, e.Name)
	c.writeTo(w, e.Args)
	return c.finish()
}

func (e *CallExp) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, e.Exp)
	c.writeTo(w, e.Args)
	return c.finish()
}

func (ac *ArgsCall) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, ac.LParenToken)
	if ac.ExpList != nil {
		c.writeTo(w, ac.ExpList)
	}
	c.writeTo(w, ac.RParenToken)
	return c.finish()
}

func (tc *TableCall) WriteTo(w io.Writer) (n int64, err error) {
	return tc.TableExp.WriteTo(w)
}

func (sc *StringCall) WriteTo(w io.Writer) (n int64, err error) {
	return sc.StringExp.WriteTo(w)
}

func (b *Block) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, stat := range b.Stats {
		if !c.writeTo(w, stat) {
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

func (s *DoStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.DoToken)
	c.writeTo(w, &s.Block)
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (s *AssignStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, &s.Left)
	c.writeTo(w, s.AssignToken)
	c.writeTo(w, &s.Right)
	return c.finish()
}

func (s *CallExprStat) WriteTo(w io.Writer) (n int64, err error) {
	return s.Exp.WriteTo(w)
}

func (s *IfStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.IfToken)
	c.writeTo(w, s.Exp)
	c.writeTo(w, s.ThenToken)
	c.writeTo(w, &s.Block)
	for _, elif := range s.ElseIfClauses {
		if !c.writeTo(w, &elif) {
			break
		}
	}
	if s.ElseClause != nil {
		c.writeTo(w, s.ElseClause)
	}
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (cl *ElseIfClause) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, cl.ElseIfToken)
	c.writeTo(w, cl.Exp)
	c.writeTo(w, cl.ThenToken)
	c.writeTo(w, &cl.Block)
	return c.finish()
}

func (cl *ElseClause) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, cl.ElseToken)
	c.writeTo(w, &cl.Block)
	return c.finish()
}

func (s *NumericForStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.ForToken)
	c.writeTo(w, s.Name)
	c.writeTo(w, s.AssignToken)
	c.writeTo(w, s.MinExp)
	c.writeTo(w, s.MaxSepToken)
	c.writeTo(w, s.MaxExp)
	if s.StepSepToken.Type.IsValid() {
		c.writeTo(w, s.StepSepToken)
		c.writeTo(w, s.StepExp)
	}
	c.writeTo(w, s.DoToken)
	c.writeTo(w, &s.Block)
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (s *GenericForStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.ForToken)
	c.writeTo(w, &s.NameList)
	c.writeTo(w, s.InToken)
	c.writeTo(w, &s.ExpList)
	c.writeTo(w, s.DoToken)
	c.writeTo(w, &s.Block)
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (s *WhileStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.WhileToken)
	c.writeTo(w, s.Exp)
	c.writeTo(w, s.DoToken)
	c.writeTo(w, &s.Block)
	c.writeTo(w, s.EndToken)
	return c.finish()
}

func (s *RepeatStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.RepeatToken)
	c.writeTo(w, &s.Block)
	c.writeTo(w, s.UntilToken)
	c.writeTo(w, s.Exp)
	return c.finish()
}

func (s *LocalVarStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.LocalToken)
	c.writeTo(w, &s.NameList)
	if s.AssignToken.Type.IsValid() {
		c.writeTo(w, s.AssignToken)
		c.writeTo(w, s.ExpList)
	}
	return c.finish()
}

func (s *LocalFunctionStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.LocalToken)
	c.writeTo(w, s.Exp.FuncToken)
	c.writeTo(w, s.Name)
	c.writeTo(w, s.Exp.LParenToken)
	if s.Exp.ParList != nil {
		c.writeTo(w, s.Exp.ParList)
		if s.Exp.VarArgSepToken.Type.IsValid() {
			c.writeTo(w, s.Exp.VarArgSepToken)
			c.writeTo(w, s.Exp.VarArgToken)
		}
	} else if s.Exp.VarArgToken.Type.IsValid() {
		c.writeTo(w, s.Exp.VarArgToken)
	}
	c.writeTo(w, s.Exp.RParenToken)
	c.writeTo(w, &s.Exp.Block)
	c.writeTo(w, s.Exp.EndToken)
	return c.finish()
}

func (s *FunctionStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.Exp.FuncToken)
	c.writeTo(w, &s.Name)
	c.writeTo(w, s.Exp.LParenToken)
	if s.Exp.ParList != nil {
		c.writeTo(w, s.Exp.ParList)
		if s.Exp.VarArgSepToken.Type.IsValid() {
			c.writeTo(w, s.Exp.VarArgSepToken)
			c.writeTo(w, s.Exp.VarArgToken)
		}
	} else if s.Exp.VarArgToken.Type.IsValid() {
		c.writeTo(w, s.Exp.VarArgToken)
	}
	c.writeTo(w, s.Exp.RParenToken)
	c.writeTo(w, &s.Exp.Block)
	c.writeTo(w, s.Exp.EndToken)
	return c.finish()
}

func (l *FuncNameList) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	for i, name := range l.Names {
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

func (s *BreakStat) WriteTo(w io.Writer) (n int64, err error) {
	return s.BreakToken.WriteTo(w)
}

func (s *ReturnStat) WriteTo(w io.Writer) (n int64, err error) {
	var c copier
	c.writeTo(w, s.ReturnToken)
	if s.ExpList != nil {
		c.writeTo(w, s.ExpList)
	}
	return c.finish()
}
