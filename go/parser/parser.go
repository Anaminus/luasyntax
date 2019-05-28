// The parser package implements a parser for Lua source files. Input may be
// provided in a variety of forms, and the output is a parse tree representing
// the Lua source.
package parser

import (
	"bytes"
	"errors"
	"github.com/anaminus/luasyntax/go/scanner"
	"github.com/anaminus/luasyntax/go/token"
	"github.com/anaminus/luasyntax/go/tree"
	"io"
	"io/ioutil"
)

// tokenstate holds information about the current token.
type tokenstate struct {
	off int           // Offset of token.
	tok token.Type    // Type of token.
	lit []byte        // Literal bytes represented by token.
	pre []tree.Prefix // Accumulated prefix tokens.
}

// parser holds the parser's state while processing a source file. It must be
// initialized with init before using.
type parser struct {
	file    *token.File
	err     error //scanner.Error
	scanner scanner.Scanner

	tokenstate // Current token state.

	look *tokenstate // Store state for single-token lookaheads.
}

// init prepares the parser to parse a source. The info sets the file to use for
// positional information. The src is the text to be parsed. The mode configures
// how the parser behaves.
func (p *parser) init(info *token.File, src []byte) {
	p.file = info
	p.scanner.Init(p.file, src, func(pos token.Position, msg string) {
		p.err = scanner.Error{Position: pos, Message: msg}
	})
	p.next()
}

// next advances to the next token.
func (p *parser) next() {
	if p.look != nil {
		// Consume stored state.
		p.tokenstate = *p.look
		p.look = nil
		return
	}

	p.off, p.tok, p.lit = p.scanner.Scan()

	p.pre = nil
	// Skip over prefix tokens, accumulating them in p.pre.
	for p.tok.IsPrefix() {
		p.pre = append(p.pre, tree.Prefix{Type: p.tok, Bytes: p.lit})
		p.off, p.tok, p.lit = p.scanner.Scan()
	}
}

// lookahead looks at the next token without consuming current state. The
// lookahead state is stored in p.look, and is consumed on the next call to
// p.next().
func (p *parser) lookahead() {
	// Save current state.
	prev := p.tokenstate
	// Get next state.
	p.next()
	// Store next state for lookahead.
	next := p.tokenstate
	p.look = &next
	// Restore previous state.
	p.tokenstate = prev
}

// bailout is used when panicking to indicate an early termination.
type bailout struct{}

// errors stores the offset and message in p.err, then causes the parser to
// terminate.
func (p *parser) error(off int, msg string) {
	p.err = scanner.Error{
		Position: p.file.Position(off),
		Message:  msg,
	}
	panic(bailout{})
}

// expect asserts that the current state is of the given type.
func (p *parser) expect(tok token.Type) {
	if p.tok != tok {
		p.error(p.off, "'"+tok.String()+"' expected")
	}
}

// token creates a token node from the current state.
func (p *parser) token() tree.Token {
	return tree.Token{
		Type:   p.tok,
		Prefix: p.pre,
		Offset: p.off,
		Bytes:  p.lit,
	}
}

// tokenNext creates a token node from the current state, then advances to the
// next token.
func (p *parser) tokenNext() tree.Token {
	tok := p.token()
	p.next()
	return tok
}

// expectToken asserts that the current state is of the given type, creates an
// token node, then advances to the next token.
func (p *parser) expectToken(t token.Type) tree.Token {
	p.expect(t)
	return p.tokenNext()
}

// isBlockFollow returns whether the current state ends a block.
func (p *parser) isBlockFollow() bool {
	switch p.tok {
	case token.EOF,
		token.ELSE,
		token.ELSEIF,
		token.UNTIL,
		token.END:
		return true
	}
	return false
}

// parseNumber creates a number node from the current state.
func (p *parser) parseNumber() (num *tree.NumberExpr) {
	switch p.tok {
	case token.NUMBERFLOAT, token.NUMBERHEX:
		num = &tree.NumberExpr{NumberToken: p.token()}
	default:
		p.error(p.off, "'"+token.NUMBERFLOAT.String()+"' expected")
	}
	p.next()
	return
}

// parseString creates a string node from the current state.
func (p *parser) parseString() (str *tree.StringExpr) {
	switch p.tok {
	case token.STRING, token.LONGSTRING:
		str = &tree.StringExpr{StringToken: p.token()}
	default:
		p.error(p.off, "'"+token.STRING.String()+"' expected")
	}
	p.next()
	return
}

