package parser

import (
	"fmt"
	"strconv"
	"strings"

	"geo/ast"
	"geo/lexer"
	"geo/token"
)

const (
	_ byte = iota
	Lowest
	Pipe       // |
	Logical    // && ||
	Equality   // == !=
	Relational // > >= < <=
	Sum        // + -
	Product    // * /
	Prefix     // -x !x
	Call       // a(b)
	Index      // a[b]
)

var precedences = map[token.TokenType]byte{
	token.Pipe:     Pipe,
	token.And:      Logical,
	token.Or:       Logical,
	token.Eq:       Equality,
	token.Neq:      Equality,
	token.Gt:       Relational,
	token.Ge:       Relational,
	token.Lt:       Relational,
	token.Le:       Relational,
	token.Plus:     Sum,
	token.Minus:    Sum,
	token.Mul:      Product,
	token.Div:      Product,
	token.LParen:   Call,
	token.LBracket: Index,
}

type prefixParseFn func() ast.Expression
type infixParseFn func(ast.Expression) ast.Expression

type parseErrors []error

func (self parseErrors) String() string {
	lines := []string{}
	for _, e := range self {
		lines = append(lines, "- "+e.Error())
	}
	return "\n" + strings.Join(lines, "\n") + "\n"
}

type Parser struct {
	l *lexer.Lexer

	curr token.Token
	next token.Token

	errors parseErrors

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         parseErrors{},
		prefixParseFns: map[token.TokenType]prefixParseFn{},
		infixParseFns:  map[token.TokenType]infixParseFn{},
	}
	p.nextToken()
	p.nextToken()

	p.prefixParseFns[token.Id] = p.parseId
	p.prefixParseFns[token.Number] = p.parseNumber
	p.prefixParseFns[token.String] = p.parseString
	p.prefixParseFns[token.LBracket] = p.parseArray
	p.prefixParseFns[token.LBrace] = p.parseHash
	p.prefixParseFns[token.True] = p.parseBool
	p.prefixParseFns[token.False] = p.parseBool
	p.prefixParseFns[token.Not] = p.parsePrefixExpression
	p.prefixParseFns[token.Minus] = p.parsePrefixExpression
	p.prefixParseFns[token.LParen] = p.parseGroupedExpression
	p.prefixParseFns[token.If] = p.parseIfExpression
	p.prefixParseFns[token.Fn] = p.parseFnExpression

	p.infixParseFns[token.Plus] = p.parseInfixExpression
	p.infixParseFns[token.Minus] = p.parseInfixExpression
	p.infixParseFns[token.Mul] = p.parseInfixExpression
	p.infixParseFns[token.Div] = p.parseInfixExpression
	p.infixParseFns[token.Eq] = p.parseInfixExpression
	p.infixParseFns[token.Neq] = p.parseInfixExpression
	p.infixParseFns[token.Gt] = p.parseInfixExpression
	p.infixParseFns[token.Ge] = p.parseInfixExpression
	p.infixParseFns[token.Lt] = p.parseInfixExpression
	p.infixParseFns[token.Le] = p.parseInfixExpression
	p.infixParseFns[token.And] = p.parseInfixExpression
	p.infixParseFns[token.Or] = p.parseInfixExpression
	p.infixParseFns[token.Pipe] = p.parseInfixExpression
	p.infixParseFns[token.LParen] = p.parseCallExpression
	p.infixParseFns[token.LBracket] = p.parseIndexExpression

	return p
}

