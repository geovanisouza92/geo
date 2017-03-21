package eval

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"geo/object"
)

func TestEval(t *testing.T) {
	Convey("numbers", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			actual := testEval(expected.input)
			testNumber(actual, expected.val)
		}
	})
	Convey("bool", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			actual := testEval(expected.input)
			testBool(actual, expected.val)
		}
	})
	Convey("string", t, func() {
		tests := []struct {
			input string
			val   string
		}{
			{`"foobar"`, "foobar"},
			{`"foo bar"`, "foo bar"},
			{`"foo \"bar"`, `foo \"bar`},
			{`"foo" + "bar"`, "foobar"},
		}

		for _, expected := range tests {
			actual := testEval(expected.input)
			testString(actual, expected.val)
		}
	})
	Convey("array", t, func() {
		input := "[1, 2 * 2, 3 + 3]"
		actual := testEval(input)
		ary, ok := actual.(*object.Array)
		So(ok, ShouldBeTrue)
		So(len(ary.Elements), ShouldEqual, 3)
		testNumber(ary.Elements[0], 1)
		testNumber(ary.Elements[1], 4)
		testNumber(ary.Elements[2], 6)
	})
	Convey("hash", t, func() {
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

		actual := testEval(input)
		res, ok := actual.(*object.Hash)
		So(ok, ShouldBeTrue)
		So(len(res.Pairs), ShouldEqual, len(expected))

		for k, v := range expected {
			Convey(fmt.Sprintf("testing %v", v), func() {
				pair, ok := res.Pairs[k]
				So(ok, ShouldBeTrue)
				testNumber(pair.Value, v)
			})
		}
	})
	Convey("index expressions", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			actual := testEval(expected.input)
			switch val := expected.val.(type) {
			case int:
				testNumber(actual, float64(val))
			case float64:
				testNumber(actual, val)
			case string:
				testString(actual, val)
			default:
				testNull(actual)
			}
		}
	})
	Convey("prefix expressions", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			Convey(expected.input, func() {
				actual := testEval(expected.input)
				testBool(actual, expected.val)
			})
		}
	})
	Convey("infix expressions", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			actual := testEval(expected.input)
			testNumber(actual, expected.val)
		}
	})
	Convey("if expressions", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			actual := testEval(expected.input)
			switch val := expected.output.(type) {
			case int:
				testNumber(actual, float64(val))
			case float64:
				testNumber(actual, val)
			default:
				testNull(actual)
			}
		}
	})
	Convey("return expressions", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			actual := testEval(expected.input)
			testNumber(actual, expected.output)
		}
	})
	Convey("error handling", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			Convey(expected.input, func() {
				actual := testEval(expected.input)
				err, ok := actual.(*object.Error)
				So(ok, ShouldBeTrue)
				So(err.Message.Error(), ShouldEqual, expected.message)
			})
		}
	})
	Convey("let statements", t, func() {
		tests := []struct {
			input string
			val   float64
		}{
			{"let a = 5; a;", 5},
			{"let a = 5 * 5; a;", 25},
			{"let a = 5; let b = a; b;", 5},
			{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
		}

		for _, expected := range tests {
			actual := testEval(expected.input)
			testNumber(actual, expected.val)
		}
	})
	Convey("functions", t, func() {
		input := "fn(x) { x + 2; };"
		expectedBody := "(x + 2)"

		actual := testEval(input)
		fn, ok := actual.(*object.Fn)
		So(ok, ShouldBeTrue)
		So(len(fn.Params), ShouldEqual, 1)
		So(fn.Params[0].String(), ShouldEqual, "x")
		So(fn.Body.String(), ShouldEqual, expectedBody)
	})
	Convey("function applications", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			Convey(expected.input, func() {
				actual := testEval(expected.input)
				testNumber(actual, expected.value)
			})
		}
	})
	Convey("closures", t, func() {
		input := `
		let newAdder = fn(x) {
			fn(y) { x + y };
		};

		let addTwo = newAdder(2);
		addTwo(2);`

		actual := testEval(input)
		testNumber(actual, 4)
	})
	Convey("builtin functions", t, func() {
		tests := []struct {
			input string
			value interface{}
		}{
			{`len("")`, 0},
			{`len("four")`, 4},
			{`len("hello world")`, 11},
			{`len(1)`, "argument to `len` must be (string, array), got number"},
			// {`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
			{`len([1, 2, 3])`, 3},
			{`len([])`, 0},
			{`puts!("hello", "world!")`, nil},
			{`head([1, 2, 3])`, 1},
			{`head([])`, nil},
			{`head(1)`, "argument to `head` must be (array), got number"},
			{`last([1, 2, 3])`, 3},
			{`last([])`, nil},
			{`last(1)`, "argument to `last` must be (array), got number"},
			{`tail([1, 2, 3])`, []int{2, 3}},
			{`tail([])`, nil},
			{`push([], 1)`, []int{1}},
			{`push(1, 1)`, "argument to `push` must be (array), got number"},
		}

		for _, expected := range tests {
			Convey(expected.input, func() {

				actual := testEval(expected.input)

				switch val := expected.value.(type) {
				case int:
					testNumber(actual, float64(val))
				case []int:
					ary, ok := actual.(*object.Array)
					So(ok, ShouldBeTrue)
					So(len(ary.Elements), ShouldEqual, len(val))
					for i, it := range val {
						Convey(fmt.Sprintf("checking for value: %v", it), func() {
							testNumber(ary.Elements[i], float64(it))
						})
					}
				case float64:
					testNumber(actual, val)
				case string:
					err, ok := actual.(*object.Error)
					So(ok, ShouldBeTrue)
					So(err.Message.Error(), ShouldEqual, val)
				}
			})
		}
	})
	Convey("map+reduce", t, func() {
		tests := []struct {
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

		for _, expected := range tests {
			actual := testEval(expected.input)

			switch val := expected.val.(type) {
			case []float64:
				ary, ok := actual.(*object.Array)
				So(ok, ShouldBeTrue)
				So(len(ary.Elements), ShouldEqual, len(val))
				for i, it := range val {
					Convey(fmt.Sprintf("checking for value: %v", it), func() {
						testNumber(ary.Elements[i], float64(it))
					})
				}
			case int:
				testNumber(actual, float64(val))
			}
		}
	})
	Convey("block scopes", t, func() {
		input := `
let x = 5;
if (true) {
	let x = 6;
};
return x;
`
		actual := testEval(input)
		testNumber(actual, 5)
	})
	Convey("pipes", t, func() {
		tests := []struct {
			input string
			val   float64
		}{
			{"fn(x, y){ x * y }(2) | fn(y, f) { f(y) }(5)", 10},
			{"[2, 3] | head | fn(x, y){ x * y } | fn(y, f) { f(y) }(5)", 10},
		}

		for _, expected := range tests {
			Convey(expected.input, func() {
				actual := testEval(expected.input)
				testNumber(actual, expected.val)
			})
		}
	})
}

func testEval(input string) object.Object {
	s, err := Compile(input)
	So(err, ShouldBeNil)
	c := NewContext(object.NewRootScope())
	return c.Eval(s)
}

func testNumber(obj object.Object, val float64) {
	num, ok := obj.(*object.Number)
	if !ok {
		err := obj.(*object.Error)
		So(err.Message, ShouldEqual, "")
	}
	So(ok, ShouldBeTrue)
	So(num.Value, ShouldEqual, val)
}

func testBool(obj object.Object, val bool) {
	b, ok := obj.(*object.Bool)
	So(ok, ShouldBeTrue)
	So(b.Value, ShouldEqual, val)
}

func testString(obj object.Object, val string) {
	str, ok := obj.(*object.String)
	So(ok, ShouldBeTrue)
	So(str.Value, ShouldEqual, val)
}

func testNull(obj object.Object) {
	So(obj, ShouldEqual, Null)
}