// parseSimpleExpr creates a simple expression node from the current state.
func (p *parser) parseSimpleExpr() (expr tree.Expr) {
	switch p.tok {
	case token.NUMBERFLOAT, token.NUMBERHEX:
		expr = p.parseNumber()
	case token.STRING, token.LONGSTRING:
		expr = p.parseString()
	case token.NIL:
		expr = &tree.NilExpr{NilToken: p.tokenNext()}
	case token.TRUE, token.FALSE:
		expr = &tree.BoolExpr{BoolToken: p.tokenNext()}
	case token.VARARG:
		expr = &tree.VarArgExpr{VarArgToken: p.tokenNext()}
	case token.LBRACE:
		expr = p.parseTableCtor()
	case token.FUNCTION:
		expr, _ = p.parseFunction(funcExpr)
	default:
		expr = p.parsePrimaryExpr()
	}
	return expr
}

// parseSubexpr recursively builds an expression chain.
func (p *parser) parseSubexpr(limit int) (expr tree.Expr) {
	if p.tok.IsUnary() {
		e := &tree.UnopExpr{}
		e.UnopToken = p.tokenNext()
		e.Operand = p.parseSubexpr(token.UnaryPrecedence)
		expr = e
	} else {
		expr = p.parseSimpleExpr()
		if expr == nil {
			p.error(p.off, "nil simpleexpr")
		}
	}

	for p.tok.IsBinary() && p.tok.Precedence()[0] > limit {
		binopToken := p.tokenNext()
		expr = &tree.BinopExpr{
			Left:       expr,
			BinopToken: binopToken,
			Right:      p.parseSubexpr(binopToken.Type.Precedence()[1]),
		}
	}

	return expr
}

// parseExpr begins parsing an expression chain.
func (p *parser) parseExpr() tree.Expr {
	return p.parseSubexpr(0)
}

// parseExprList creates a list of expressions.
func (p *parser) parseExprList() *tree.ExprList {
	list := &tree.ExprList{Items: []tree.Expr{p.parseExpr()}}
	for p.tok == token.COMMA {
		list.Seps = append(list.Seps, p.tokenNext())
		list.Items = append(list.Items, p.parseExpr())
	}
	return list
}

// parseBlockBody creates a block terminated by a specified token.
func (p *parser) parseBlockBody(term token.Type) tree.Block {
	block := p.parseBlock()
	if p.tok != term {
		p.error(p.off, term.String()+" expected")
	}
	return block
}

// parseDoStmt creates a `do` statement node.
func (p *parser) parseDoStmt() tree.Stmt {
	stmt := &tree.DoStmt{}
	stmt.DoToken = p.expectToken(token.DO)
	stmt.Body = p.parseBlockBody(token.END)
	stmt.EndToken = p.expectToken(token.END)
	return stmt
}

// parseWhileStmt creates a `while` statement node.
func (p *parser) parseWhileStmt() tree.Stmt {
	stmt := &tree.WhileStmt{}
	stmt.WhileToken = p.expectToken(token.WHILE)
	stmt.Cond = p.parseExpr()
	stmt.DoToken = p.expectToken(token.DO)
	stmt.Body = p.parseBlockBody(token.END)
	stmt.EndToken = p.expectToken(token.END)
	return stmt
}

// parseRepeatStmt creates a `repeat` statement node.
func (p *parser) parseRepeatStmt() tree.Stmt {
	stmt := &tree.RepeatStmt{}
	stmt.RepeatToken = p.expectToken(token.REPEAT)
	stmt.Body = p.parseBlockBody(token.UNTIL)
	stmt.UntilToken = p.expectToken(token.UNTIL)
	stmt.Cond = p.parseExpr()
	return stmt
}

// parseIfStmt creates an `if` statement node.
func (p *parser) parseIfStmt() tree.Stmt {
	stmt := &tree.IfStmt{}
	stmt.IfToken = p.expectToken(token.IF)
	stmt.Cond = p.parseExpr()
	stmt.ThenToken = p.expectToken(token.THEN)
	stmt.Body = p.parseBlock()
	for p.tok == token.ELSEIF {
		clause := tree.ElseIfClause{}
		clause.ElseIfToken = p.expectToken(token.ELSEIF)
		clause.Cond = p.parseExpr()
		clause.ThenToken = p.expectToken(token.THEN)
		clause.Body = p.parseBlock()
		stmt.ElseIf = append(stmt.ElseIf, clause)
	}
	if p.tok == token.ELSE {
		stmt.Else = &tree.ElseClause{}
		stmt.Else.ElseToken = p.expectToken(token.ELSE)
		stmt.Else.Body = p.parseBlock()
	}
	stmt.EndToken = p.expectToken(token.END)
	return stmt
}

