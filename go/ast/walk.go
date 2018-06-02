package ast

import (
	"fmt"
)

// A Visitor's Visit method is called for each node encountered by Walk. If
// the result visitor w is not nil, Walk visits each child of the node with w,
// followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Walk traverses an AST in depth-first, lexical order. It starts by calling
// v.Visit(node); node must not be nil. If the returned visitor w is not nil,
// then Walk is called recursively with w for each non-nil child of the node,
// followed by a call of w.Visit(nil).
//
// Note that a node will be traversed even if it is not valid.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch node := node.(type) {
	case *File:
		Walk(v, &node.Body)

	case *Block:
		for _, item := range node.Items {
			if item != nil {
				Walk(v, item)
			}
		}

	case *ExprList:
		for _, item := range node.Items {
			if item != nil {
				Walk(v, item)
			}
		}

	case *NameList:

	case *NumberExpr:

	case *StringExpr:

	case *NilExpr:

	case *BoolExpr:

	case *VarArgExpr:

	case *UnopExpr:
		if node.Operand != nil {
			Walk(v, node.Operand)
		}

	case *BinopExpr:
		if node.Left != nil {
			Walk(v, node.Left)
		}
		if node.Right != nil {
			Walk(v, node.Right)
		}

	case *ParenExpr:
		if node.Value != nil {
			Walk(v, node.Value)
		}

	case *VariableExpr:

	case *TableCtor:
		Walk(v, &node.Entries)

	case *EntryList:
		for _, item := range node.Items {
			if item != nil {
				Walk(v, item)
			}
		}

	case *IndexEntry:
		if node.Key != nil {
			Walk(v, node.Key)
		}
		if node.Value != nil {
			Walk(v, node.Value)
		}

	case *FieldEntry:
		if node.Value != nil {
			Walk(v, node.Value)
		}

	case *ValueEntry:
		if node.Value != nil {
			Walk(v, node.Value)
		}

	case *FunctionExpr:
		if node.Params != nil {
			Walk(v, node.Params)
		}
		Walk(v, &node.Body)

	case *FieldExpr:
		if node.Value != nil {
			Walk(v, node.Value)
		}

	case *IndexExpr:
		if node.Value != nil {
			Walk(v, node.Value)
		}
		if node.Index != nil {
			Walk(v, node.Index)
		}

	case *MethodExpr:
		if node.Value != nil {
			Walk(v, node.Value)
		}
		if node.Args != nil {
			Walk(v, node.Args)
		}

	case *CallExpr:
		if node.Value != nil {
			Walk(v, node.Value)
		}
		if node.Args != nil {
			Walk(v, node.Args)
		}

	case *ListArgs:
		if node.Values != nil {
			Walk(v, node.Values)
		}

	case *TableArg:
		Walk(v, &node.Value)

	case *StringArg:
		Walk(v, &node.Value)

	case *DoStmt:
		Walk(v, &node.Body)

	case *AssignStmt:
		Walk(v, &node.Left)
		Walk(v, &node.Right)

	case *CallStmt:
		if node.Call != nil {
			Walk(v, node.Call)
		}

	case *IfStmt:
		if node.Cond != nil {
			Walk(v, node.Cond)
		}
		Walk(v, &node.Body)
		for _, elif := range node.ElseIf {
			Walk(v, &elif)
		}
		if node.Else != nil {
			Walk(v, node.Else)
		}

	case *ElseIfClause:
		if node.Cond != nil {
			Walk(v, node.Cond)
		}
		Walk(v, &node.Body)

	case *ElseClause:
		Walk(v, &node.Body)

	case *NumericForStmt:
		if node.Min != nil {
			Walk(v, node.Min)
		}
		if node.Max != nil {
			Walk(v, node.Max)
		}
		if node.Step != nil {
			Walk(v, node.Step)
		}
		Walk(v, &node.Body)

	case *GenericForStmt:
		Walk(v, &node.Names)
		Walk(v, &node.Iterator)
		Walk(v, &node.Body)

	case *WhileStmt:
		if node.Cond != nil {
			Walk(v, node.Cond)
		}
		Walk(v, &node.Body)

	case *RepeatStmt:
		Walk(v, &node.Body)
		if node.Cond != nil {
			Walk(v, node.Cond)
		}

	case *LocalVarStmt:
		Walk(v, &node.Names)
		if node.Values != nil {
			Walk(v, node.Values)
		}

	case *LocalFunctionStmt:
		Walk(v, &node.Func)

	case *FunctionStmt:
		Walk(v, &node.Name)
		Walk(v, &node.Func)

	case *FuncNameList:

	case *BreakStmt:

	case *ReturnStmt:
		if node.Values != nil {
			Walk(v, node.Values)
		}

	default:
		panic(fmt.Sprintf("unexpected node type %T", node))
	}
}

