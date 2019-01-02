package lexer

import (
	"strings"
	"testing"

	"github.com/geovanisouza92/geo/token"
)

func TestNextToken(t *testing.T) {
	input := `
fn let return true false
123 1.23 1.4e5
foo _foo f12 io! option? 1f
=+-*/==!!=>>=<<=|&&||
;,:(){}[]
"foobar" "foo bar" "foo \"bar"
[1] [1, 2]
{} {"foo": "bar"} {"foo": "bar", "baz": "goo"}
世界
`

	tt := []struct {
		Type    token.TokenType
		Literal string
		Line    int
		Col     int
	}{
		{token.Fn, "fn", 2, 3},
		{token.Let, "let", 2, 7},
		{token.Return, "return", 2, 14},
		{token.True, "true", 2, 19},
		{token.False, "false", 2, 25},
		{token.Number, "123", 3, 4},
		{token.Number, "1.23", 3, 9},
		{token.Number, "1.4e5", 3, 15},
		{token.Id, "foo", 4, 4},
		{token.Id, "_foo", 4, 9},
		{token.Id, "f12", 4, 13},
		{token.Id, "io!", 4, 17},
		{token.Id, "option?", 4, 25},
		{token.Number, "1", 4, 27},
		{token.Id, "f", 4, 28},
		{token.Assign, "=", 5, 2},
		{token.Plus, "+", 5, 3},
		{token.Minus, "-", 5, 4},
		{token.Mul, "*", 5, 5},
		{token.Div, "/", 5, 6},
		{token.Eq, "==", 5, 8},
		{token.Not, "!", 5, 9},
		{token.Neq, "!=", 5, 11},
		{token.Gt, ">", 5, 12},
		{token.Ge, ">=", 5, 14},
		{token.Lt, "<", 5, 15},
		{token.Le, "<=", 5, 17},
		{token.Pipe, "|", 5, 18},
		{token.And, "&&", 5, 20},
		{token.Or, "||", 5, 22},
		{token.EOL, ";", 6, 2},
		{token.Comma, ",", 6, 3},
		{token.Colon, ":", 6, 4},
		{token.LParen, "(", 6, 5},
		{token.RParen, ")", 6, 6},
		{token.LBrace, "{", 6, 7},
		{token.RBrace, "}", 6, 8},
		{token.LBracket, "[", 6, 9},
		{token.RBracket, "]", 6, 10},
		{token.String, "foobar", 7, 7},
		{token.String, "foo bar", 7, 17},
		{token.String, `foo \"bar`, 7, 29},
		{token.LBracket, "[", 8, 2},
		{token.Number, "1", 8, 3},
		{token.RBracket, "]", 8, 4},
		{token.LBracket, "[", 8, 6},
		{token.Number, "1", 8, 7},
		{token.Comma, ",", 8, 8},
		{token.Number, "2", 8, 10},
		{token.RBracket, "]", 8, 11},
		{token.LBrace, "{", 9, 2},
		{token.RBrace, "}", 9, 3},
		{token.LBrace, "{", 9, 5},
		{token.String, "foo", 9, 8},
		{token.Colon, ":", 9, 11},
		{token.String, "bar", 9, 15},
		{token.RBrace, "}", 9, 18},
		{token.LBrace, "{", 9, 20},
		{token.String, "foo", 9, 23},
		{token.Colon, ":", 9, 26},
		{token.String, "bar", 9, 30},
		{token.Comma, ",", 9, 33},
		{token.String, "baz", 9, 37},
		{token.Colon, ":", 9, 40},
		{token.String, "goo", 9, 44},
		{token.RBrace, "}", 9, 47},
		{token.Id, "世界", 10, 3},
		{token.EOF, "", 11, 1},
	}

	l := New(strings.NewReader(input))

	for _, tc := range tt {
		t.Run("Token_"+tc.Type.String(), func(t *testing.T) {
			tok := l.NextToken()
			if tok.Type != tc.Type {
				t.Errorf("token should be of type %q; got %q", tc.Type, tok.Type)
			}
			if tok.Literal != tc.Literal {
				t.Errorf("token should have the literal %q; got %q", tc.Literal, tok.Literal)
			}
			if tok.Line != tc.Line {
				t.Errorf("token should be at the line %d; got %d", tc.Line, tok.Line)
			}
			if tok.Col != tc.Col {
				t.Errorf("token should be at column %d; got %d", tc.Col, tok.Col)
			}
		})
	}
}