// parseIfStmt creates a `for` statement node.
func (p *parser) parseForStmt() (stmt tree.Stmt) {
	forToken := p.expectToken(token.FOR)
	name := p.expectToken(token.NAME)
	switch p.tok {
	case token.ASSIGN:
		st := &tree.NumericForStmt{}
		st.ForToken = forToken
		st.NameToken = name
		st.AssignToken = p.expectToken(token.ASSIGN)
		st.Min = p.parseExpr()
		st.MaxSepToken = p.expectToken(token.COMMA)
		st.Max = p.parseExpr()
		if p.tok == token.COMMA {
			st.StepSepToken = p.expectToken(token.COMMA)
			st.Step = p.parseExpr()
		}
		st.DoToken = p.expectToken(token.DO)
		st.Body = p.parseBlockBody(token.END)
		st.EndToken = p.expectToken(token.END)
		stmt = st
	case token.COMMA, token.IN:
		st := &tree.GenericForStmt{}
		st.ForToken = forToken
		st.Names.Items = append(st.Names.Items, name)
		for p.tok == token.COMMA {
			st.Names.Seps = append(st.Names.Seps, p.tokenNext())
			st.Names.Items = append(st.Names.Items, p.expectToken(token.NAME))
		}
		st.InToken = p.expectToken(token.IN)
		st.Iterator = *p.parseExprList()
		st.DoToken = p.expectToken(token.DO)
		st.Body = p.parseBlockBody(token.END)
		st.EndToken = p.expectToken(token.END)
		stmt = st
	default:
		p.error(p.off, "'=' or 'in' expected")
	}
	return stmt
}

const (
	funcExpr  uint8 = iota // `function...` expression (anonymous).
	funcLocal              // `local function name...` statement.
	funcStmt               // `function name...` statement.
)

// parseFunction creates a node representing a function. The name of the
// function is parsed depending on the given type.
func (p *parser) parseFunction(typ uint8) (expr *tree.FunctionExpr, names tree.FuncNameList) {
	expr = &tree.FunctionExpr{}
	expr.FuncToken = p.expectToken(token.FUNCTION)
	if typ > funcExpr {
		names.Items = append(names.Items, p.expectToken(token.NAME))
		if typ > funcLocal {
			for p.tok == token.DOT {
				names.Seps = append(names.Seps, p.tokenNext())
				names.Items = append(names.Items, p.expectToken(token.NAME))
			}
			if p.tok == token.COLON {
				names.ColonToken = p.tokenNext()
				names.MethodToken = p.expectToken(token.NAME)
			}
		}
	}
	expr.LParenToken = p.expectToken(token.LPAREN)
	if p.tok == token.NAME {
		expr.Params = &tree.NameList{Items: []tree.Token{p.expectToken(token.NAME)}}
		for p.tok == token.COMMA {
			sepToken := p.tokenNext()
			if p.tok == token.VARARG {
				expr.VarArgSepToken = sepToken
				expr.VarArgToken = p.tokenNext()
				break
			}
			expr.Params.Seps = append(expr.Params.Seps, sepToken)
			expr.Params.Items = append(expr.Params.Items, p.expectToken(token.NAME))
		}
	} else if p.tok == token.VARARG {
		expr.VarArgToken = p.tokenNext()
	}
	expr.RParenToken = p.expectToken(token.RPAREN)
	expr.Body = p.parseBlockBody(token.END)
	expr.EndToken = p.expectToken(token.END)
	return expr, names
}

// parseLocalStmt creates a `local` statement node.
func (p *parser) parseLocalStmt() tree.Stmt {
	localToken := p.expectToken(token.LOCAL)
	if p.tok == token.FUNCTION {
		expr, names := p.parseFunction(funcLocal)
		return &tree.LocalFunctionStmt{
			LocalToken: localToken,
			NameToken:  names.Items[0],
			Func:       *expr,
		}
	}
	stmt := &tree.LocalVarStmt{}
	stmt.LocalToken = localToken
	stmt.Names.Items = append(stmt.Names.Items, p.expectToken(token.NAME))
	for p.tok == token.COMMA {
		stmt.Names.Seps = append(stmt.Names.Seps, p.tokenNext())
		stmt.Names.Items = append(stmt.Names.Items, p.expectToken(token.NAME))
	}
	if p.tok == token.ASSIGN {
		stmt.AssignToken = p.tokenNext()
		stmt.Values = p.parseExprList()
	}
	return stmt
}

