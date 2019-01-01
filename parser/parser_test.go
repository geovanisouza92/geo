package parser

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/geovanisouza92/geo/ast"
	"github.com/geovanisouza92/geo/lexer"
)

func TestParse(t *testing.T) {
	Convey("Parse module", t, func() {
		input := `
		let x = 5;
		let y = true;
		let foobar = y;
		`

		m := testEval(input, 3)

		tests := []struct {
			id  string
			val interface{}
		}{
			{"x", 5},
			{"y", true},
			{"foobar", "y"},
		}

		for i, expected := range tests {
			Convey(fmt.Sprintf("test let statement #%d", i), func() {
				testLetStatement(t, m.Statements[i], expected.id, expected.val)
			})
		}
	})
}

func TestReturn(t *testing.T) {
	Convey("Return statement", t, func() {
		tests := []struct {
			input string
			val   interface{}
		}{
			{"return 5;", 5},
			{"return 10;", 10},
			{"return 993322;", 993322},
		}

		for _, expected := range tests {
			Convey(expected.input, func() {
				m := testEval(expected.input, 1)

				ret, ok := m.Statements[0].(*ast.ReturnStatement)
				So(ok, ShouldBeTrue)
				So(ret.TokenLiteral(), ShouldEqual, "return")
				testLiteral(t, ret.Value, expected.val)
			})
		}
	})
}

