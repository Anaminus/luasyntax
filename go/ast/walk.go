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
