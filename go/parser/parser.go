// The parser package implements a parser for Lua source files. Input may be
// provided in a variety of forms, and the output is an abstract syntax tree
// (AST) representing the Lua source.
package parser

import (
	"bytes"
	"errors"
	"github.com/anaminus/luasyntax/go/ast"
	"github.com/anaminus/luasyntax/go/scanner"
	"github.com/anaminus/luasyntax/go/token"
	"io"
	"io/ioutil"
	"math"
	"strconv"
)

// tokenstate holds information about the current token.
type tokenstate struct {
	off int          // Offset of token.
	tok token.Type   // Type of token.
	lit []byte       // Literal bytes represented by token.
	pre []ast.Prefix // Accumulated prefix tokens.
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

// init prepares the parser to parse a source. The filename sets the name to
// use for positional information. The src is the text to be parsed.
func (p *parser) init(filename string, src []byte) {
	p.file = token.NewFile(filename)
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
		p.pre = append(p.pre, ast.Prefix{Type: p.tok, Bytes: p.lit})
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
func (p *parser) token() ast.Token {
	return ast.Token{
		Type:   p.tok,
		Prefix: p.pre,
		Offset: p.off,
		Bytes:  p.lit,
	}
}

// tokenNext creates a token node from the current state, then advances to the
// next token.
func (p *parser) tokenNext() ast.Token {
	tok := p.token()
	p.next()
	return tok
}

// expectToken asserts that the current state is of the given type, creates an
// token node, then advances to the next token.
func (p *parser) expectToken(t token.Type) ast.Token {
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

// parseName creates a name node from the current state.
func (p *parser) parseName() (name ast.Name) {
	p.expect(token.NAME)
	name = ast.Name{Token: p.token(), Value: string(p.lit)}
	p.next()
	return
}

// parseNumber creates a number node from the current state.
func (p *parser) parseNumber() (num *ast.Number) {
	var n float64
	var err error
	switch p.tok {
	case token.NUMBERFLOAT:
		// Actual parsing of the number depends on the compiler (strtod), so
		// technically it's correct to just use Go's parser.
		n, err = strconv.ParseFloat(string(p.lit), 64)
	case token.NUMBERHEX:
		var i uint64
		// Trim leading `0x`.
		i, err = strconv.ParseUint(string(p.lit[2:]), 16, 32)
		n = float64(i)
	default:
		p.error(p.off, "'"+token.NUMBERFLOAT.String()+"' expected")
	}
	if err != nil {
		n = math.NaN()
	}
	num = &ast.Number{Token: p.token(), Value: n}
	p.next()
	return
}

// parseQuotedString parses literal quoted string into actual text.
func parseQuotedString(b []byte) string {
	b = b[1 : len(b)-1]          // Trim quotes.
	c := make([]byte, 0, len(b)) // Result will never be larger than source.
	for i := 0; i < len(b); i++ {
		ch := b[i]
		if ch == '\\' {
			i++
			ch = b[i]
			switch ch {
			case 'a':
				ch = '\a'
			case 'b':
				ch = '\b'
			case 'f':
				ch = '\f'
			case 'n':
				ch = '\n'
			case 'r':
				ch = '\r'
			case 't':
				ch = '\t'
			case 'v':
				ch = '\v'
			default:
				if '0' <= ch && ch <= '9' {
					var n byte
					for j := 0; j < 3 && '0' <= b[i] && b[i] <= '9'; j++ {
						n = n*10 + (b[i] - '0')
						i++
					}
					// Size of number was already checked by scanner.
					ch = n
				}
			}
		}
		c = append(c, ch)
	}
	return string(c)
}

// parseBlockString parses a literal long string into actual text.
func parseBlockString(b []byte) string {
	// Assumes string is wrapped in a `[==[]==]`-like block.
	b = b[1:] // Trim first `[`
	for i, c := range b {
		if c == '[' {
			// Trim to second '[', as well as trailing block.
			b = b[i+1 : len(b)-i-2]
		}
	}
	// Skip first newline.
	if len(b) > 0 && (b[0] == '\n' || b[0] == '\r') {
		if len(b) > 1 && (b[1] == '\n' || b[1] == '\r') && b[1] != b[0] {
			b = b[2:]
		} else {
			b = b[1:]
		}
	}
	return string(b)
}

// parseString creates a string node from the current state.
func (p *parser) parseString() (str *ast.String) {
	switch p.tok {
	case token.STRING:
		str = &ast.String{Token: p.token(), Value: parseQuotedString(p.lit)}
	case token.LONGSTRING:
		str = &ast.String{Token: p.token(), Value: parseBlockString(p.lit)}
	default:
		p.error(p.off, "'"+token.STRING.String()+"' expected")
	}
	p.next()
	return
}

// parseSimpleExpr creates a simple expression node from the current state.
func (p *parser) parseSimpleExpr() (expr ast.Expr) {
	switch p.tok {
	case token.NUMBERFLOAT, token.NUMBERHEX:
		expr = p.parseNumber()
	case token.STRING, token.LONGSTRING:
		expr = p.parseString()
	case token.NIL:
		expr = &ast.Nil{Token: p.tokenNext()}
	case token.TRUE:
		expr = &ast.Bool{Token: p.tokenNext(), Value: true}
	case token.FALSE:
		expr = &ast.Bool{Token: p.tokenNext(), Value: false}
	case token.VARARG:
		expr = &ast.VarArg{Token: p.tokenNext()}
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
func (p *parser) parseSubexpr(limit int) (expr ast.Expr) {
	if p.tok.IsUnary() {
		e := &ast.UnopExpr{}
		e.UnopToken = p.tokenNext()
		e.Expr = p.parseSubexpr(token.UnaryPrecedence)
		expr = e
	} else {
		expr = p.parseSimpleExpr()
		if expr == nil {
			p.error(p.off, "nil simpleexpr")
		}
	}

	for p.tok.IsBinary() && p.tok.Precedence()[0] > limit {
		binopToken := p.tokenNext()
		expr = &ast.BinopExpr{
			Left:       expr,
			BinopToken: binopToken,
			Right:      p.parseSubexpr(binopToken.Type.Precedence()[1]),
		}
	}

	return expr
}

// parseExpr begins parsing an expression chain.
func (p *parser) parseExpr() ast.Expr {
	return p.parseSubexpr(0)
}

// parseExprList creates a list of expressions.
func (p *parser) parseExprList() *ast.ExprList {
	list := &ast.ExprList{Exprs: []ast.Expr{p.parseExpr()}}
	for p.tok == token.COMMA {
		list.Seps = append(list.Seps, p.tokenNext())
		list.Exprs = append(list.Exprs, p.parseExpr())
	}
	return list
}

// parseBlockBody creates a block terminated by a specified token.
func (p *parser) parseBlockBody(term token.Type) ast.Block {
	block := p.parseBlock()
	if p.tok != term {
		p.error(p.off, term.String()+" expected")
	}
	return block
}

// parseDoStmt creates a `do` statement node.
func (p *parser) parseDoStmt() ast.Stmt {
	stmt := &ast.DoStmt{}
	stmt.DoToken = p.expectToken(token.DO)
	stmt.Block = p.parseBlockBody(token.END)
	stmt.EndToken = p.expectToken(token.END)
	return stmt
}

// parseWhileStmt creates a `while` statement node.
func (p *parser) parseWhileStmt() ast.Stmt {
	stmt := &ast.WhileStmt{}
	stmt.WhileToken = p.expectToken(token.WHILE)
	stmt.Expr = p.parseExpr()
	stmt.DoToken = p.expectToken(token.DO)
	stmt.Block = p.parseBlockBody(token.END)
	stmt.EndToken = p.expectToken(token.END)
	return stmt
}

// parseRepeatStmt creates a `repeat` statement node.
func (p *parser) parseRepeatStmt() ast.Stmt {
	stmt := &ast.RepeatStmt{}
	stmt.RepeatToken = p.expectToken(token.REPEAT)
	stmt.Block = p.parseBlockBody(token.UNTIL)
	stmt.UntilToken = p.expectToken(token.UNTIL)
	stmt.Expr = p.parseExpr()
	return stmt
}

// parseIfStmt creates an `if` statement node.
func (p *parser) parseIfStmt() ast.Stmt {
	stmt := &ast.IfStmt{}
	stmt.IfToken = p.expectToken(token.IF)
	stmt.Expr = p.parseExpr()
	stmt.ThenToken = p.expectToken(token.THEN)
	stmt.Block = p.parseBlock()
	for p.tok == token.ELSEIF {
		clause := ast.ElseIfClause{}
		clause.ElseIfToken = p.expectToken(token.ELSEIF)
		clause.Expr = p.parseExpr()
		clause.ThenToken = p.expectToken(token.THEN)
		clause.Block = p.parseBlock()
		stmt.ElseIfClauses = append(stmt.ElseIfClauses, clause)
	}
	if p.tok == token.ELSE {
		stmt.ElseClause = &ast.ElseClause{}
		stmt.ElseClause.ElseToken = p.expectToken(token.ELSE)
		stmt.ElseClause.Block = p.parseBlock()
	}
	stmt.EndToken = p.expectToken(token.END)
	return stmt
}

// parseIfStmt creates a `for` statement node.
func (p *parser) parseForStmt() (stmt ast.Stmt) {
	forToken := p.expectToken(token.FOR)
	name := p.parseName()
	switch p.tok {
	case token.ASSIGN:
		st := &ast.NumericForStmt{}
		st.ForToken = forToken
		st.Name = name
		st.AssignToken = p.expectToken(token.ASSIGN)
		st.MinExpr = p.parseExpr()
		st.MaxSepToken = p.expectToken(token.COMMA)
		st.MaxExpr = p.parseExpr()
		if p.tok == token.COMMA {
			st.StepSepToken = p.expectToken(token.COMMA)
			st.StepExpr = p.parseExpr()
		}
		st.DoToken = p.expectToken(token.DO)
		st.Block = p.parseBlockBody(token.END)
		st.EndToken = p.expectToken(token.END)
		stmt = st
	case token.COMMA, token.IN:
		st := &ast.GenericForStmt{}
		st.ForToken = forToken
		st.NameList.Names = append(st.NameList.Names, name)
		for p.tok == token.COMMA {
			st.NameList.Seps = append(st.NameList.Seps, p.tokenNext())
			st.NameList.Names = append(st.NameList.Names, p.parseName())
		}
		st.InToken = p.expectToken(token.IN)
		st.ExprList = *p.parseExprList()
		st.DoToken = p.expectToken(token.DO)
		st.Block = p.parseBlockBody(token.END)
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
func (p *parser) parseFunction(typ uint8) (expr *ast.FunctionExpr, names ast.FuncNameList) {
	expr = &ast.FunctionExpr{}
	expr.FuncToken = p.expectToken(token.FUNCTION)
	if typ > funcExpr {
		names.Names = append(names.Names, p.parseName())
		if typ > funcLocal {
			for p.tok == token.DOT {
				names.Seps = append(names.Seps, p.tokenNext())
				names.Names = append(names.Names, p.parseName())
			}
			if p.tok == token.COLON {
				names.Seps = append(names.Seps, p.tokenNext())
				names.Names = append(names.Names, p.parseName())
			}
		}
	}
	expr.LParenToken = p.expectToken(token.LPAREN)
	if p.tok == token.NAME {
		expr.ParamList = &ast.NameList{Names: []ast.Name{p.parseName()}}
		for p.tok == token.COMMA {
			sepToken := p.tokenNext()
			if p.tok == token.VARARG {
				expr.VarArgSepToken = sepToken
				expr.VarArgToken = p.tokenNext()
				break
			}
			expr.ParamList.Seps = append(expr.ParamList.Seps, sepToken)
			expr.ParamList.Names = append(expr.ParamList.Names, p.parseName())
		}
	} else if p.tok == token.VARARG {
		expr.VarArgToken = p.tokenNext()
	}
	expr.RParenToken = p.expectToken(token.RPAREN)
	expr.Block = p.parseBlockBody(token.END)
	expr.EndToken = p.expectToken(token.END)
	return expr, names
}

// parseLocalStmt creates a `local` statement node.
func (p *parser) parseLocalStmt() ast.Stmt {
	localToken := p.expectToken(token.LOCAL)
	if p.tok == token.FUNCTION {
		expr, names := p.parseFunction(funcLocal)
		return &ast.LocalFunctionStmt{
			LocalToken: localToken,
			Name:       names.Names[0],
			Expr:       *expr,
		}
	}
	stmt := &ast.LocalVarStmt{}
	stmt.LocalToken = localToken
	stmt.NameList.Names = append(stmt.NameList.Names, p.parseName())
	for p.tok == token.COMMA {
		stmt.NameList.Seps = append(stmt.NameList.Seps, p.tokenNext())
		stmt.NameList.Names = append(stmt.NameList.Names, p.parseName())
	}
	if p.tok == token.ASSIGN {
		stmt.AssignToken = p.tokenNext()
		stmt.ExprList = p.parseExprList()
	}
	return stmt
}

// parseFunctionStmt creates a `function` statement node.
func (p *parser) parseFunctionStmt() ast.Stmt {
	expr, names := p.parseFunction(funcStmt)
	return &ast.FunctionStmt{
		Name: names,
		Expr: *expr,
	}
}

// parseReturnStmt creates a `return` statement node.
func (p *parser) parseReturnStmt() ast.Stmt {
	stmt := &ast.ReturnStmt{}
	stmt.ReturnToken = p.expectToken(token.RETURN)
	if p.isBlockFollow() || p.tok == token.SEMICOLON {
		return stmt
	}
	stmt.ExprList = p.parseExprList()
	return stmt
}

// parseBreakStmt creates a `break` statement node.
func (p *parser) parseBreakStmt() ast.Stmt {
	stmt := &ast.BreakStmt{}
	stmt.BreakToken = p.expectToken(token.BREAK)
	return stmt
}

// parsePrefixExpr creates an expression node that begins a primary
// expression.
func (p *parser) parsePrefixExpr() (expr ast.Expr) {
	switch p.tok {
	case token.LPAREN:
		e := &ast.ParenExpr{}
		e.LParenToken = p.tokenNext()
		e.Expr = p.parseExpr()
		e.RParenToken = p.expectToken(token.RPAREN)
		expr = e
	case token.NAME:
		e := &ast.VariableExpr{}
		e.NameToken = p.parseName()
		expr = e
	default:
		p.error(p.off, "unexpected symbol")
	}
	return expr
}

// parseTableCtor creates a table constructor node.
func (p *parser) parseTableCtor() (ctor *ast.TableCtor) {
	ctor = &ast.TableCtor{}
	ctor.LBraceToken = p.expectToken(token.LBRACE)
	for p.tok != token.RBRACE {
		var entry ast.Entry
		if p.tok == token.LBRACK {
			e := &ast.IndexEntry{}
			e.LBrackToken = p.tokenNext()
			e.KeyExpr = p.parseExpr()
			e.RBrackToken = p.expectToken(token.RBRACK)
			e.AssignToken = p.expectToken(token.ASSIGN)
			e.ValueExpr = p.parseExpr()
			entry = e
		} else if p.lookahead(); p.tok == token.NAME && p.look.tok == token.ASSIGN {
			e := &ast.FieldEntry{}
			e.Name = p.parseName()
			e.AssignToken = p.expectToken(token.ASSIGN)
			e.Value = p.parseExpr()
			entry = e
		} else {
			e := &ast.ValueEntry{}
			e.Value = p.parseExpr()
			entry = e
		}
		ctor.EntryList.Entries = append(ctor.EntryList.Entries, entry)
		if p.tok == token.COMMA || p.tok == token.SEMICOLON {
			ctor.EntryList.Seps = append(ctor.EntryList.Seps, p.tokenNext())
		} else {
			break
		}
	}
	ctor.RBraceToken = p.expectToken(token.RBRACE)
	return ctor
}

// parseFuncArgs creates a node representing the arguments of a function call.
func (p *parser) parseFuncArgs() (args ast.CallArgs) {
	switch p.tok {
	case token.LPAREN:
		a := &ast.ArgsCall{}
		a.LParenToken = p.tokenNext()
		for p.tok != token.RPAREN {
			if a.ExprList == nil {
				a.ExprList = &ast.ExprList{}
			}
			a.ExprList.Exprs = append(a.ExprList.Exprs, p.parseExpr())
			if p.tok == token.COMMA {
				a.ExprList.Seps = append(a.ExprList.Seps, p.tokenNext())
			} else {
				break
			}
		}
		a.RParenToken = p.expectToken(token.RPAREN)
		args = a
	case token.LBRACE:
		a := &ast.TableCall{}
		a.TableExpr = *p.parseTableCtor()
		args = a
	case token.STRING, token.LONGSTRING:
		a := &ast.StringCall{}
		a.StringExpr = *p.parseString()
		args = a
	default:
		p.error(p.off, "function arguments expected")
	}
	return args
}

// parsePrimaryExpr creates a primary expression node that begins an
// expression chain.
func (p *parser) parsePrimaryExpr() (expr ast.Expr) {
loop:
	for expr = p.parsePrefixExpr(); ; {
		switch p.tok {
		case token.DOT:
			e := &ast.FieldExpr{}
			e.Expr = expr
			e.DotToken = p.tokenNext()
			e.Field = p.parseName()
			expr = e
		case token.COLON:
			e := &ast.MethodExpr{}
			e.Expr = expr
			e.ColonToken = p.tokenNext()
			e.Name = p.parseName()
			e.Args = p.parseFuncArgs()
			expr = e
		case token.LBRACK:
			e := &ast.IndexExpr{}
			e.Expr = expr
			e.LBrackToken = p.tokenNext()
			e.Index = p.parseExpr()
			e.RBrackToken = p.expectToken(token.RBRACK)
			expr = e
		case token.LBRACE, token.LPAREN:
			e := &ast.CallExpr{}
			e.Expr = expr
			e.Args = p.parseFuncArgs()
			expr = e
		default:
			break loop
		}
	}
	return expr
}

// parseExprStmt creates an expression statement node.
func (p *parser) parseExprStmt() ast.Stmt {
	expr := p.parsePrimaryExpr()
	switch expr.(type) {
	case *ast.MethodExpr, *ast.CallExpr:
		return &ast.CallExprStmt{Expr: expr}
	}

	stmt := &ast.AssignStmt{Left: ast.ExprList{Exprs: []ast.Expr{expr}}}
	for p.tok == token.COMMA {
		stmt.Left.Seps = append(stmt.Left.Seps, p.tokenNext())
		switch expr := p.parsePrimaryExpr().(type) {
		case *ast.MethodExpr, *ast.CallExpr:
			p.error(p.off, "syntax error")
		default:
			stmt.Left.Exprs = append(stmt.Left.Exprs, expr)
		}
	}
	stmt.AssignToken = p.expectToken(token.ASSIGN)
	stmt.Right = *p.parseExprList()
	return stmt
}

// parseStmt creates a statement node. Returns the node, and whether the
// statement is meant to be the last statement in the block.
func (p *parser) parseStmt() (stmt ast.Stmt, last bool) {
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
func (p *parser) parseBlock() (block ast.Block) {
	for last := false; !last && !p.isBlockFollow(); {
		var stmt ast.Stmt
		stmt, last = p.parseStmt()
		block.Stmts = append(block.Stmts, stmt)
		var semi ast.Token
		if p.tok == token.SEMICOLON {
			semi = p.tokenNext()
		}
		block.Seps = append(block.Seps, semi)
	}
	return
}

// parseFile creates a file node from the current source.
func (p *parser) parseFile() *ast.File {
	return &ast.File{
		Name:     p.file.Name(),
		Block:    p.parseBlock(),
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
// ast.File node representing the parsed file, any any errors that may have
// occurred while parsing.
//
// The src argument may be a string, []byte, *bytes.Buffer, or io.Reader. In
// these cases, the filename is used only when recording positional
// information. If src is nil, the source is read from the file specified by
// filename.
func ParseFile(filename string, src interface{}) (f *ast.File, err error) {
	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}

		if f == nil {
			f = &ast.File{Name: filename}
		}

		err = p.err
	}()

	p.init(filename, text)
	f = p.parseFile()
	return f, err
}