func TestExpressions(t *testing.T) {
	Convey("Id literal", t, func() {
		input := `foobar;`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)

		testLiteral(t, exp.Expression, "foobar")
	})
	Convey("Num literal", t, func() {
		input := `5;`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)

		testLiteral(t, exp.Expression, 5)
	})
	Convey("String literal", t, func() {
		input := `"Hello world";`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)

		testStringLiteral(exp.Expression, "Hello world")
	})
	Convey("Array literal", t, func() {
		input := `[1, 2 * 2, 3 + 3];`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)
		ary, ok := exp.Expression.(*ast.Array)
		So(ok, ShouldBeTrue)
		So(len(ary.Elements), ShouldEqual, 3)

		testNumberLiteral(ary.Elements[0], 1)
		testInfixExpression(t, ary.Elements[1], 2, "*", 2)
		testInfixExpression(t, ary.Elements[2], 3, "+", 3)
	})
	Convey("Hash literal", t, func() {
		tests := []struct {
			input  string
			length int
			val    interface{}
		}{
			{"{};", 0, map[string]float64{}},
			{`{"one": 1, "two": 2, "three": 3};`, 3, map[string]float64{"one": 1, "two": 2, "three": 3}},
			{`{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`, 3, map[string]func(ast.Expression){
				"one": func(e ast.Expression) {
					testInfixExpression(t, e, 0, "+", 1)
				},
				"two": func(e ast.Expression) {
					testInfixExpression(t, e, 10, "-", 8)
				},
				"three": func(e ast.Expression) {
					testInfixExpression(t, e, 15, "/", 5)
				},
			}},
		}

		for _, expected := range tests {
			Convey(expected.input, func() {
				m := testEval(expected.input, 1)

				exp, ok := m.Statements[0].(*ast.ExpressionStatement)
				So(ok, ShouldBeTrue)
				hash, ok := exp.Expression.(*ast.Hash)
				So(ok, ShouldBeTrue)
				So(len(hash.Pairs), ShouldEqual, expected.length)

				switch val := expected.val.(type) {
				case map[string]float64:

					i := 1
					for k, v := range hash.Pairs {
						Convey(fmt.Sprintf("testing %d", i), func() {
							key, ok := k.(*ast.String)
							So(ok, ShouldBeTrue)
							testNumberLiteral(v, val[key.Value])
						})
						i += 1
					}
				case map[string]func(ast.Expression):

					i := 1
					for k, v := range hash.Pairs {
						Convey(fmt.Sprintf("testing %d", i), func() {
							key, ok := k.(*ast.String)
							So(ok, ShouldBeTrue)
							val[key.Value](v)
						})
						i += 1
					}
				}
			})
		}
	})
	Convey("Prefix expressions", t, func() {
		tests := []struct {
			input string
			op    string
			val   interface{}
		}{
			{"!5;", "!", 5},
			{"-15;", "-", 15},
			{"!true;", "!", true},
			{"!false;", "!", false},
		}

		for _, expected := range tests {
			Convey(expected.input, func() {
				m := testEval(expected.input, 1)

				exp, ok := m.Statements[0].(*ast.ExpressionStatement)
				So(ok, ShouldBeTrue)

				prefix, ok := exp.Expression.(*ast.PrefixExpression)
				So(ok, ShouldBeTrue)
				So(prefix.Op, ShouldEqual, expected.op)
				testLiteral(t, prefix.Right, expected.val)
			})
		}
	})
	Convey("Infix expressions", t, func() {
		tests := []struct {
			input string
			left  interface{}
			op    string
			right interface{}
		}{
			{"5 + 5;", 5, "+", 5},
			{"5 - 5;", 5, "-", 5},
			{"5 * 5;", 5, "*", 5},
			{"5 / 5;", 5, "/", 5},
			{"5 == 5;", 5, "==", 5},
			{"5 != 5;", 5, "!=", 5},
			{"5 > 5;", 5, ">", 5},
			{"5 >= 5;", 5, ">=", 5},
			{"5 < 5;", 5, "<", 5},
			{"5 <= 5;", 5, "<=", 5},
			{"5 | 5;", 5, "|", 5},
			{"5 && 5;", 5, "&&", 5},
			{"5 || 5;", 5, "||", 5},
			{"true == true", true, "==", true},
			{"true != false", true, "!=", false},
			{"false == false", false, "==", false},
		}

		for _, expected := range tests {
			Convey(expected.input, func() {
				m := testEval(expected.input, 1)

				exp, ok := m.Statements[0].(*ast.ExpressionStatement)
				So(ok, ShouldBeTrue)

				testInfixExpression(t, exp.Expression, expected.left, expected.op, expected.right)
			})
		}
	})
	Convey("Index expression", t, func() {
		input := `myArray[1 + 1];`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)
		idx, ok := exp.Expression.(*ast.Index)

		testIdLiteral(idx.Left, "myArray")
		testInfixExpression(t, idx.Index, 1, "+", 1)
	})
	Convey("Function expressions", t, func() {
		input := `fn(x, y) { x + y; }`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)

		fn, ok := exp.Expression.(*ast.Fn)
		So(ok, ShouldBeTrue)
		testLiteral(t, fn.Params[0], "x")
		testLiteral(t, fn.Params[1], "y")

		So(len(fn.Body.Statements), ShouldEqual, 1)
		body, ok := fn.Body.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)
		testInfixExpression(t, body.Expression, "x", "+", "y")
	})
	Convey("Function params", t, func() {
		tests := []struct {
			input  string
			params []string
		}{
			{"fn(){}", []string{}},
			{"fn(x){}", []string{"x"}},
			{"fn(x,y){}", []string{"x", "y"}},
		}

		for _, expected := range tests {
			Convey(expected.input, func() {
				m := testEval(expected.input, 1)

				exp, ok := m.Statements[0].(*ast.ExpressionStatement)
				So(ok, ShouldBeTrue)

				fn, ok := exp.Expression.(*ast.Fn)
				So(ok, ShouldBeTrue)
				So(len(fn.Params), ShouldEqual, len(expected.params))
				for i, p := range expected.params {
					testLiteral(t, fn.Params[i], p)
				}
			})
		}
	})
	Convey("Call expressions", t, func() {
		input := `add(1, 2 * 3, 4 + 5)`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)

		call, ok := exp.Expression.(*ast.Call)
		So(ok, ShouldBeTrue)

		testIdLiteral(call.Fn, "add")

		So(len(call.Args), ShouldEqual, 3)
		testLiteral(t, call.Args[0], 1)
		testInfixExpression(t, call.Args[1], 2, "*", 3)
		testInfixExpression(t, call.Args[2], 4, "+", 5)
	})
}

func TestIfExpressions(t *testing.T) {
	Convey("If expressions", t, func() {
		input := `if (x < y) { x }`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)

		if_, ok := exp.Expression.(*ast.IfExpression)
		So(ok, ShouldBeTrue)

		testInfixExpression(t, if_.Condition, "x", "<", "y")

		So(len(if_.Consequence.Statements), ShouldEqual, 1)
		cons, ok := if_.Consequence.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)

		testIdLiteral(cons.Expression, "x")

		So(if_.Alternative, ShouldBeNil)
	})
	Convey("If/else expressions", t, func() {
		input := `if (x < y) { x } else { y }`
		m := testEval(input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)

		if_, ok := exp.Expression.(*ast.IfExpression)
		So(ok, ShouldBeTrue)

		testInfixExpression(t, if_.Condition, "x", "<", "y")

		So(len(if_.Consequence.Statements), ShouldEqual, 1)
		cons, ok := if_.Consequence.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)
		testIdLiteral(cons.Expression, "x")

		So(len(if_.Alternative.Statements), ShouldEqual, 1)
		alt, ok := if_.Alternative.Statements[0].(*ast.ExpressionStatement)
		So(ok, ShouldBeTrue)
		testIdLiteral(alt.Expression, "y")
	})
}

