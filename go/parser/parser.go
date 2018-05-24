package parser

import (
	"bytes"
	"errors"
	"github.com/anaminus/luasyntax/go/ast"
	"github.com/anaminus/luasyntax/go/scanner"
	"github.com/anaminus/luasyntax/go/token"
	"io"
	"io/ioutil"
)

type lookstate struct {
	off   int
	tok   token.Token
	lit   []byte
	space int
}

type parser struct {
	file    *token.File
	err     error //scanner.Error
	scanner scanner.Scanner

	off   int
	tok   token.Token
	lit   []byte
	space int // Offset of previous whitespace.

	look *lookstate // Store state for single-token lookaheads.
}

func (p *parser) init(filename string, src []byte) {
	p.file = token.NewFile(filename)
	p.scanner.Init(p.file, src, func(pos token.Position, msg string) {
		p.err = scanner.Error{Position: pos, Message: msg}
	})
	p.next()
}

func (p *parser) next() {
	if p.look != nil {
		// Consume stored state.
		p.off = p.look.off
		p.tok = p.look.tok
		p.lit = p.look.lit
		p.space = p.look.space
		p.look = nil
		return
	}
	// Mark as invalid for when there are no spaces.
	p.space = -1

	// TODO: Handle comments. Currently they are just treated as whitespace.
	prevTok := p.tok
	p.off, p.tok, p.lit = p.scanner.Scan()
	// Skip over space tokens, setting p.space to the offset of the first
	// space.
	for p.tok.IsSpace() {
		if !prevTok.IsSpace() {
			p.space = p.off
		}
		prevTok = p.tok
		p.off, p.tok, p.lit = p.scanner.Scan()
	}
}

// Look at next token without consuming current token. The lookahead state is
// stored in p.look, and is consumed on the next call to p.next().
func (p *parser) lookahead() {
	// Save current state.
	prev := lookstate{p.off, p.tok, p.lit, p.space}
	// Get next state.
	p.next()
	// Store next state for lookahead.
	p.look = &lookstate{p.off, p.tok, p.lit, p.space}
	// Restore previous state.
	p.off = prev.off
	p.tok = prev.tok
	p.lit = prev.lit
	p.space = prev.space
}

type bailout struct{}

func (p *parser) error(off int, msg string) {
	p.err = scanner.Error{
		Position: p.file.Position(off),
		Message:  msg,
	}
	panic(bailout{})
}

func (p *parser) expect(tok token.Token) {
	if p.tok != tok {
		p.error(p.off, "'"+tok.String()+"' expected")
	}
}

func (p *parser) token() ast.Token {
	return ast.Token{
		Space:  p.space,
		Offset: p.off,
		Type:   p.tok,
	}
}

func (p *parser) tokenNext() ast.Token {
	tok := p.token()
	p.next()
	return tok
}

func (p *parser) expectToken(t token.Token) ast.Token {
	p.expect(t)
	return p.tokenNext()
}

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

func (p *parser) parseName() (name ast.Name) {
	p.expect(token.NAME)
	name = ast.Name{Token: p.token(), Value: p.lit}
	p.next()
	return
}

func (p *parser) parseNumber() (num ast.Number) {
	p.expect(token.NUMBER)
	num = ast.Number{Token: p.token(), Value: p.lit}
	p.next()
	return
}

func (p *parser) parseString() (str ast.String) {
	if p.tok != token.STRING && p.tok != token.LONGSTRING {
		p.error(p.off, "'"+token.STRING.String()+"' expected")
	}
	str = ast.String{Token: p.token(), Value: p.lit}
	p.next()
	return
}

func (p *parser) parseSimpleExp() (exp ast.Exp) {
	switch p.tok {
	case token.NUMBER:
		exp = p.parseNumber()
	case token.STRING, token.LONGSTRING:
		exp = p.parseString()
	case token.NIL:
		exp = ast.Nil{Token: p.tokenNext()}
	case token.TRUE:
		exp = ast.Bool{Token: p.tokenNext(), Value: true}
	case token.FALSE:
		exp = ast.Bool{Token: p.tokenNext(), Value: false}
	case token.VARARG:
		exp = ast.VarArg{Token: p.tokenNext()}
	case token.LBRACE:
		exp = p.parseTableCtor()
	case token.FUNCTION:
		exp = p.parseFunction(funcExp)
	default:
		exp = p.parsePrimaryExp()
	}
	return exp
}

