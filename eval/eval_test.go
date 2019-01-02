package eval

import (
	"fmt"
	"testing"

	"github.com/geovanisouza92/geo/object"
)

func TestEval(t *testing.T) {
	t.Run("numbers", func(t *testing.T) {
		tt := []struct {
			input string
			val   float64
		}{
			{"5", 5},
			{"10", 10},
			{"5.3", 5.3},
			{"12.1", 12.1},
			{"-5", -5},
			{"-10", -10},
			{"-5.3", -5.3},
			{"-12.1", -12.1},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)
			testNumber(t, actual, tc.val)
		}
	})

	t.Run("bool", func(t *testing.T) {
		tt := []struct {
			input string
			val   bool
		}{
			{"true", true},
			{"false", false},
			{"1 < 2", true},
			{"1 > 2", false},
			{"1 <= 1", true},
			{"1 >= 1", true},
			{"1 == 1", true},
			{"1 != 1", false},
			{"1 == 2", false},
			{"1 != 2", true},
			{"true == true", true},
			{"false == false", true},
			{"true == false", false},
			{"true != false", true},
			{"false != true", true},
			{"(1 < 2) == true", true},
			{"(1 < 2) == false", false},
			{"(1 > 2) == true", false},
			{"(1 > 2) == false", true},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)
			testBool(t, actual, tc.val)
		}
	})

	t.Run("string", func(t *testing.T) {
		tt := []struct {
			input string
			val   string
		}{
			{`"foobar"`, "foobar"},
			{`"foo bar"`, "foo bar"},
			{`"foo \"bar"`, `foo \"bar`},
			{`"foo" + "bar"`, "foobar"},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)
			testString(t, actual, tc.val)
		}
	})

	t.Run("array", func(t *testing.T) {
		input := "[1, 2 * 2, 3 + 3]"
		actual := testEval(t, input)
		ary, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("value should be *object.Array; got %T", actual)
		}
		if len(ary.Elements) != 3 {
			t.Errorf("array should have 3 elements; got %d", len(ary.Elements))
		}
		testNumber(t, ary.Elements[0], 1)
		testNumber(t, ary.Elements[1], 4)
		testNumber(t, ary.Elements[2], 6)
	})

	t.Run("hash", func(t *testing.T) {
		input := `let two = "two";
				{
					"one": 10 - 9,
					two: 1 + 1,
					"thr" + "ee": 6 / 2,
					4: 4,
					true: 5,
					false: 6,
				}`

		expected := map[object.HashKey]float64{
			(object.NewString("one").HashKey()):   1,
			(object.NewString("two").HashKey()):   2,
			(object.NewString("three").HashKey()): 3,
			(object.NewNumber(4).HashKey()):       4,
			(True.HashKey()):                      5,
			(False.HashKey()):                     6,
		}

		actual := testEval(t, input)
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("value should be *object.Hash; got %T", actual)
		}
		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash should have %d pairs; got %d", len(expected), len(hash.Pairs))
		}

		for k, v := range expected {
			t.Run(fmt.Sprintf("testing %v", v), func(t *testing.T) {
				pair, ok := hash.Pairs[k]
				if !ok {
					t.Errorf("hash value for key %q should exist", k)
				}
				testNumber(t, pair.Value, v)
			})
		}
	})

	t.Run("index expressions", func(t *testing.T) {
		tt := []struct {
			input string
			val   interface{}
		}{
			{"[1, 2, 3][0]", 1},
			{"[1, 2, 3][1]", 2},
			{"[1, 2, 3][2]", 3},
			{"let i = 0; [1][i];", 1},
			{"[1, 2, 3][1 + 1];", 3},
			{"let myArray = [1, 2, 3]; myArray[2];", 3},
			{"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];", 6},
			{"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]", 2},
			{"[1, 2, 3][3]", nil},
			{"[1, 2, 3][-1]", nil}, // NOTE: len + index
			{`{"foo": 5}["foo"]`, 5},
			{`{"foo": 5}["bar"]`, nil},
			{`let key = "foo"; {"foo": 5}[key]`, 5},
			{`{}["foo"]`, nil},
			{`{5: 5}[5]`, 5},
			{`{true: 5}[true]`, 5},
			{`{false: 5}[false]`, 5},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)
			switch val := tc.val.(type) {
			case int:
				testNumber(t, actual, float64(val))
			case float64:
				testNumber(t, actual, val)
			case string:
				testString(t, actual, val)
			default:
				testNull(t, actual)
			}
		}
	})

	t.Run("prefix expressions", func(t *testing.T) {
		tt := []struct {
			input string
			val   bool
		}{
			{"!true", false},
			{"!false", true},
			{"!5", false},
			{"!!5", true},
			{"!!true", true},
			{"!!false", false},
			{`!""`, true},
			{`!"Hello"`, false},
			{"![]", true},
			{"![1]", false},
			{"!{}", true},
			{`!{"foo": "bar"}`, false},
		}

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				actual := testEval(t, tc.input)
				testBool(t, actual, tc.val)
			})
		}
	})

	t.Run("infix expressions", func(t *testing.T) {
		tt := []struct {
			input string
			val   float64
		}{
			{"5", 5},
			{"10", 10},
			{"-5", -5},
			{"-10", -10},
			{"5 + 5 + 5 + 5 - 10", 10},
			{"2 * 2 * 2 * 2 * 2", 32},
			{"-50 + 100 + -50", 0},
			{"5 * 2 + 10", 20},
			{"5 + 2 * 10", 25},
			{"20 + 2 * -10", 0},
			{"50 / 2 * 2 + 10", 60},
			{"2 * (5 + 10)", 30},
			{"3 * 3 * 3 + 10", 37},
			{"3 * (3 * 3) + 10", 37},
			{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)
			testNumber(t, actual, tc.val)
		}
	})

	t.Run("if expressions", func(t *testing.T) {
		tt := []struct {
			input  string
			output interface{}
		}{
			{"if (true) { 10 }", 10},
			{"if (false) { 10 }", nil},
			{"if (1) { 10 }", 10},
			{"if (1 < 2) { 10 }", 10},
			{"if (1 > 2) { 10 }", nil},
			{"if (1 > 2) { 10 } else { 20 }", 20},
			{"if (1 < 2) { 10 } else { 20 }", 10},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)
			switch val := tc.output.(type) {
			case int:
				testNumber(t, actual, float64(val))
			case float64:
				testNumber(t, actual, val)
			default:
				testNull(t, actual)
			}
		}
	})

	t.Run("return expressions", func(t *testing.T) {
		tt := []struct {
			input  string
			output float64
		}{
			{"return 10;", 10},
			{"return 10; 9;", 10},
			{"return 2 * 5; 9;", 10},
			{"9; return 2 * 5; 9;", 10},
			{
				`
		if (10 > 1) {
			if (10 > 1) {
				return 10;
			}
			return 1;
		}
		`,
				10,
			},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)
			testNumber(t, actual, tc.output)
		}
	})

	t.Run("error handling", func(t *testing.T) {
		tt := []struct {
			input   string
			message string
		}{
			{
				"5 + true;",
				"type mismatch: TypeNumber + TypeBool",
			},
			{
				"5 + true; 5;",
				"type mismatch: TypeNumber + TypeBool",
			},
			{
				"-true",
				"unknown operator: -TypeBool",
			},
			{
				"true + false;",
				"unknown operator: TypeBool + TypeBool",
			},
			{
				"5; true + false; 5",
				"unknown operator: TypeBool + TypeBool",
			},
			{
				"if (10 > 1) { true + false; }",
				"unknown operator: TypeBool + TypeBool",
			},
			{
				`
		if (10 > 1) {
			if (10 > 1) {
				return true + false;
			}
			return 1;
		}
		`,
				"unknown operator: TypeBool + TypeBool",
			},
			{
				"foobar",
				"identifier not found: foobar",
			},
			{
				`"Hello" - "World"`,
				"unknown operator: TypeString - TypeString",
			},
			{
				`{"name": "Monkey"}[fn(x) { x }]`,
				"unusable as hash key: TypeFn",
			},
		}

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				actual := testEval(t, tc.input)
				err, ok := actual.(*object.Error)
				if !ok {
					t.Errorf("value should be *object.Error; got %T", actual)
				}
				if err.Message.Error() != tc.message {
					t.Errorf("error message should be %q; got %q", tc.message, err.Message.Error())
				}
			})
		}
	})

	t.Run("let statements", func(t *testing.T) {
		tt := []struct {
			input string
			val   float64
		}{
			{"let a = 5; a;", 5},
			{"let a = 5 * 5; a;", 25},
			{"let a = 5; let b = a; b;", 5},
			{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)
			testNumber(t, actual, tc.val)
		}
	})

	t.Run("functions", func(t *testing.T) {
		input := "fn(x) { x + 2; };"
		expectedBody := "(x + 2)"

		actual := testEval(t, input)
		fn, ok := actual.(*object.Fn)
		if !ok {
			t.Errorf("value should be *object.Fn; got %T", actual)
		}
		if len(fn.Params) != 1 {
			t.Errorf("function should have one parameter; got %d", len(fn.Params))
		}
		if fn.Params[0].String() != "x" {
			t.Errorf(`function's first parameter name should be "x"; got %q`, fn.Params[0].String())
		}
		if fn.Body.String() != expectedBody {
			t.Errorf("function body should be %q; got %q", expectedBody, fn.Body.String())
		}
	})

	t.Run("function applications", func(t *testing.T) {
		tt := []struct {
			input string
			value float64
		}{
			{"let identity = fn(x) { x; }; identity(5);", 5},
			{"let identity = fn(x) { return x; }; identity(5);", 5},
			{"let double = fn(x) { x * 2; }; double(5);", 10},
			{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
			{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
			{"fn(x) { x; }(5)", 5},
			{"fn(x, y) { x + y; }(2)(3)", 5},
			{"fn(x, y) { x + y; }(2)(3, 4)", 5},
			{"fn(x, y) { x + y; }(2, 3, 5)", 5},
		}

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				actual := testEval(t, tc.input)
				testNumber(t, actual, tc.value)
			})
		}
	})

	t.Run("closures", func(t *testing.T) {
		input := `
		let newAdder = fn(x) {
			fn(y) { x + y };
		};

		let addTwo = newAdder(2);
		addTwo(2);`

		actual := testEval(t, input)
		testNumber(t, actual, 4)
	})

	t.Run("builtin functions", func(t *testing.T) {
		tt := []struct {
			input string
			value interface{}
		}{
			{`len("")`, 0},
			{`len("four")`, 4},
			{`len("hello world")`, 11},
			{`len(1)`, "argument to `len` must be (TypeString, TypeArray), got TypeNumber"},
			// {`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
			{`len([1, 2, 3])`, 3},
			{`len([])`, 0},
			{`puts!("hello", "world!")`, nil},
			{`head([1, 2, 3])`, 1},
			{`head([])`, nil},
			{`head(1)`, "argument to `head` must be (TypeArray), got TypeNumber"},
			{`last([1, 2, 3])`, 3},
			{`last([])`, nil},
			{`last(1)`, "argument to `last` must be (TypeArray), got TypeNumber"},
			{`tail([1, 2, 3])`, []int{2, 3}},
			{`tail([])`, nil},
			{`push([], 1)`, []int{1}},
			{`push(1, 1)`, "argument to `push` must be (TypeArray), got TypeNumber"},
		}

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {

				actual := testEval(t, tc.input)

				switch val := tc.value.(type) {
				case int:
					testNumber(t, actual, float64(val))
				case []int:
					ary, ok := actual.(*object.Array)
					if !ok {
						t.Errorf("value should be *object.Array; got %T", actual)
					}
					if len(ary.Elements) != len(val) {
						t.Errorf("array should have %d elements; got %d", len(val), len(ary.Elements))
					}
					for i, it := range val {
						t.Run(fmt.Sprintf("checking for value: %v", it), func(t *testing.T) {
							testNumber(t, ary.Elements[i], float64(it))
						})
					}
				case float64:
					testNumber(t, actual, val)
				case string:
					err, ok := actual.(*object.Error)
					if !ok {
						t.Errorf("value should be *object.Error; got %T", actual)
					}
					if err.Message.Error() != val {
						t.Errorf("error message should be %q; got %q", val, err.Message.Error())
					}
				}
			})
		}
	})

	t.Run("map+reduce", func(t *testing.T) {
		tt := []struct {
			input string
			val   interface{}
		}{
			{`
		let map = fn(f, arr) {
			let iter = fn(acc, arr) {
				if (len(arr) == 0) {
					return acc
				}
				iter(push(acc, f(head(arr))), tail(arr))
			};

			iter([], arr);
		};

		let a = [1, 2, 3, 4];
		let double = fn(x) { x * 2 };
		a | map(double)
		`, []float64{2, 4, 6, 8}},
			{`
		let reduce = fn(f, seed, arr) {
			let iter = fn(acc, arr) {
				if (len(arr) == 0) {
					return acc
				}
				iter(f(acc, head(arr)), tail(arr))
			};

			iter(seed, arr);
		};

		let sum = fn(arr) {
			reduce(fn(acc, it) { acc + it }, 0, arr);
		};

		[1, 2, 3, 4, 5] | sum;
		`, 15},
		}

		for _, tc := range tt {
			actual := testEval(t, tc.input)

			switch val := tc.val.(type) {
			case []float64:
				ary, ok := actual.(*object.Array)
				if !ok {
					t.Errorf("value should be *object.Array; got %T", actual)
				}
				if len(ary.Elements) != len(val) {
					t.Errorf("array should have %d elements; got %d", len(val), len(ary.Elements))
				}
				for i, it := range val {
					t.Run(fmt.Sprintf("checking for value: %v", it), func(t *testing.T) {
						testNumber(t, ary.Elements[i], float64(it))
					})
				}
			case int:
				testNumber(t, actual, float64(val))
			}
		}
	})

	t.Run("block scopes", func(t *testing.T) {
		input := `
			let x = 5;
			if (true) {
				let x = 6;
			};
			return x;
		`
		actual := testEval(t, input)
		testNumber(t, actual, 5)
	})

	t.Run("pipes", func(t *testing.T) {
		tt := []struct {
			input string
			val   float64
		}{
			{"fn(x, y){ x * y }(2) | fn(y, f) { f(y) }(5)", 10},
			{"[2, 3] | head | fn(x, y){ x * y } | fn(y, f) { f(y) }(5)", 10},
		}

		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				actual := testEval(t, tc.input)
				testNumber(t, actual, tc.val)
			})
		}
	})
}

