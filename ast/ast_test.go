package ast

import (
	"testing"

	"github.com/geovanisouza92/geo/token"
)

func TestString(t *testing.T) {
	m := &Module{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{token.Let, "let", 0, 0},
				Name: &Id{
					Token: token.Token{token.Id, "myVar", 0, 0},
					Value: "myVar",
				},
				Value: &Id{
					Token: token.Token{token.Id, "anotherVar", 0, 0},
					Value: "anotherVar",
				},
			},
		},
	}

	if m.String() != "let myVar = anotherVar;" {
		t.Errorf("string should dump module code; got %q", m.String())
	}
}