func (p *parser) parseSubexp(limit int) (exp ast.Exp) {
	if p.tok.IsUnary() {
		e := &ast.UnopExp{}
		e.UnopToken = p.tokenNext()
		e.Exp = p.parseSubexp(token.UnaryPrecedence)
		exp = e
	} else {
		exp = p.parseSimpleExp()
		if exp == nil {
			p.error(p.off, "nil simpleexp")
		}
	}

	for p.tok.IsBinary() && p.tok.Precedence()[0] > limit {
		binopToken := p.tokenNext()
		exp = &ast.BinopExp{
			Left:       exp,
			BinopToken: binopToken,
			Right:      p.parseSubexp(binopToken.Type.Precedence()[1]),
		}
	}

	return exp
}

func (p *parser) parseExp() ast.Exp {
	return p.parseSubexp(0)
}

func (p *parser) parseExpList() *ast.ExpList {
	list := &ast.ExpList{Exps: []ast.Exp{p.parseExp()}}
	for p.tok == token.COMMA {
		list.Seps = append(list.Seps, p.tokenNext())
		list.Exps = append(list.Exps, p.parseExp())
	}
	return list
}

func (p *parser) parseBlockBody(term token.Token) *ast.Block {
	block := p.parseBlock()
	if p.tok != term {
		p.error(p.off, term.String()+" expected")
	}
	return block
}

func (p *parser) parseDoStat() ast.Stat {
	stat := &ast.DoStat{}
	stat.DoToken = p.expectToken(token.DO)
	stat.Block = p.parseBlockBody(token.END)
	stat.EndToken = p.expectToken(token.END)
	return stat
}

func (p *parser) parseWhileStat() ast.Stat {
	stat := &ast.WhileStat{}
	stat.WhileToken = p.expectToken(token.WHILE)
	stat.Exp = p.parseExp()
	stat.DoToken = p.expectToken(token.DO)
	stat.Block = p.parseBlockBody(token.END)
	stat.EndToken = p.expectToken(token.END)
	return stat
}

func (p *parser) parseRepeatStat() ast.Stat {
	stat := &ast.RepeatStat{}
	stat.RepeatToken = p.expectToken(token.REPEAT)
	stat.Block = p.parseBlockBody(token.UNTIL)
	stat.UntilToken = p.expectToken(token.UNTIL)
	stat.Exp = p.parseExp()
	return stat
}

func (p *parser) parseIfStat() ast.Stat {
	stat := &ast.IfStat{}
	stat.IfToken = p.expectToken(token.IF)
	stat.Exp = p.parseExp()
	stat.ThenToken = p.expectToken(token.THEN)
	stat.Block = p.parseBlock()
	for p.tok == token.ELSEIF {
		clause := &ast.ElseIfClause{}
		clause.ElseIfToken = p.expectToken(token.ELSEIF)
		clause.Exp = p.parseExp()
		clause.ThenToken = p.expectToken(token.THEN)
		clause.Block = p.parseBlock()
		stat.ElseIfClauses = append(stat.ElseIfClauses, clause)
	}
	if p.tok == token.ELSE {
		stat.ElseClause = &ast.ElseClause{}
		stat.ElseClause.ElseToken = p.expectToken(token.ELSE)
		stat.ElseClause.Block = p.parseBlock()
	}
	stat.EndToken = p.expectToken(token.END)
	return stat
}

func (p *parser) parseForStat() (stat ast.Stat) {
	forToken := p.expectToken(token.FOR)
	name := p.parseName()
	switch p.tok {
	case token.ASSIGN:
		st := &ast.NumericForStat{}
		st.ForToken = forToken
		st.Name = name
		st.AssignToken = p.expectToken(token.ASSIGN)
		st.MinExp = p.parseExp()
		st.MaxSepToken = p.expectToken(token.COMMA)
		st.MaxExp = p.parseExp()
		if p.tok == token.COMMA {
			st.StepSepToken = p.expectToken(token.COMMA)
			st.StepExp = p.parseExp()
		}
		st.DoToken = p.expectToken(token.DO)
		st.Block = p.parseBlockBody(token.END)
		st.EndToken = p.expectToken(token.END)
		stat = st
	case token.COMMA, token.IN:
		st := &ast.GenericForStat{}
		st.ForToken = forToken
		st.NameList = &ast.NameList{Names: []ast.Name{name}}
		for p.tok == token.COMMA {
			st.NameList.Seps = append(st.NameList.Seps, p.tokenNext())
			st.NameList.Names = append(st.NameList.Names, p.parseName())
		}
		st.InToken = p.expectToken(token.IN)
		st.ExpList = p.parseExpList()
		st.DoToken = p.expectToken(token.DO)
		st.Block = p.parseBlockBody(token.END)
		st.EndToken = p.expectToken(token.END)
		stat = st
	default:
		p.error(p.off, "'=' or 'in' expected")
	}
	return stat
}