// A TokenVisitor's Visit method is called for each node encountered by
// WalkTokens. If the result visitor w is not nil, WalkTokens visits each
// child of the node with w, followed by a call of w.Visit(nil). Each token
// encountered along the way is called with w.VisitToken(node, n, token),
// where n indicates the nth token of the current node.
type TokenVisitor interface {
	Visit(node Node) (w TokenVisitor)
	VisitToken(node Node, n int, tok *Token)
}

// WalkTokens traverses an AST, visiting each token in depth-first, lexical
// order. It starts by calling v.Visit(node); node must not be nil. If the
// returned visitor w is not nil, then WalkTokens is called recursively with w
// for each non-nil child of the node, followed by a call of w.Visit(nil).
// Each token encountered along the way is called with w.VisitToken(node, n,
// token), where n indicates the nth token of the current node. Tokens are
// guaranteed to be visited in lexical order.
//
// Note that a node or token will be visited even if it is not valid.
func WalkTokens(v TokenVisitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch node := node.(type) {
	case *File:
		WalkTokens(v, &node.Body)
		v.VisitToken(node, 0, &node.EOFToken)

	case *Block:
		for i, item := range node.Items {
			if item != nil {
				WalkTokens(v, item)
			}
			if i < len(node.Seps) {
				v.VisitToken(node, i, &node.Seps[i])
			}
		}

	case *ExprList:
		for i, item := range node.Items {
			if item != nil {
				WalkTokens(v, item)
			}
			if i < len(node.Seps) {
				v.VisitToken(node, i, &node.Seps[i])
			}
		}

	case *NameList:
		n := 0
		for i := range node.Items {
			v.VisitToken(node, n, &node.Items[i])
			n++
			if i < len(node.Seps) {
				v.VisitToken(node, n, &node.Seps[i])
				n++
			}
		}

	case *NumberExpr:
		v.VisitToken(node, 0, &node.NumberToken)

	case *StringExpr:
		v.VisitToken(node, 0, &node.StringToken)

	case *NilExpr:
		v.VisitToken(node, 0, &node.NilToken)

	case *BoolExpr:
		v.VisitToken(node, 0, &node.BoolToken)

	case *VarArgExpr:
		v.VisitToken(node, 0, &node.VarArgToken)

	case *UnopExpr:
		v.VisitToken(node, 0, &node.UnopToken)
		if node.Operand != nil {
			WalkTokens(v, node.Operand)
		}

	case *BinopExpr:
		if node.Left != nil {
			WalkTokens(v, node.Left)
		}
		v.VisitToken(node, 0, &node.BinopToken)
		if node.Right != nil {
			WalkTokens(v, node.Right)
		}

	case *ParenExpr:
		v.VisitToken(node, 0, &node.LParenToken)
		if node.Value != nil {
			WalkTokens(v, node.Value)
		}
		v.VisitToken(node, 1, &node.RParenToken)

	case *VariableExpr:
		v.VisitToken(node, 0, &node.NameToken)

	case *TableCtor:
		v.VisitToken(node, 0, &node.LBraceToken)
		WalkTokens(v, &node.Entries)
		v.VisitToken(node, 1, &node.RBraceToken)

	case *EntryList:
		for i, item := range node.Items {
			if item != nil {
				WalkTokens(v, item)
			}
			if i < len(node.Seps) {
				v.VisitToken(node, i, &node.Seps[i])
			}
		}

	case *IndexEntry:
		v.VisitToken(node, 0, &node.LBrackToken)
		if node.Key != nil {
			WalkTokens(v, node.Key)
		}
		v.VisitToken(node, 1, &node.RBrackToken)
		v.VisitToken(node, 2, &node.AssignToken)
		if node.Value != nil {
			WalkTokens(v, node.Value)
		}

	case *FieldEntry:
		v.VisitToken(node, 0, &node.NameToken)
		v.VisitToken(node, 1, &node.AssignToken)
		if node.Value != nil {
			WalkTokens(v, node.Value)
		}

	case *ValueEntry:
		if node.Value != nil {
			WalkTokens(v, node.Value)
		}

	case *FunctionExpr:
		v.VisitToken(node, 0, &node.FuncToken)
		v.VisitToken(node, 1, &node.LParenToken)
		if node.Params != nil {
			WalkTokens(v, node.Params)
		}
		v.VisitToken(node, 2, &node.VarArgSepToken)
		v.VisitToken(node, 3, &node.VarArgToken)
		v.VisitToken(node, 4, &node.RParenToken)
		WalkTokens(v, &node.Body)
		v.VisitToken(node, 5, &node.EndToken)

	case *FieldExpr:
		if node.Value != nil {
			WalkTokens(v, node.Value)
		}
		v.VisitToken(node, 0, &node.DotToken)
		v.VisitToken(node, 1, &node.NameToken)

	case *IndexExpr:
		if node.Value != nil {
			WalkTokens(v, node.Value)
		}
		v.VisitToken(node, 0, &node.LBrackToken)
		if node.Index != nil {
			WalkTokens(v, node.Index)
		}
		v.VisitToken(node, 1, &node.RBrackToken)

	case *MethodExpr:
		if node.Value != nil {
			WalkTokens(v, node.Value)
		}
		v.VisitToken(node, 0, &node.ColonToken)
		v.VisitToken(node, 1, &node.NameToken)
		if node.Args != nil {
			WalkTokens(v, node.Args)
		}

	case *CallExpr:
		if node.Value != nil {
			WalkTokens(v, node.Value)
		}
		if node.Args != nil {
			WalkTokens(v, node.Args)
		}

	case *ListArgs:
		v.VisitToken(node, 0, &node.LParenToken)
		if node.Values != nil {
			WalkTokens(v, node.Values)
		}
		v.VisitToken(node, 1, &node.RParenToken)

	case *TableArg:
		WalkTokens(v, &node.Value)

	case *StringArg:
		WalkTokens(v, &node.Value)

	case *DoStmt:
		v.VisitToken(node, 0, &node.DoToken)
		WalkTokens(v, &node.Body)
		v.VisitToken(node, 1, &node.EndToken)

	case *AssignStmt:
		WalkTokens(v, &node.Left)
		v.VisitToken(node, 0, &node.AssignToken)
		WalkTokens(v, &node.Right)

	case *CallStmt:
		if node.Call != nil {
			WalkTokens(v, node.Call)
		}

	case *IfStmt:
		v.VisitToken(node, 0, &node.IfToken)
		if node.Cond != nil {
			WalkTokens(v, node.Cond)
		}
		v.VisitToken(node, 1, &node.ThenToken)
		WalkTokens(v, &node.Body)
		for _, elif := range node.ElseIf {
			WalkTokens(v, &elif)
		}
		if node.Else != nil {
			WalkTokens(v, node.Else)
		}

	case *ElseIfClause:
		v.VisitToken(node, 0, &node.ElseIfToken)
		if node.Cond != nil {
			WalkTokens(v, node.Cond)
		}
		v.VisitToken(node, 1, &node.ThenToken)
		WalkTokens(v, &node.Body)

	case *ElseClause:
		v.VisitToken(node, 0, &node.ElseToken)
		WalkTokens(v, &node.Body)

	case *NumericForStmt:
		v.VisitToken(node, 0, &node.ForToken)
		v.VisitToken(node, 1, &node.NameToken)
		v.VisitToken(node, 2, &node.AssignToken)
		if node.Min != nil {
			WalkTokens(v, node.Min)
		}
		v.VisitToken(node, 3, &node.MaxSepToken)
		if node.Max != nil {
			WalkTokens(v, node.Max)
		}
		v.VisitToken(node, 4, &node.StepSepToken)
		if node.Step != nil {
			WalkTokens(v, node.Step)
		}
		v.VisitToken(node, 5, &node.DoToken)
		WalkTokens(v, &node.Body)
		v.VisitToken(node, 6, &node.EndToken)

	case *GenericForStmt:
		v.VisitToken(node, 0, &node.ForToken)
		WalkTokens(v, &node.Names)
		v.VisitToken(node, 1, &node.InToken)
		WalkTokens(v, &node.Iterator)
		v.VisitToken(node, 2, &node.DoToken)
		WalkTokens(v, &node.Body)
		v.VisitToken(node, 3, &node.EndToken)

	case *WhileStmt:
		v.VisitToken(node, 0, &node.WhileToken)
		if node.Cond != nil {
			WalkTokens(v, node.Cond)
		}
		v.VisitToken(node, 1, &node.DoToken)
		WalkTokens(v, &node.Body)
		v.VisitToken(node, 2, &node.EndToken)

	case *RepeatStmt:
		v.VisitToken(node, 0, &node.RepeatToken)
		WalkTokens(v, &node.Body)
		v.VisitToken(node, 1, &node.UntilToken)
		if node.Cond != nil {
			WalkTokens(v, node.Cond)
		}

	case *LocalVarStmt:
		v.VisitToken(node, 0, &node.LocalToken)
		WalkTokens(v, &node.Names)
		v.VisitToken(node, 1, &node.AssignToken)
		if node.Values != nil {
			WalkTokens(v, node.Values)
		}

	case *LocalFunctionStmt:
		v.VisitToken(node, 0, &node.LocalToken)
		v.VisitToken(node, 1, &node.Func.FuncToken)
		v.VisitToken(node, 2, &node.NameToken)
		v.VisitToken(node, 3, &node.Func.LParenToken)
		if node.Func.Params != nil {
			WalkTokens(v, node.Func.Params)
		}
		v.VisitToken(node, 4, &node.Func.VarArgSepToken)
		v.VisitToken(node, 5, &node.Func.VarArgToken)
		v.VisitToken(node, 6, &node.Func.RParenToken)
		WalkTokens(v, &node.Func.Body)
		v.VisitToken(node, 7, &node.Func.EndToken)

	case *FunctionStmt:
		v.VisitToken(node, 0, &node.Func.FuncToken)
		WalkTokens(v, &node.Name)
		v.VisitToken(node, 2, &node.Func.LParenToken)
		if node.Func.Params != nil {
			WalkTokens(v, node.Func.Params)
		}
		v.VisitToken(node, 3, &node.Func.VarArgSepToken)
		v.VisitToken(node, 4, &node.Func.VarArgToken)
		v.VisitToken(node, 5, &node.Func.RParenToken)
		WalkTokens(v, &node.Func.Body)
		v.VisitToken(node, 6, &node.Func.EndToken)

	case *FuncNameList:
		n := 0
		for i := range node.Items {
			v.VisitToken(node, n, &node.Items[i])
			n++
			if i < len(node.Seps) {
				v.VisitToken(node, n, &node.Seps[i])
				n++
			}
		}
		v.VisitToken(node, n+1, &node.ColonToken)
		v.VisitToken(node, n+2, &node.MethodToken)

	case *BreakStmt:
		v.VisitToken(node, 0, &node.BreakToken)

	case *ReturnStmt:
		v.VisitToken(node, 0, &node.ReturnToken)
		if node.Values != nil {
			WalkTokens(v, node.Values)
		}

	default:
		panic(fmt.Sprintf("unexpected node type %T", node))
	}
}