// parseFunctionStmt creates a `function` statement node.
func (p *parser) parseFunctionStmt() tree.Stmt {
	expr, names := p.parseFunction(funcStmt)
	return &tree.FunctionStmt{
		Name: names,
		Func: *expr,
	}
}

// parseReturnStmt creates a `return` statement node.
func (p *parser) parseReturnStmt() tree.Stmt {
	stmt := &tree.ReturnStmt{}
	stmt.ReturnToken = p.expectToken(token.RETURN)
	if p.isBlockFollow() || p.tok == token.SEMICOLON {
		return stmt
	}
	stmt.Values = p.parseExprList()
	return stmt
}

// parseBreakStmt creates a `break` statement node.
func (p *parser) parseBreakStmt() tree.Stmt {
	stmt := &tree.BreakStmt{}
	stmt.BreakToken = p.expectToken(token.BREAK)
	return stmt
}

// parsePrefixExpr creates an expression node that begins a primary expression.
func (p *parser) parsePrefixExpr() (expr tree.Expr) {
	switch p.tok {
	case token.LPAREN:
		e := &tree.ParenExpr{}
		e.LParenToken = p.tokenNext()
		e.Value = p.parseExpr()
		e.RParenToken = p.expectToken(token.RPAREN)
		expr = e
	case token.NAME:
		e := &tree.VariableExpr{}
		e.NameToken = p.expectToken(token.NAME)
		expr = e
	default:
		p.error(p.off, "unexpected symbol")
	}
	return expr
}

// parseTableCtor creates a table constructor node.
func (p *parser) parseTableCtor() (ctor *tree.TableCtor) {
	ctor = &tree.TableCtor{}
	ctor.LBraceToken = p.expectToken(token.LBRACE)
	for p.tok != token.RBRACE {
		var entry tree.Entry
		if p.tok == token.LBRACK {
			e := &tree.IndexEntry{}
			e.LBrackToken = p.tokenNext()
			e.Key = p.parseExpr()
			e.RBrackToken = p.expectToken(token.RBRACK)
			e.AssignToken = p.expectToken(token.ASSIGN)
			e.Value = p.parseExpr()
			entry = e
		} else if p.lookahead(); p.tok == token.NAME && p.look.tok == token.ASSIGN {
			e := &tree.FieldEntry{}
			e.NameToken = p.expectToken(token.NAME)
			e.AssignToken = p.expectToken(token.ASSIGN)
			e.Value = p.parseExpr()
			entry = e
		} else {
			e := &tree.ValueEntry{}
			e.Value = p.parseExpr()
			entry = e
		}
		ctor.Entries.Items = append(ctor.Entries.Items, entry)
		if p.tok == token.COMMA || p.tok == token.SEMICOLON {
			ctor.Entries.Seps = append(ctor.Entries.Seps, p.tokenNext())
		} else {
			break
		}
	}
	ctor.RBraceToken = p.expectToken(token.RBRACE)
	return ctor
}

// parseFuncArgs creates a node representing the arguments of a function call.
func (p *parser) parseFuncArgs() (args tree.Args) {
	switch p.tok {
	case token.LPAREN:
		a := &tree.ListArgs{}
		a.LParenToken = p.tokenNext()
		for p.tok != token.RPAREN {
			if a.Values == nil {
				a.Values = &tree.ExprList{}
			}
			a.Values.Items = append(a.Values.Items, p.parseExpr())
			if p.tok == token.COMMA {
				a.Values.Seps = append(a.Values.Seps, p.tokenNext())
			} else {
				break
			}
		}
		a.RParenToken = p.expectToken(token.RPAREN)
		args = a
	case token.LBRACE:
		a := &tree.TableArg{}
		a.Value = *p.parseTableCtor()
		args = a
	case token.STRING, token.LONGSTRING:
		a := &tree.StringArg{}
		a.Value = *p.parseString()
		args = a
	default:
		p.error(p.off, "function arguments expected")
	}
	return args
}