const (
	funcExp uint8 = iota
	funcLocal
	funcStat
)

func (p *parser) parseFunction(typ uint8) *ast.Function {
	stat := &ast.Function{}
	stat.FuncToken = p.expectToken(token.FUNCTION)
	if typ > funcExp {
		stat.FuncName = &ast.NameList{
			Names: []ast.Name{p.parseName()},
		}
		if typ > funcLocal {
			for p.tok == token.DOT {
				stat.FuncName.Seps = append(stat.FuncName.Seps, p.tokenNext())
				stat.FuncName.Names = append(stat.FuncName.Names, p.parseName())
			}
			if p.tok == token.COLON {
				stat.FuncName.Seps = append(stat.FuncName.Seps, p.tokenNext())
				stat.FuncName.Names = append(stat.FuncName.Names, p.parseName())
			}
		}
	}
	stat.LParenToken = p.expectToken(token.LPAREN)
	if p.tok == token.NAME {
		stat.ParList = &ast.NameList{Names: []ast.Name{p.parseName()}}
		for p.tok == token.COMMA {
			sepToken := p.tokenNext()
			if p.tok == token.VARARG {
				stat.VarArgSepToken = sepToken
				stat.VarArgToken = p.tokenNext()
				break
			}
			stat.ParList.Seps = append(stat.ParList.Seps, sepToken)
			stat.ParList.Names = append(stat.ParList.Names, p.parseName())
		}
	} else if p.tok == token.VARARG {
		stat.VarArgToken = p.tokenNext()
	}
	stat.RParenToken = p.expectToken(token.RPAREN)
	stat.Block = p.parseBlockBody(token.END)
	stat.EndToken = p.expectToken(token.END)
	return stat
}

func (p *parser) parseLocalStat() ast.Stat {
	localToken := p.expectToken(token.LOCAL)
	if p.tok == token.FUNCTION {
		return &ast.LocalFunctionStat{
			LocalToken: localToken,
			Function:   *p.parseFunction(funcLocal),
		}
	}
	stat := &ast.LocalVarStat{
		LocalToken: localToken,
		NameList:   &ast.NameList{Names: []ast.Name{p.parseName()}},
	}
	for p.tok == token.COMMA {
		stat.NameList.Seps = append(stat.NameList.Seps, p.tokenNext())
		stat.NameList.Names = append(stat.NameList.Names, p.parseName())
	}
	if p.tok == token.ASSIGN {
		stat.AssignToken = p.tokenNext()
		stat.ExpList = p.parseExpList()
	}
	return stat
}

func (p *parser) parseReturnStat() ast.Stat {
	stat := &ast.ReturnStat{}
	stat.ReturnToken = p.expectToken(token.RETURN)
	if p.isBlockFollow() || p.tok == token.SEMICOLON {
		return stat
	}
	stat.ExpList = p.parseExpList()
	return stat
}

func (p *parser) parseBreakStat() ast.Stat {
	stat := &ast.BreakStat{}
	stat.BreakToken = p.expectToken(token.BREAK)
	return stat
}

func (p *parser) parsePrefixExp() (exp ast.Exp) {
	switch p.tok {
	case token.LPAREN:
		e := &ast.ParenExp{}
		e.LParenToken = p.tokenNext()
		e.Exp = p.parseExp()
		e.RParenToken = p.expectToken(token.RPAREN)
		exp = e
	case token.NAME:
		e := &ast.VariableExp{}
		e.NameToken = p.parseName()
		exp = e
	default:
		p.error(p.off, "unexpected symbol")
	}
	return exp
}

func (p *parser) parseTableCtor() (ctor *ast.TableCtor) {
	ctor = &ast.TableCtor{}
	ctor.LBraceToken = p.expectToken(token.LBRACE)
	if p.tok != token.RBRACE {
		ctor.Fields = &ast.FieldList{}
	}
	for p.tok != token.RBRACE {
		var entry ast.Entry
		if p.tok == token.LBRACK {
			e := &ast.IndexEntry{}
			e.LBrackToken = p.tokenNext()
			e.KeyExp = p.parseExp()
			e.RBrackToken = p.expectToken(token.RBRACK)
			e.AssignToken = p.expectToken(token.ASSIGN)
			e.ValueExp = p.parseExp()
			entry = e
		} else if p.lookahead(); p.tok == token.NAME && p.look.tok == token.ASSIGN {
			e := &ast.FieldEntry{}
			e.Name = p.parseName()
			e.AssignToken = p.expectToken(token.ASSIGN)
			e.Value = p.parseExp()
			entry = e
		} else {
			e := &ast.ValueEntry{}
			e.Value = p.parseExp()
			entry = e
		}
		ctor.Fields.Entries = append(ctor.Fields.Entries, entry)
		if p.tok == token.COMMA || p.tok == token.SEMICOLON {
			ctor.Fields.Seps = append(ctor.Fields.Seps, p.tokenNext())
		} else {
			break
		}
	}
	ctor.RBraceToken = p.expectToken(token.RBRACE)
	return ctor
}

