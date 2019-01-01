package eval

import (
	"fmt"

	"github.com/geovanisouza92/geo/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Name:   "len",
		Params: []object.ObjectType{object.TypeArray | object.TypeString},
		Fn: func(args ...object.Object) object.Object {
			switch arg := args[0].(type) {
			case *object.Array:
				return object.NewNumber(float64(len(arg.Elements)))
			case *object.String:
				return object.NewNumber(float64(len(arg.Value)))
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"head": &object.Builtin{
		Name:   "head",
		Params: []object.ObjectType{object.TypeArray},
		Fn: func(args ...object.Object) object.Object {
			ary := args[0].(*object.Array)
			if len(ary.Elements) > 0 {
				return ary.Elements[0]
			}

			return Null
		},
	},
	"last": &object.Builtin{
		Name:   "last",
		Params: []object.ObjectType{object.TypeArray},
		Fn: func(args ...object.Object) object.Object {
			ary := args[0].(*object.Array)
			length := len(ary.Elements)
			if length > 0 {
				return ary.Elements[length-1]
			}

			return Null
		},
	},
	"tail": &object.Builtin{
		Name:   "tail",
		Params: []object.ObjectType{object.TypeArray},
		Fn: func(args ...object.Object) object.Object {
			ary := args[0].(*object.Array)
			length := len(ary.Elements)
			if length > 0 {
				elms := make([]object.Object, length-1, length-1)
				copy(elms, ary.Elements[1:length])
				return &object.Array{elms}
			}

			return Null
		},
	},
	"push": &object.Builtin{
		Name:   "push",
		Params: []object.ObjectType{object.TypeArray, object.TypeAny},
		Fn: func(args ...object.Object) object.Object {
			ary := args[0].(*object.Array)
			length := len(ary.Elements)

			elms := make([]object.Object, length+1, length+1)
			copy(elms, ary.Elements)
			elms[length] = args[1]

			return &object.Array{elms}
		},
	},
	"puts!": &object.Builtin{
		Name: "puts!",
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.String())
			}

			return Null
		},
	},
}
