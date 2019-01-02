package parser

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/geovanisouza92/geo/ast"
	"github.com/geovanisouza92/geo/lexer"
)

func TestParse(t *testing.T) {
	input := `
let x = 5;
let y = true;
let foobar = y;
`

	m := assertEval(t, input, 3)

	tt := []struct {
		id  string
		val interface{}
	}{
		{"x", 5},
		{"y", true},
		{"foobar", "y"},
	}

	for i, tc := range tt {
		t.Run("test let statement #"+strconv.Itoa(i), func(t *testing.T) {
			testLetStatement(t, m.Statements[i], tc.id, tc.val)
		})
	}
}

func TestReturn(t *testing.T) {
	tt := []struct {
		input string
		val   interface{}
	}{
		{"return 5;", 5},
		{"return 10;", 10},
		{"return 993322;", 993322},
	}

	for _, tc := range tt {
		t.Run(tc.input, func(t *testing.T) {
			m := assertEval(t, tc.input, 1)

			ret, ok := m.Statements[0].(*ast.ReturnStatement)
			if !ok {
				t.Errorf("statement should be *ast.ReturnStatement; got %T", ret)
			}
			if ret.TokenLiteral() != "return" {
				t.Errorf(`return literal should be "return"; got %q`, ret.TokenLiteral())
			}
			testLiteral(t, ret.Value, tc.val)
		})
	}
}

func TestExpressions(t *testing.T) {
	t.Run("Id literal", func(t *testing.T) {
		input := `foobar;`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}

		testLiteral(t, exp.Expression, "foobar")
	})

	t.Run("Num literal", func(t *testing.T) {
		input := `5;`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}

		testLiteral(t, exp.Expression, 5)
	})

	t.Run("String literal", func(t *testing.T) {
		input := `"Hello world";`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}

		testStringLiteral(t, exp.Expression, "Hello world")
	})

	t.Run("Array literal", func(t *testing.T) {
		input := `[1, 2 * 2, 3 + 3];`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}
		ary, ok := exp.Expression.(*ast.Array)
		if !ok {
			t.Errorf("expression should be *ast.Array; got %T", m.Statements[0])
		}
		if len(ary.Elements) != 3 {
			t.Errorf("array should have 3 elements; got %d", len(ary.Elements))
		}

		testNumberLiteral(t, ary.Elements[0], 1)
		testInfixExpression(t, ary.Elements[1], 2, "*", 2)
		testInfixExpression(t, ary.Elements[2], 3, "+", 3)
	})

	t.Run("Hash literal", func(t *testing.T) {
		tt := []struct {
			input  string
			length int
			val    interface{}
		}{
			{"{};", 0, map[string]float64{}},
			{`{"one": 1, "two": 2, "three": 3};`, 3, map[string]float64{"one": 1, "two": 2, "three": 3}},
			{`{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`, 3, map[string]func(ast.Expression){
				"one":   func(e ast.Expression) { testInfixExpression(t, e, 0, "+", 1) },
				"two":   func(e ast.Expression) { testInfixExpression(t, e, 10, "-", 8) },
				"three": func(e ast.Expression) { testInfixExpression(t, e, 15, "/", 5) },
			}},
		}

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				m := assertEval(t, tc.input, 1)

				exp, ok := m.Statements[0].(*ast.ExpressionStatement)
				if !ok {
					t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
				}
				hash, ok := exp.Expression.(*ast.Hash)
				if !ok {
					t.Errorf("expression should be *ast.Hash; got %T", exp.Expression)
				}
				if len(hash.Pairs) != tc.length {
					t.Errorf("hash should have %d pairs; got %d", tc.length, len(hash.Pairs))
				}

				switch val := tc.val.(type) {
				case map[string]float64:

					i := 1
					for k, v := range hash.Pairs {
						t.Run(fmt.Sprintf("testing %d", i), func(t *testing.T) {
							key, ok := k.(*ast.String)
							if !ok {
								t.Errorf("key should be *ast.String; got %T", k)
							}
							testNumberLiteral(t, v, val[key.Value])
						})
						i++
					}
				case map[string]func(ast.Expression):

					i := 1
					for k, v := range hash.Pairs {
						t.Run(fmt.Sprintf("testing %d", i), func(t *testing.T) {
							key, ok := k.(*ast.String)
							if !ok {
								t.Errorf("key should be *ast.String; got %T", k)
							}
							val[key.Value](v)
						})
						i++
					}
				}
			})
		}
	})

	t.Run("Prefix expressions", func(t *testing.T) {
		tt := []struct {
			input string
			op    string
			val   interface{}
		}{
			{"!5;", "!", 5},
			{"-15;", "-", 15},
			{"!true;", "!", true},
			{"!false;", "!", false},
		}

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				m := assertEval(t, tc.input, 1)

				exp, ok := m.Statements[0].(*ast.ExpressionStatement)
				if !ok {
					t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
				}

				prefix, ok := exp.Expression.(*ast.PrefixExpression)
				if !ok {
					t.Errorf("prefix expression should be *ast.PrefixExpression; got %T", exp.Expression)
				}
				if prefix.Op != tc.op {
					t.Errorf("prefix expression should have operator %q; got %q", tc.op, prefix.Op)
				}
				testLiteral(t, prefix.Right, tc.val)
			})
		}
	})

	t.Run("Infix expressions", func(t *testing.T) {
		tt := []struct {
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

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				m := assertEval(t, tc.input, 1)

				exp, ok := m.Statements[0].(*ast.ExpressionStatement)
				if !ok {
					t.Errorf("expression should be *ast.ExpressionStatement; got %T", m.Statements[0])
				}

				testInfixExpression(t, exp.Expression, tc.left, tc.op, tc.right)
			})
		}
	})

	t.Run("Index expression", func(t *testing.T) {
		input := `myArray[1 + 1];`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}
		idx, ok := exp.Expression.(*ast.Index)
		if !ok {
			t.Errorf("index expression should be *ast.Index; got %T", exp.Expression)
		}

		testIdLiteral(t, idx.Left, "myArray")
		testInfixExpression(t, idx.Index, 1, "+", 1)
	})

	t.Run("Function expressions", func(t *testing.T) {
		input := `fn(x, y) { x + y; }`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}

		fn, ok := exp.Expression.(*ast.Fn)
		if !ok {
			t.Errorf("expression should be *ast.Fn; got %T", exp.Expression)
		}
		testLiteral(t, fn.Params[0], "x")
		testLiteral(t, fn.Params[1], "y")

		if len(fn.Body.Statements) != 1 {
			t.Errorf("function body should have 1 statement; got %d", len(fn.Body.Statements))
		}
		body, ok := fn.Body.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("function body should be *ast.ExpressionStatement; got %T", fn.Body.Statements[0])
		}
		testInfixExpression(t, body.Expression, "x", "+", "y")
	})

	t.Run("Function params", func(t *testing.T) {
		tt := []struct {
			input  string
			params []string
		}{
			{"fn(){}", []string{}},
			{"fn(x){}", []string{"x"}},
			{"fn(x,y){}", []string{"x", "y"}},
		}

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				m := assertEval(t, tc.input, 1)

				exp, ok := m.Statements[0].(*ast.ExpressionStatement)
				if !ok {
					t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
				}

				fn, ok := exp.Expression.(*ast.Fn)
				if !ok {
					t.Errorf("expression should be *ast.Fn; got %T", exp.Expression)
				}
				if len(fn.Params) != len(tc.params) {
					t.Errorf("function should have %d params; got %d", len(tc.params), len(fn.Params))
				}
				for i, p := range tc.params {
					testLiteral(t, fn.Params[i], p)
				}
			})
		}
	})

	t.Run("Call expressions", func(t *testing.T) {
		input := `add(1, 2 * 3, 4 + 5)`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}

		call, ok := exp.Expression.(*ast.Call)
		if !ok {
			t.Errorf("expression should be *ast.Call; got %T", exp.Expression)
		}

		testIdLiteral(t, call.Fn, "add")

		if len(call.Args) != 3 {
			t.Errorf("call expression should have 3 arguments; got %d", len(call.Args))
		}
		testLiteral(t, call.Args[0], 1)
		testInfixExpression(t, call.Args[1], 2, "*", 3)
		testInfixExpression(t, call.Args[2], 4, "+", 5)
	})
}