func (p *Parser) Parse() (*ast.Module, parseErrors) {
	m := &ast.Module{Statements: []ast.Statement{}}

	for p.curr.Type != token.EOF {
		if s := p.parseStatement(); s != nil {
			m.Statements = append(m.Statements, s)
		}
		p.nextToken()
	}

	return m, p.errors
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curr.Type {
	case token.Let:
		return p.parseLetStatement()
	case token.Return:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	s := &ast.LetStatement{Token: p.curr}

	if !p.assertNextIs(token.Id) {
		return nil
	}

	s.Name = &ast.Id{Token: p.curr, Value: p.curr.Literal}

	if !p.assertNextIs(token.Assign) {
		return nil
	}

	p.nextToken()
	s.Value = p.parseExpression(Lowest)
	if p.next.Type == token.EOL {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	s := &ast.ReturnStatement{Token: p.curr}
	p.nextToken()

	s.Value = p.parseExpression(Lowest)
	if p.next.Type == token.EOL {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	s := &ast.ExpressionStatement{Token: p.curr}
	s.Expression = p.parseExpression(Lowest)

	if p.next.Type == token.EOL {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseExpression(precedence byte) ast.Expression {
	prefix := p.prefixParseFns[p.curr.Type]
	if prefix == nil {
		p.addError("no prefix func for %s", p.curr.Type)
		return nil
	}
	left := prefix()

	for p.next.Type != token.EOL && precedence < p.nextPrecedence() {
		infix := p.infixParseFns[p.next.Type]
		if infix == nil {
			return left
		}
		p.nextToken()
		left = infix(left)
	}

	return left
}

func (p *Parser) parseId() ast.Expression {
	return &ast.Id{p.curr, p.curr.Literal}
}

func (p *Parser) parseNumber() ast.Expression {
	v, err := strconv.ParseFloat(p.curr.Literal, 64)
	if err != nil {
		p.addError("could not parse %q as number", p.curr.Literal)
		return nil
	}
	return &ast.Number{p.curr, v}
}

func (p *Parser) parseString() ast.Expression {
	return &ast.String{p.curr, p.curr.Literal}
}

func (p *Parser) parseBool() ast.Expression {
	return &ast.Bool{p.curr, p.curr.Type == token.True}
}

func (p *Parser) parseArray() ast.Expression {
	ary := &ast.Array{Token: p.curr}
	ary.Elements = p.parseExpressionList(token.RBracket)
	return ary
}

func (p *Parser) parseHash() ast.Expression {
	hash := &ast.Hash{Token: p.curr}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for p.next.Type != token.RBrace {
		p.nextToken()
		key := p.parseExpression(Lowest)

		if !p.assertNextIs(token.Colon) {
			return nil
		}
		p.nextToken()

		hash.Pairs[key] = p.parseExpression(Lowest)

		if p.next.Type != token.RBrace && !p.assertNextIs(token.Comma) {
			return nil
		}
	}

	if !p.assertNextIs(token.RBrace) {
		return nil
	}

	return hash
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	e := &ast.PrefixExpression{
		Token: p.curr,
		Op:    p.curr.Literal,
	}
	p.nextToken()

	e.Right = p.parseExpression(Prefix)

	return e
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	e := &ast.InfixExpression{
		Token: p.curr,
		Left:  left,
		Op:    p.curr.Literal,
	}

	precedence := p.currPrecedence()
	p.nextToken()
	e.Right = p.parseExpression(precedence)

	return e
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	e := p.parseExpression(Lowest)
	if !p.assertNextIs(token.RParen) {
		return nil
	}
	return e
}

func (p *Parser) parseIfExpression() ast.Expression {
	e := &ast.IfExpression{Token: p.curr}

	if !p.assertNextIs(token.LParen) {
		return nil
	}

	p.nextToken()
	e.Condition = p.parseExpression(Lowest)

	if !p.assertNextIs(token.RParen) {
		return nil
	}

	if !p.assertNextIs(token.LBrace) {
		return nil
	}

	e.Consequence = p.parseBlockStatement()

	if p.next.Type == token.Else {
		p.nextToken()

		if !p.assertNextIs(token.LBrace) {
			return nil
		}

		e.Alternative = p.parseBlockStatement()
	}

	return e
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	b := &ast.BlockStatement{Token: p.curr, Statements: []ast.Statement{}}
	p.nextToken()

	for p.curr.Type != token.RBrace && p.curr.Type != token.EOF {
		if s := p.parseStatement(); s != nil {
			b.Statements = append(b.Statements, s)
		}
		p.nextToken()
	}

	return b
}

func (p *Parser) parseFnExpression() ast.Expression {
	e := &ast.Fn{Token: p.curr}

	if !p.assertNextIs(token.LParen) {
		return nil
	}

	e.Params = p.parseFnParams()

	if !p.assertNextIs(token.LBrace) {
		return nil
	}

	e.Body = p.parseBlockStatement()

	return e
}

func (p *Parser) parseFnParams() []*ast.Id {
	ids := []*ast.Id{}

	// no params
	if p.next.Type == token.RParen {
		p.nextToken()
		return ids
	}
	p.nextToken()

	// one param
	id := &ast.Id{Token: p.curr, Value: p.curr.Literal}
	ids = append(ids, id)

	// multiple params
	for p.next.Type == token.Comma {
		p.nextToken()
		p.nextToken()
		id := &ast.Id{Token: p.curr, Value: p.curr.Literal}
		ids = append(ids, id)
	}

	if !p.assertNextIs(token.RParen) {
		return nil
	}

	return ids
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	e := &ast.Call{Token: p.curr, Fn: left}
	e.Args = p.parseExpressionList(token.RParen)
	return e
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	args := []ast.Expression{}

	// no args
	if p.next.Type == end {
		p.nextToken()
		return args
	}
	p.nextToken()

	// one arg
	args = append(args, p.parseExpression(Lowest))

	// multiple args
	for p.next.Type == token.Comma {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(Lowest))
	}

	if !p.assertNextIs(end) {
		return nil
	}

	return args
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	e := &ast.Index{Token: p.curr, Left: left}

	p.nextToken()

	e.Index = p.parseExpression(Lowest)

	if !p.assertNextIs(token.RBracket) {
		return nil
	}

	return e
}

func (p *Parser) assertNextIs(t token.TokenType) bool {
	if p.next.Type == t {
		p.nextToken()
		return true
	}
	p.addError("expected next token to be %s, got %s instead", t, p.next.Type)
	return false
}

func (p *Parser) nextToken() {
	p.curr = p.next
	p.next = p.l.NextToken()
}

func (p *Parser) currPrecedence() byte {
	if p, ok := precedences[p.curr.Type]; ok {
		return p
	}
	return Lowest
}

func (p *Parser) nextPrecedence() byte {
	if p, ok := precedences[p.next.Type]; ok {
		return p
	}
	return Lowest
}

func (p *Parser) addError(msg string, args ...interface{}) {
	err := fmt.Errorf("at line %d, column %d: %s", p.curr.Line, p.curr.Col, fmt.Sprintf(msg, args...))
	p.errors = append(p.errors, err)
}