// parsePrimaryExpr creates a primary expression node that begins an expression
// chain.
func (p *parser) parsePrimaryExpr() (expr tree.Expr) {
loop:
	for expr = p.parsePrefixExpr(); ; {
		switch p.tok {
		case token.DOT:
			e := &tree.FieldExpr{}
			e.Value = expr
			e.DotToken = p.tokenNext()
			e.NameToken = p.expectToken(token.NAME)
			expr = e
		case token.COLON:
			e := &tree.MethodExpr{}
			e.Value = expr
			e.ColonToken = p.tokenNext()
			e.NameToken = p.expectToken(token.NAME)
			e.Args = p.parseFuncArgs()
			expr = e
		case token.LBRACK:
			e := &tree.IndexExpr{}
			e.Value = expr
			e.LBrackToken = p.tokenNext()
			e.Index = p.parseExpr()
			e.RBrackToken = p.expectToken(token.RBRACK)
			expr = e
		case token.LBRACE, token.LPAREN:
			e := &tree.CallExpr{}
			e.Value = expr
			e.Args = p.parseFuncArgs()
			expr = e
		default:
			break loop
		}
	}
	return expr
}

// parseExprStmt creates an expression statement node.
func (p *parser) parseExprStmt() tree.Stmt {
	expr := p.parsePrimaryExpr()
	if call, ok := expr.(tree.Call); ok {
		return &tree.CallStmt{Call: call}
	}

	stmt := &tree.AssignStmt{Left: tree.ExprList{Items: []tree.Expr{expr}}}
	for p.tok == token.COMMA {
		stmt.Left.Seps = append(stmt.Left.Seps, p.tokenNext())
		switch expr := p.parsePrimaryExpr().(type) {
		case *tree.MethodExpr, *tree.CallExpr:
			p.error(p.off, "syntax error")
		default:
			stmt.Left.Items = append(stmt.Left.Items, expr)
		}
	}
	stmt.AssignToken = p.expectToken(token.ASSIGN)
	stmt.Right = *p.parseExprList()
	return stmt
}

// parseStmt creates a statement node. Returns the node, and whether the
// statement is meant to be the last statement in the block.
func (p *parser) parseStmt() (stmt tree.Stmt, last bool) {
	switch p.tok {
	case token.DO:
		return p.parseDoStmt(), false
	case token.WHILE:
		return p.parseWhileStmt(), false
	case token.REPEAT:
		return p.parseRepeatStmt(), false
	case token.IF:
		return p.parseIfStmt(), false
	case token.FOR:
		return p.parseForStmt(), false
	case token.FUNCTION:
		return p.parseFunctionStmt(), false
	case token.LOCAL:
		return p.parseLocalStmt(), false
	case token.RETURN:
		return p.parseReturnStmt(), true
	case token.BREAK:
		return p.parseBreakStmt(), true
	}
	return p.parseExprStmt(), false
}

// parseBlock creates a block node.
func (p *parser) parseBlock() (block tree.Block) {
	for last := false; !last && !p.isBlockFollow(); {
		var stmt tree.Stmt
		stmt, last = p.parseStmt()
		block.Items = append(block.Items, stmt)
		var semi tree.Token
		if p.tok == token.SEMICOLON {
			semi = p.tokenNext()
		}
		block.Seps = append(block.Seps, semi)
	}
	return
}

// parseFile creates a file node from the current source.
func (p *parser) parseFile() *tree.File {
	return &tree.File{
		Info:     p.file,
		Body:     p.parseBlock(),
		EOFToken: p.tokenNext(),
	}
}

// readSource retrieves the bytes from several types of values.
func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, s); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
		return nil, errors.New("invalid source")
	}
	return ioutil.ReadFile(filename)
}

// ParseFile parses the source code of a single Lua file. It returns a root
// tree.File node representing the parsed file, any any errors that may have
// occurred while parsing.
//
// The src argument may be a string, []byte, *bytes.Buffer, or io.Reader. In
// these cases, the filename is used only when recording positional information.
// If src is nil, the source is read from the file specified by filename.
func ParseFile(filename string, src interface{}) (f *tree.File, err error) {
	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	info := token.NewFile(filename)
	var p parser
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}

		if f == nil {
			f = &tree.File{Info: info}
		}

		err = p.err
	}()

	p.init(info, text)
	f = p.parseFile()
	return f, err
}
