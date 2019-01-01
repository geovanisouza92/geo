package lexer

import (
	"io"
	"text/scanner"

	"github.com/geovanisouza92/geo/token"
)

type Lexer struct {
	s scanner.Scanner

	curr rune
}

func New(in io.Reader) *Lexer {
	var s scanner.Scanner
	s.Init(in)
	l := &Lexer{s: s}
	l.readRune()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var t token.Token

	switch l.curr {
	case '=':
		t = l.either('=', token.Eq, token.Assign)
	case '+':
		t = l.token(token.Plus)
	case '-':
		t = l.token(token.Minus)
	case '*':
		t = l.token(token.Mul)
	case '/':
		t = l.token(token.Div)
	case '!':
		t = l.either('=', token.Neq, token.Not)
	case '>':
		t = l.either('=', token.Ge, token.Gt)
	case '<':
		t = l.either('=', token.Le, token.Lt)
	case '&':
		t = l.either('&', token.And, token.Error)
	case '|':
		t = l.either('|', token.Or, token.Pipe)
	case ';':
		t = l.token(token.EOL)
	case ',':
		t = l.token(token.Comma)
	case ':':
		t = l.token(token.Colon)
	case '(':
		t = l.token(token.LParen)
	case ')':
		t = l.token(token.RParen)
	case '[':
		t = l.token(token.LBracket)
	case ']':
		t = l.token(token.RBracket)
	case '{':
		t = l.token(token.LBrace)
	case '}':
		t = l.token(token.RBrace)
	case scanner.Ident:
		p := l.s.Pos()
		lit := l.s.TokenText()
		col := p.Column
		if la := l.s.Peek(); la == '?' || la == '!' {
			l.readRune()
			lit += l.s.TokenText()
			col += 1
		}
		t = token.Token{
			Type:    token.LookupId(lit),
			Literal: lit,
			Line:    p.Line,
			Col:     col,
		}
	case scanner.Int, scanner.Float:
		p := l.s.Pos()
		lit := l.s.TokenText()
		t = token.Token{
			Type:    token.Number,
			Literal: lit,
			Line:    p.Line,
			Col:     p.Column,
		}
	case scanner.String:
		p := l.s.Pos()
		lit := l.s.TokenText()
		t = token.Token{
			Type:    token.String,
			Literal: lit[1 : len(lit)-1],
			Line:    p.Line,
			Col:     p.Column - 2,
		}
	case scanner.EOF:
		p := l.s.Pos()
		t = token.Token{Type: token.EOF, Literal: "", Line: p.Line, Col: p.Column}
	default:
		p := l.s.Pos()
		lit := l.s.TokenText()
		t = token.Token{Type: token.Error, Literal: lit, Line: p.Line, Col: p.Column}
	}

	l.readRune()
	return t
}

func (l *Lexer) readRune() {
	l.curr = l.s.Scan()
}

func (l *Lexer) token(ty token.TokenType) token.Token {
	p := l.s.Pos()
	lit := l.s.TokenText()
	return token.Token{Type: ty, Literal: lit, Line: p.Line, Col: p.Column}
}

func (l *Lexer) either(lookAhead rune, option, alternative token.TokenType) token.Token {
	p := l.s.Pos()
	lit := l.s.TokenText()
	col := p.Column
	if l.s.Peek() == lookAhead {
		l.readRune()
		lit += l.s.TokenText()
		col += 1
		return token.Token{Type: option, Literal: lit, Line: p.Line, Col: col}
	} else {
		return token.Token{Type: alternative, Literal: lit, Line: p.Line, Col: p.Column}
	}
}
