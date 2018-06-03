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

// A TokenVisitor extends a Visitor. If a Visitor is a TokenVisitor, then Walk
// will call w.VisitToken(node, n, token) for each token encountered.
//
// The n argument indicates nth token of the current node. Combined with the
// node type, this can be used to identify the token field being referred to.
// For example, (DoStmt, 0) refers to DoStmt.DoToken, which is a DO token, and
// (DoStmt, 1) refers to DoStmt.EndToken, which is an END token. For nodes
// like NameList, an even value indicates a NAME token in the Items field,
// while an odd value indicates a COMMA token in the Seps field.
//
// Tokens are guaranteed to be visited in lexical order.
type TokenVisitor interface {
	Visitor
	VisitToken(node Node, n int, tok *Token)
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
	tv, tvok := v.(TokenVisitor)

	switch node := node.(type) {
	case *File:
		Walk(v, &node.Body)
		if tvok {
			tv.VisitToken(node, 0, &node.EOFToken)
		}

	case *Block:
		if tvok {
			for i, item := range node.Items {
				if item != nil {
					Walk(v, item)
				}
				if i < len(node.Seps) {
					tv.VisitToken(node, i, &node.Seps[i])
				}
			}
		} else {
			for _, item := range node.Items {
				if item != nil {
					Walk(v, item)
				}
			}
		}

	case *ExprList:
		if tvok {
			for i, item := range node.Items {
				if item != nil {
					Walk(v, item)
				}
				if i < len(node.Seps) {
					tv.VisitToken(node, i, &node.Seps[i])
				}
			}
		} else {
			for _, item := range node.Items {
				if item != nil {
					Walk(v, item)
				}
			}
		}

	case *NameList:
		if tvok {
			n := 0
			for i := range node.Items {
				tv.VisitToken(node, n, &node.Items[i])
				n++
				if i < len(node.Seps) {
					tv.VisitToken(node, n, &node.Seps[i])
					n++
				}
			}
		}

	case *NumberExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.NumberToken)
		}

	case *StringExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.StringToken)
		}

	case *NilExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.NilToken)
		}

	case *BoolExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.BoolToken)
		}

	case *VarArgExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.VarArgToken)
		}

	case *UnopExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.UnopToken)
		}
		if node.Operand != nil {
			Walk(v, node.Operand)
		}

	case *BinopExpr:
		if node.Left != nil {
			Walk(v, node.Left)
		}
		if tvok {
			tv.VisitToken(node, 0, &node.BinopToken)
		}
		if node.Right != nil {
			Walk(v, node.Right)
		}

	case *ParenExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.LParenToken)
		}
		if node.Value != nil {
			Walk(v, node.Value)
		}
		if tvok {
			tv.VisitToken(node, 1, &node.RParenToken)
		}

	case *VariableExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.NameToken)
		}

	case *TableCtor:
		if tvok {
			tv.VisitToken(node, 0, &node.LBraceToken)
		}
		Walk(v, &node.Entries)
		if tvok {
			tv.VisitToken(node, 1, &node.RBraceToken)
		}

	case *EntryList:
		if tvok {
			for i, item := range node.Items {
				if item != nil {
					Walk(v, item)
				}
				if i < len(node.Seps) {
					tv.VisitToken(node, i, &node.Seps[i])
				}
			}
		} else {
			for _, item := range node.Items {
				if item != nil {
					Walk(v, item)
				}
			}
		}

	case *IndexEntry:
		if tvok {
			tv.VisitToken(node, 0, &node.LBrackToken)
		}
		if node.Key != nil {
			Walk(v, node.Key)
		}
		if tvok {
			tv.VisitToken(node, 1, &node.RBrackToken)
			tv.VisitToken(node, 2, &node.AssignToken)
		}
		if node.Value != nil {
			Walk(v, node.Value)
		}

	case *FieldEntry:
		if tvok {
			tv.VisitToken(node, 0, &node.NameToken)
			tv.VisitToken(node, 1, &node.AssignToken)
		}
		if node.Value != nil {
			Walk(v, node.Value)
		}

	case *ValueEntry:
		if node.Value != nil {
			Walk(v, node.Value)
		}

	case *FunctionExpr:
		if tvok {
			tv.VisitToken(node, 0, &node.FuncToken)
			tv.VisitToken(node, 1, &node.LParenToken)
		}
		if node.Params != nil {
			Walk(v, node.Params)
		}
		if tvok {
			tv.VisitToken(node, 2, &node.VarArgSepToken)
			tv.VisitToken(node, 3, &node.VarArgToken)
			tv.VisitToken(node, 4, &node.RParenToken)
		}
		Walk(v, &node.Body)
		if tvok {
			tv.VisitToken(node, 5, &node.EndToken)
		}

	case *FieldExpr:
		if node.Value != nil {
			Walk(v, node.Value)
		}
		if tvok {
			tv.VisitToken(node, 0, &node.DotToken)
			tv.VisitToken(node, 1, &node.NameToken)
		}

	case *IndexExpr:
		if node.Value != nil {
			Walk(v, node.Value)
		}
		if tvok {
			tv.VisitToken(node, 0, &node.LBrackToken)
		}
		if node.Index != nil {
			Walk(v, node.Index)
		}
		if tvok {
			tv.VisitToken(node, 1, &node.RBrackToken)
		}

	case *MethodExpr:
		if node.Value != nil {
			Walk(v, node.Value)
		}
		if tvok {
			tv.VisitToken(node, 0, &node.ColonToken)
			tv.VisitToken(node, 1, &node.NameToken)
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
		if tvok {
			tv.VisitToken(node, 0, &node.LParenToken)
		}
		if node.Values != nil {
			Walk(v, node.Values)
		}
		if tvok {
			tv.VisitToken(node, 1, &node.RParenToken)
		}

	case *TableArg:
		Walk(v, &node.Value)

	case *StringArg:
		Walk(v, &node.Value)

	case *DoStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.DoToken)
		}
		Walk(v, &node.Body)
		if tvok {
			tv.VisitToken(node, 1, &node.EndToken)
		}

	case *AssignStmt:
		Walk(v, &node.Left)
		if tvok {
			tv.VisitToken(node, 0, &node.AssignToken)
		}
		Walk(v, &node.Right)

	case *CallStmt:
		if node.Call != nil {
			Walk(v, node.Call)
		}

	case *IfStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.IfToken)
		}
		if node.Cond != nil {
			Walk(v, node.Cond)
		}
		if tvok {
			tv.VisitToken(node, 1, &node.ThenToken)
		}
		Walk(v, &node.Body)
		for _, elif := range node.ElseIf {
			Walk(v, &elif)
		}
		if node.Else != nil {
			Walk(v, node.Else)
		}
		if tvok {
			tv.VisitToken(node, 1, &node.EndToken)
		}

	case *ElseIfClause:
		if tvok {
			tv.VisitToken(node, 0, &node.ElseIfToken)
		}
		if node.Cond != nil {
			Walk(v, node.Cond)
		}
		if tvok {
			tv.VisitToken(node, 1, &node.ThenToken)
		}
		Walk(v, &node.Body)

	case *ElseClause:
		if tvok {
			tv.VisitToken(node, 0, &node.ElseToken)
		}
		Walk(v, &node.Body)

	case *NumericForStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.ForToken)
			tv.VisitToken(node, 1, &node.NameToken)
			tv.VisitToken(node, 2, &node.AssignToken)
		}
		if node.Min != nil {
			Walk(v, node.Min)
		}
		if tvok {
			tv.VisitToken(node, 3, &node.MaxSepToken)
		}
		if node.Max != nil {
			Walk(v, node.Max)
		}
		if tvok {
			tv.VisitToken(node, 4, &node.StepSepToken)
		}
		if node.Step != nil {
			Walk(v, node.Step)
		}
		if tvok {
			tv.VisitToken(node, 5, &node.DoToken)
		}
		Walk(v, &node.Body)
		if tvok {
			tv.VisitToken(node, 6, &node.EndToken)
		}

	case *GenericForStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.ForToken)
		}
		Walk(v, &node.Names)
		if tvok {
			tv.VisitToken(node, 1, &node.InToken)
		}
		Walk(v, &node.Iterator)
		if tvok {
			tv.VisitToken(node, 2, &node.DoToken)
		}
		Walk(v, &node.Body)
		if tvok {
			tv.VisitToken(node, 3, &node.EndToken)
		}

	case *WhileStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.WhileToken)
		}
		if node.Cond != nil {
			Walk(v, node.Cond)
		}
		if tvok {
			tv.VisitToken(node, 1, &node.DoToken)
		}
		Walk(v, &node.Body)
		if tvok {
			tv.VisitToken(node, 2, &node.EndToken)
		}

	case *RepeatStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.RepeatToken)
		}
		Walk(v, &node.Body)
		if tvok {
			tv.VisitToken(node, 1, &node.UntilToken)
		}
		if node.Cond != nil {
			Walk(v, node.Cond)
		}

	case *LocalVarStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.LocalToken)
		}
		Walk(v, &node.Names)
		if tvok {
			tv.VisitToken(node, 1, &node.AssignToken)
		}
		if node.Values != nil {
			Walk(v, node.Values)
		}

	case *LocalFunctionStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.LocalToken)
			tv.VisitToken(node, 1, &node.Func.FuncToken)
			tv.VisitToken(node, 2, &node.NameToken)
			tv.VisitToken(node, 3, &node.Func.LParenToken)
		}
		if node.Func.Params != nil {
			Walk(v, node.Func.Params)
		}
		if tvok {
			tv.VisitToken(node, 4, &node.Func.VarArgSepToken)
			tv.VisitToken(node, 5, &node.Func.VarArgToken)
			tv.VisitToken(node, 6, &node.Func.RParenToken)
		}
		Walk(v, &node.Func.Body)
		if tvok {
			tv.VisitToken(node, 7, &node.Func.EndToken)
		}

	case *FunctionStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.Func.FuncToken)
		}
		Walk(v, &node.Name)
		if tvok {
			tv.VisitToken(node, 2, &node.Func.LParenToken)
		}
		if node.Func.Params != nil {
			Walk(v, node.Func.Params)
		}
		if tvok {
			tv.VisitToken(node, 3, &node.Func.VarArgSepToken)
			tv.VisitToken(node, 4, &node.Func.VarArgToken)
			tv.VisitToken(node, 5, &node.Func.RParenToken)
		}
		Walk(v, &node.Func.Body)
		if tvok {
			tv.VisitToken(node, 6, &node.Func.EndToken)
		}

	case *FuncNameList:
		if tvok {
			n := 0
			for i := range node.Items {
				tv.VisitToken(node, n, &node.Items[i])
				n++
				if i < len(node.Seps) {
					tv.VisitToken(node, n, &node.Seps[i])
					n++
				}
			}
			tv.VisitToken(node, n+1, &node.ColonToken)
			tv.VisitToken(node, n+2, &node.MethodToken)
		}

	case *BreakStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.BreakToken)
		}

	case *ReturnStmt:
		if tvok {
			tv.VisitToken(node, 0, &node.ReturnToken)
		}
		if node.Values != nil {
			Walk(v, node.Values)
		}

	default:
		panic(fmt.Sprintf("unexpected node type %T", node))
	}
}
