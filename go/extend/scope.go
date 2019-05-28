package extend

import (
	"github.com/anaminus/luasyntax/go/tree"
)

// FileScope contains information about the scopes of a file, including
// variables and their associations with a parse tree.
type FileScope struct {
	// Root is the root scope.
	Root *Scope
	// Globals is a list of global variables that have been assigned to.
	Globals []*Variable
	// VariableMap maps a NAME token to a Variable.
	VariableMap map[*tree.Token]*Variable
	// ScopeMap maps a Node to the scope that is opened by or is otherwise
	// associated with the node.
	ScopeMap map[tree.Node]*Scope
}

// Scope contains a list of the variables declared in the scope.
type Scope struct {
	// Parent is the outer, surrounding scope.
	Parent *Scope
	// Children is a list of inner scopes.
	Children []*Scope
	// Variables is the list of variables declared in the scope.
	Variables []*Variable
	// Node is the tree node that opens or is otherwise associated with the
	// scope. May be nil.
	Node tree.Node
}

// NewScope creates an inner scope, optionally associating the scope with the
// node that opens it.
func NewScope(parent *Scope, node tree.Node) *Scope {
	scope := &Scope{Parent: parent, Node: node}
	if parent != nil {
		parent.Children = append(parent.Children, scope)
	}
	return scope
}

// VariableType indicates the type of Variable.
type VariableType uint8

const (
	InvalidVar VariableType = iota

	LocalVar  // LocalVar indicates a variable local to its scope.
	GlobalVar // GlobalVar indicates a variable defined in the global table.
)

func (t VariableType) String() string {
	switch t {
	case LocalVar:
		return "Local"
	case GlobalVar:
		return "Global"
	}
	return "<invalid>"
}

// Variable describes a single named entity within a parse tree.
type Variable struct {
	// Type is the variable type.
	Type VariableType
	// Name is the name of the variable.
	Name string
	// References is a list of NAME tokens that refer to the entity. When the
	// variable is local, the first value is the declaration of the variable.
	References []*tree.Token
}

// scopeParser holds the scope state while walking a parse tree. It must be
// initialized with init before using.
type scopeParser struct {
	fileScope    *FileScope
	currentScope *Scope
}

// init prepares the parser to walk a parse tree.
func (p *scopeParser) init() {
	p.currentScope = nil
	p.fileScope = &FileScope{
		VariableMap: make(map[*tree.Token]*Variable, 4),
		ScopeMap:    make(map[tree.Node]*Scope, 4),
	}
}

// openScope creates a new scope, setting it as an inner scope of the current
// scope, and then sets it as the current scope. The scope can optionally be
// associated with a node.
func (p *scopeParser) openScope(node tree.Node) {
	p.currentScope = NewScope(p.currentScope, node)
	if node != nil {
		p.fileScope.ScopeMap[node] = p.currentScope
	}
}

// closeScope sets the current scope to its parent.
func (p *scopeParser) closeScope() {
	p.currentScope = p.currentScope.Parent
}

// AddLocalVar creates a new Variable, named by the given NAME token, and adds
// it to the current scope.
func (p *scopeParser) addLocalVar(name *tree.Token) {
	v := &Variable{
		Type:       LocalVar,
		Name:       string(name.Bytes),
		References: []*tree.Token{name},
	}
	p.currentScope.Variables = append(p.currentScope.Variables, v)
	p.fileScope.VariableMap[name] = v
}

// getLocalVar retrieves a variable, named by the given NAME token, from the
// current scope, or each outer scope until it is found. Returns nil if no
// variable of the given name could be found.
func (p *scopeParser) getLocalVar(name *tree.Token) *Variable {
	for scope := p.currentScope; scope != nil; scope = scope.Parent {
		// Iterate in reverse order to handle shadowing correctly.
		for i := len(scope.Variables) - 1; i >= 0; i-- {
			if scope.Variables[i].Name == string(name.Bytes) {
				return scope.Variables[i]
			}
		}
	}
	return nil
}

// referenceVariable adds a reference to the variable named by the given NAME
// token. The variable may be local or global.
func (p *scopeParser) referenceVariable(name *tree.Token) *Variable {
	v := p.getLocalVar(name)
	if v != nil {
		v.References = append(v.References, name)
	} else {
		v = p.addGlobalVar(name)
	}
	return v
}

