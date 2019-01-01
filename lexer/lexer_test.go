package lexer

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/geovanisouza92/geo/token"
)

func TestNextToken(t *testing.T) {
	Convey("Lexer interpret tokens", t, func() {
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

		tests := []struct {
			Type    token.TokenType
			Literal string
			Line    int
			Col     int
		}{
			{token.Fn, "fn", 2, 5},
			{token.Let, "let", 2, 9},
			{token.Return, "return", 2, 16},
			{token.True, "true", 2, 21},
			{token.False, "false", 2, 27},
			{token.Number, "123", 3, 6},
			{token.Number, "1.23", 3, 11},
			{token.Number, "1.4e5", 3, 17},
			{token.Id, "foo", 4, 6},
			{token.Id, "_foo", 4, 11},
			{token.Id, "f12", 4, 15},
			{token.Id, "io!", 4, 19},
			{token.Id, "option?", 4, 27},
			{token.Number, "1", 4, 29},
			{token.Id, "f", 4, 30},
			{token.Assign, "=", 5, 4},
			{token.Plus, "+", 5, 5},
			{token.Minus, "-", 5, 6},
			{token.Mul, "*", 5, 7},
			{token.Div, "/", 5, 8},
			{token.Eq, "==", 5, 10},
			{token.Not, "!", 5, 11},
			{token.Neq, "!=", 5, 13},
			{token.Gt, ">", 5, 14},
			{token.Ge, ">=", 5, 16},
			{token.Lt, "<", 5, 17},
			{token.Le, "<=", 5, 19},
			{token.Pipe, "|", 5, 20},
			{token.And, "&&", 5, 22},
			{token.Or, "||", 5, 24},
			{token.EOL, ";", 6, 4},
			{token.Comma, ",", 6, 5},
			{token.Colon, ":", 6, 6},
			{token.LParen, "(", 6, 7},
			{token.RParen, ")", 6, 8},
			{token.LBrace, "{", 6, 9},
			{token.RBrace, "}", 6, 10},
			{token.LBracket, "[", 6, 11},
			{token.RBracket, "]", 6, 12},
			{token.String, "foobar", 7, 9},
			{token.String, "foo bar", 7, 19},
			{token.String, `foo \"bar`, 7, 31},
			{token.LBracket, "[", 8, 4},
			{token.Number, "1", 8, 5},
			{token.RBracket, "]", 8, 6},
			{token.LBracket, "[", 8, 8},
			{token.Number, "1", 8, 9},
			{token.Comma, ",", 8, 10},
			{token.Number, "2", 8, 12},
			{token.RBracket, "]", 8, 13},
			{token.LBrace, "{", 9, 4},
			{token.RBrace, "}", 9, 5},
			{token.LBrace, "{", 9, 7},
			{token.String, "foo", 9, 10},
			{token.Colon, ":", 9, 13},
			{token.String, "bar", 9, 17},
			{token.RBrace, "}", 9, 20},
			{token.LBrace, "{", 9, 22},
			{token.String, "foo", 9, 25},
			{token.Colon, ":", 9, 28},
			{token.String, "bar", 9, 32},
			{token.Comma, ",", 9, 35},
			{token.String, "baz", 9, 39},
			{token.Colon, ":", 9, 42},
			{token.String, "goo", 9, 46},
			{token.RBrace, "}", 9, 49},
			{token.Id, "世界", 10, 5},
			{token.EOF, "", 11, 3},
		}

		l := New(strings.NewReader(input))

		for _, expected := range tests {
			tok := l.NextToken()
			Convey(fmt.Sprintf("Token: %v", tok), func() {
				So(tok.Type, ShouldEqual, expected.Type)
				So(tok.Literal, ShouldEqual, expected.Literal)
				So(tok.Line, ShouldEqual, expected.Line)
				So(tok.Col, ShouldEqual, expected.Col)
			})
		}
	})
}