func TestOperatorPrecedence(t *testing.T) {
	Convey("operator precedence", t, func() {
		tests := []struct {
			input  string
			output string
			length int
		}{
			{"-a * b", "((-a) * b)", 1},
			{"!-a", "(!(-a))", 1},
			{"a + b + c", "((a + b) + c)", 1},
			{"a + b - c", "((a + b) - c)", 1},
			{"a * b * c", "((a * b) * c)", 1},
			{"a * b / c", "((a * b) / c)", 1},
			{"a + b / c", "(a + (b / c))", 1},
			{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)", 1},
			{"3 + 4; -5 * 5", "(3 + 4)((-5) * 5)", 2},
			{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))", 1},
			{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))", 1},
			{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))", 1},
			{"3 + 4 * 5 == 3 * 1 + 4 * 7", "((3 + (4 * 5)) == ((3 * 1) + (4 * 7)))", 1},
			{"true", "true", 1},
			{"false", "false", 1},
			{"3 > 5 == false", "((3 > 5) == false)", 1},
			{"3 < 5 == true", "((3 < 5) == true)", 1},
			{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)", 1},
			{"(5 + 5) * 2", "((5 + 5) * 2)", 1},
			{"2.3 / (5.4 + 5.7)", "(2.3 / (5.4 + 5.7))", 1},
			{"-(5 + 5)", "(-(5 + 5))", 1},
			{"!(true == true)", "(!(true == true))", 1},
			{"a + add(b * c) + d", "((a + add((b * c))) + d)", 1},
			{"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))", "add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))", 1},
			{"add(a + b + c * d / f + g)", "add((((a + b) + ((c * d) / f)) + g))", 1},
			{"a * [1, 2, 3, 4][b * c] * d", "((a * ([1, 2, 3, 4][(b * c)])) * d)", 1},
			{"add(a * b[2], b[1], 2 * [1, 2][1])", "add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))", 1},
		}

		for _, expected := range tests {
			Convey(expected.input, func() {
				m := testEval(expected.input, expected.length)

				actual := m.String()
				So(actual, ShouldEqual, expected.output)
			})
		}
	})
}

func testLetStatement(t *testing.T, s ast.Statement, name string, val interface{}) {
	So(s.TokenLiteral(), ShouldEqual, "let")
	let, ok := s.(*ast.LetStatement)
	So(ok, ShouldBeTrue)
	So(let.Name.TokenLiteral(), ShouldEqual, name)
	testLiteral(t, let.Value, val)
}

func testLiteral(t *testing.T, exp ast.Expression, val interface{}) {
	switch val := val.(type) {
	case string:
		testIdLiteral(exp, val)
	case int:
		testNumberLiteral(exp, float64(val))
	case float64:
		testNumberLiteral(exp, val)
	case bool:
		testBoolLiteral(exp, val)
	default:
		t.Errorf("type of exp not handled: %T", val)
	}
}

func testIdLiteral(exp ast.Expression, val string) {
	id, ok := exp.(*ast.Id)
	So(ok, ShouldBeTrue)
	So(id.Value, ShouldEqual, val)
	So(id.TokenLiteral(), ShouldEqual, val)
}

func testNumberLiteral(exp ast.Expression, val float64) {
	num, ok := exp.(*ast.Number)
	So(ok, ShouldBeTrue)
	So(num.Value, ShouldEqual, val)
	So(num.TokenLiteral(), ShouldEqual, fmt.Sprintf("%v", val))
}

func testStringLiteral(exp ast.Expression, val string) {
	str, ok := exp.(*ast.String)
	So(ok, ShouldBeTrue)
	So(str.Value, ShouldEqual, val)
	So(str.TokenLiteral(), ShouldEqual, val)
}

func testBoolLiteral(exp ast.Expression, val bool) {
	b, ok := exp.(*ast.Bool)
	So(ok, ShouldBeTrue)
	So(b.Value, ShouldEqual, val)
	So(b.TokenLiteral(), ShouldEqual, fmt.Sprintf("%t", val))
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, op string, right interface{}) {
	infix, ok := exp.(*ast.InfixExpression)
	So(ok, ShouldBeTrue)
	testLiteral(t, infix.Left, left)
	So(infix.Op, ShouldEqual, op)
	testLiteral(t, infix.Right, right)
}

func testEval(input string, expectedStatementsLen int) *ast.Module {
	l := lexer.New(strings.NewReader(input))
	p := New(l)
	m, errors := p.Parse()
	So(errors, ShouldBeEmpty)
	if expectedStatementsLen > 0 {
		So(len(m.Statements), ShouldEqual, expectedStatementsLen)
	}
	return m
}