func testEval(t *testing.T, input string) object.Object {
	s, err := Compile(input)
	if err != nil {
		t.Errorf("compilation should succeed; got err %v", err)
	}
	c := NewContext(object.NewRootScope())
	return c.Eval(s)
}

func testNumber(t *testing.T, obj object.Object, val float64) {
	num, ok := obj.(*object.Number)
	if !ok {
		err := obj.(*object.Error)
		if err.Message.Error() != "" {
			t.Errorf("number should be valid; got err %v", err.Message)
		}
	}
	if num.Value != val {
		t.Errorf("number value should be %v; got %v", val, num.Value)
	}
}

func testBool(t *testing.T, obj object.Object, val bool) {
	b, ok := obj.(*object.Bool)
	if !ok {
		t.Errorf("object should be *object.Bool; got %T", obj)
	}
	if b.Value != val {
		t.Errorf("boolean value should be %v; got %v", val, b.Value)
	}
}

func testString(t *testing.T, obj object.Object, val string) {
	str, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object should be *object.String; got %T", obj)
	}
	if str.Value != val {
		t.Errorf("string value should be %q; got %q", val, str.Value)
	}
}

func testNull(t *testing.T, obj object.Object) {
	if obj != Null {
		t.Errorf("object should be 'null'; got %v", obj)
	}
}