func TestIfExpressions(t *testing.T) {
	t.Run("If expressions", func(t *testing.T) {
		input := `if (x < y) { x }`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}

		cond, ok := exp.Expression.(*ast.IfExpression)
		if !ok {
			t.Errorf("expression should be *ast.IfExpression; got %T", exp.Expression)
		}

		testInfixExpression(t, cond.Condition, "x", "<", "y")

		if len(cond.Consequence.Statements) != 1 {
			t.Errorf("if consequence branch should have 1 statement; got %d", len(cond.Consequence.Statements))
		}
		cons, ok := cond.Consequence.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("if consequence's branch statement should be *ast.ExpressionStatement; got %T", cond.Consequence.Statements[0])
		}

		testIdLiteral(t, cons.Expression, "x")

		if cond.Alternative != nil {
			t.Errorf("if alternative branch should be nil; got %v", cond.Alternative)
		}
	})

	t.Run("If/else expressions", func(t *testing.T) {
		input := `if (x < y) { x } else { y }`
		m := assertEval(t, input, 1)

		exp, ok := m.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement should be *ast.ExpressionStatement; got %T", m.Statements[0])
		}

		cond, ok := exp.Expression.(*ast.IfExpression)
		if !ok {
			t.Errorf("expression should be *ast.IfExpression; got %T", exp.Expression)
		}

		testInfixExpression(t, cond.Condition, "x", "<", "y")

		if len(cond.Consequence.Statements) != 1 {
			t.Errorf("if consequence branch should have 1 statement; got %d", len(cond.Consequence.Statements))
		}
		cons, ok := cond.Consequence.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("if consequence's branch statement should be *ast.ExpressionStatement; got %T", cond.Consequence.Statements[0])
		}
		testIdLiteral(t, cons.Expression, "x")

		if len(cond.Alternative.Statements) != 1 {
			t.Errorf("if alternative branch should have 1 statement; got %d", len(cond.Alternative.Statements))
		}
		alt, ok := cond.Alternative.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("if alternative's branch statement should be *ast.ExpressionStatement; got %T", cond.Alternative.Statements[0])
		}
		testIdLiteral(t, alt.Expression, "y")
	})
}