// addGlobalVar adds a reference to a global variable, named by the given NAME
// token. A new Variable is created, if necessary.
func (p *scopeParser) addGlobalVar(name *tree.Token) (v *Variable) {
	for _, g := range p.fileScope.Globals {
		if g.Name == string(name.Bytes) {
			v = g
			break
		}
	}
	if v == nil {
		v = &Variable{
			Type: GlobalVar,
			Name: string(name.Bytes),
		}
		p.fileScope.Globals = append(p.fileScope.Globals, v)
	}
	v.References = append(v.References, name)
	p.fileScope.VariableMap[name] = v
	return v
}

// Visit implements the tree.Visitor interface.
func (p *scopeParser) Visit(node tree.Node) tree.Visitor {
	switch node := node.(type) {
	case *tree.File:
		if p.fileScope.Root != nil {
			panic("only one file can be read!")
		}
		p.openScope(node)
		p.fileScope.Root = p.currentScope
		tree.Walk(p, &node.Body)
		p.closeScope()
		return nil

	case *tree.NameList:
		for i := range node.Items {
			p.addLocalVar(&node.Items[i])
		}
		return nil

	case *tree.VariableExpr:
		p.referenceVariable(&node.NameToken)
		return nil

	case *tree.FunctionExpr:
		// Also handles body of LocalFunctionStmt and FunctionStmt.
		p.openScope(node)
		if node.Params != nil {
			tree.Walk(p, node.Params)
		}
		tree.Walk(p, &node.Body)
		p.closeScope()
		return nil

	case *tree.DoStmt:
		p.openScope(node)
		tree.Walk(p, &node.Body)
		p.closeScope()
		return nil

	case *tree.IfStmt:
		p.openScope(node)
		if node.Cond != nil {
			tree.Walk(p, node.Cond)
		}
		tree.Walk(p, &node.Body)
		for i := range node.ElseIf {
			// Close previous if/elseif scope.
			p.closeScope()
			p.openScope(&node.ElseIf[i])
			tree.Walk(p, &node.ElseIf[i])
		}
		if node.Else != nil {
			// Close previous if/elseif scope.
			p.closeScope()
			p.openScope(node.Else)
			tree.Walk(p, node.Else)
		}
		p.closeScope()
		return nil

	case *tree.NumericForStmt:
		// Open a separate scope for range expressions, which should appear
		// before the scope of the body, but not as a parent.

		// TODO: Figure out a better way to map this scope to a node.
		p.openScope(node.Min)
		if node.Min != nil {
			tree.Walk(p, node.Min)
		}
		if node.Max != nil {
			tree.Walk(p, node.Max)
		}
		if node.Step != nil {
			tree.Walk(p, node.Step)
		}
		p.closeScope()
		p.openScope(node)
		p.addLocalVar(&node.NameToken)
		tree.Walk(p, &node.Body)
		p.closeScope()
		return nil

	case *tree.GenericForStmt:
		// Open a separate scope for iterator expressions, which must appear
		// before the scope of the body, but not as a parent.
		p.openScope(&node.Iterator)
		tree.Walk(p, &node.Iterator)
		p.closeScope()
		p.openScope(node)
		tree.Walk(p, &node.Names)
		tree.Walk(p, &node.Body)
		p.closeScope()
		return nil

	case *tree.WhileStmt:
		p.openScope(node)
		if node.Cond != nil {
			tree.Walk(p, node.Cond)
		}
		tree.Walk(p, &node.Body)
		p.closeScope()
		return nil

	case *tree.RepeatStmt:
		p.openScope(node)
		tree.Walk(p, &node.Body)
		if node.Cond != nil {
			tree.Walk(p, node.Cond)
		}
		p.closeScope()
		return nil

	case *tree.LocalVarStmt:
		// Expressions must be added first.
		if node.Values != nil {
			tree.Walk(p, node.Values)
		}
		// Add variables.
		tree.Walk(p, &node.Names)
		return nil

	case *tree.LocalFunctionStmt:
		// Add local variable assignment.
		p.addLocalVar(&node.NameToken)
		// Body handled by FunctionExpr.

	case *tree.FunctionStmt:
		// Body handled by FunctionExpr.

	case *tree.FuncNameList:
		// Refer to first name in list.
		p.referenceVariable(&node.Items[0])
		return nil

	default:
	}
	return p
}

// BuildFileScope walks the given parse tree, building a tree of scopes and the
// variables they contain.
func BuildFileScope(file *tree.File) *FileScope {
	var p scopeParser
	p.init()
	tree.Walk(&p, file)
	if p.currentScope != nil {
		panic("unbalanced scopes")
	}
	return p.fileScope
}
