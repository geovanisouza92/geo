package ast

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"geo/token"
)

func TestString(t *testing.T) {
	Convey("String should write the module code", t, func() {
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
		So(m.String(), ShouldEqual, "let myVar = anotherVar;")
	})
}