func TestOperatorPrecedence(t *testing.T) {
	t.Run("operator precedence", func(t *testing.T) {
		tt := []struct {
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

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				m := assertEval(t, tc.input, tc.length)

				actual := m.String()
				if actual != tc.output {
					t.Errorf("precedence notation should be %q; got %q", tc.output, actual)
				}
			})
		}
	})
}

func testLetStatement(t *testing.T, s ast.Statement, name string, val interface{}) {
	if s.TokenLiteral() != "let" {
		t.Errorf(`first token for let statement should be "let"; got %q`, s.TokenLiteral())
	}
	let, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("statement should be *ast.LetStatement; got %T", s)
	}
	if let.Name.TokenLiteral() != name {
		t.Errorf("name of the variable should be %q; got %q", name, let.Name.TokenLiteral())
	}
	testLiteral(t, let.Value, val)
}

func testLiteral(t *testing.T, exp ast.Expression, val interface{}) {
	switch val := val.(type) {
	case string:
		testIdLiteral(t, exp, val)
	case int:
		testNumberLiteral(t, exp, float64(val))
	case float64:
		testNumberLiteral(t, exp, val)
	case bool:
		testBoolLiteral(t, exp, val)
	default:
		t.Errorf("type of exp not handled: %T", val)
	}
}

func testIdLiteral(t *testing.T, exp ast.Expression, val string) {
	id, ok := exp.(*ast.Id)
	if !ok {
		t.Errorf("expression should be *ast.Id; got %T", exp)
	}
	if id.Value != val {
		t.Errorf("identifier value (it's name) should be %q; got %q", val, id.Value)
	}
	if id.TokenLiteral() != val {
		t.Errorf("identifier literal (it's name) should be %q; got %q", val, id.TokenLiteral())
	}
}

func testNumberLiteral(t *testing.T, exp ast.Expression, val float64) {
	num, ok := exp.(*ast.Number)
	if !ok {
		t.Errorf("expression should be *ast.Number; got %T", exp)
	}
	if num.Value != val {
		t.Errorf("number value should be %v; got %v", val, num.Value)
	}
	vals := fmt.Sprintf("%v", val)
	if num.TokenLiteral() != vals {
		t.Errorf("number literal should be %v; got %v", vals, num.TokenLiteral())
	}
}

func testStringLiteral(t *testing.T, exp ast.Expression, val string) {
	str, ok := exp.(*ast.String)
	if !ok {
		t.Errorf("expression should be *ast.String; got %T", exp)
	}
	if str.Value != val {
		t.Errorf("string value should be %q; got %q", val, str.Value)
	}
	if str.TokenLiteral() != val {
		t.Errorf("string literal should be %q; got %q", val, str.TokenLiteral())
	}
}

func testBoolLiteral(t *testing.T, exp ast.Expression, val bool) {
	b, ok := exp.(*ast.Bool)
	if !ok {
		t.Errorf("expression should be *ast.Bool; got %T", exp)
	}
	if b.Value != val {
		t.Errorf("boolean value should be %v; got %v", val, b.Value)
	}
	bs := fmt.Sprintf("%t", val)
	if b.TokenLiteral() != bs {
		t.Errorf("boolean literal should be %v; got %v", bs, b.TokenLiteral())
	}
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, op string, right interface{}) {
	infix, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("statement should be *ast.InfixExpression; got %T", exp)
	}
	testLiteral(t, infix.Left, left)
	if infix.Op != op {
		t.Errorf("infix operator should be %q; got %q", op, infix.Op)
	}
	testLiteral(t, infix.Right, right)
}

func assertEval(t *testing.T, input string, expectedStatementsLen int) *ast.Module {
	l := lexer.New(strings.NewReader(input))
	p := New(l)
	m, errors := p.Parse()
	if len(errors) > 0 {
		t.Errorf("parse of the code should not produce errors; got %v", errors)
	}

	if expectedStatementsLen > 0 && len(m.Statements) != expectedStatementsLen {
		t.Errorf("statements produced should match %d; got %d", expectedStatementsLen, len(m.Statements))
	}

	return m
}