func (p *parser) parseFuncArgs() (args ast.CallArgs) {
	switch p.tok {
	case token.LPAREN:
		a := &ast.ArgsCall{}
		a.LParenToken = p.tokenNext()
		for p.tok != token.RPAREN {
			if a.ExpList == nil {
				a.ExpList = &ast.ExpList{}
			}
			a.ExpList.Exps = append(a.ExpList.Exps, p.parseExp())
			if p.tok == token.COMMA {
				a.ExpList.Seps = append(a.ExpList.Seps, p.tokenNext())
			} else {
				break
			}
		}
		a.RParenToken = p.expectToken(token.RPAREN)
		args = a
	case token.LBRACE:
		a := &ast.TableCall{}
		a.TableExp = p.parseTableCtor()
		args = a
	case token.STRING, token.LONGSTRING:
		a := &ast.StringCall{}
		a.StringExp = p.parseString()
		args = a
	default:
		p.error(p.off, "function arguments expected")
	}
	return args
}

func (p *parser) parsePrimaryExp() (exp ast.Exp) {
loop:
	for exp = p.parsePrefixExp(); ; {
		switch p.tok {
		case token.DOT:
			e := &ast.FieldExp{}
			e.Exp = exp
			e.DotToken = p.tokenNext()
			e.Field = p.parseName()
			exp = e
		case token.COLON:
			e := &ast.MethodExp{}
			e.Exp = exp
			e.ColonToken = p.tokenNext()
			e.Name = p.parseName()
			e.Args = p.parseFuncArgs()
			exp = e
		case token.LBRACK:
			e := &ast.IndexExp{}
			e.Exp = exp
			e.LBrackToken = p.tokenNext()
			e.Index = p.parseExp()
			e.RBrackToken = p.expectToken(token.RBRACK)
			exp = e
		case token.LBRACE, token.LPAREN:
			e := &ast.CallExp{}
			e.Exp = exp
			e.Args = p.parseFuncArgs()
			exp = e
		default:
			break loop
		}
	}
	return exp
}

func (p *parser) parseExpStat() ast.Stat {
	exp := p.parsePrimaryExp()
	switch exp.(type) {
	case *ast.MethodExp, *ast.CallExp:
		return &ast.CallExprStat{Exp: exp}
	}

	stat := &ast.AssignStat{Left: &ast.ExpList{Exps: []ast.Exp{exp}}}
	for p.tok == token.COMMA {
		stat.Left.Seps = append(stat.Left.Seps, p.tokenNext())
		switch exp := p.parsePrimaryExp().(type) {
		case *ast.MethodExp, *ast.CallExp:
			p.error(p.off, "syntax error")
		default:
			stat.Left.Exps = append(stat.Left.Exps, exp)
		}
	}
	stat.AssignToken = p.expectToken(token.ASSIGN)
	stat.Right = p.parseExpList()
	return stat
}

func (p *parser) parseStat() (stat ast.Stat, last bool) {
	switch p.tok {
	case token.DO:
		return p.parseDoStat(), false
	case token.WHILE:
		return p.parseWhileStat(), false
	case token.REPEAT:
		return p.parseRepeatStat(), false
	case token.IF:
		return p.parseIfStat(), false
	case token.FOR:
		return p.parseForStat(), false
	case token.FUNCTION:
		return p.parseFunction(funcStat), false
	case token.LOCAL:
		return p.parseLocalStat(), false
	case token.RETURN:
		return p.parseReturnStat(), true
	case token.BREAK:
		return p.parseBreakStat(), true
	}
	return p.parseExpStat(), false
}

func (p *parser) parseBlock() (block *ast.Block) {
	block = &ast.Block{}
	for last := false; !last && !p.isBlockFollow(); {
		var stat ast.Stat
		stat, last = p.parseStat()
		block.Stats = append(block.Stats, stat)
		var semi ast.Token
		if p.tok == token.SEMICOLON {
			semi = p.tokenNext()
		}
		block.Seps = append(block.Seps, semi)
	}
	return
}

func (p *parser) parseFile() *ast.File {
	return &ast.File{
		Name:  p.file.Name(),
		Block: p.parseBlock(),
	}
}

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
