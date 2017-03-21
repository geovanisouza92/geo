package object

//go:generate stringer -type=ObjectType

import (
	"strings"
)

type ObjectType int

const (
	TypeError ObjectType = 1 << iota
	TypeNumber
	TypeBool
	TypeString
	TypeArray
	TypeHash
	TypeNull
	TypeReturn
	TypeFn
	TypeBuiltin
)

const TypeAny = TypeError | TypeNumber | TypeBool | TypeString | TypeArray | TypeHash | TypeNull | TypeReturn | TypeFn | TypeBuiltin

var types = map[ObjectType]string{
	TypeError:   "error",
	TypeNumber:  "number",
	TypeBool:    "bool",
	TypeString:  "string",
	TypeArray:   "array",
	TypeHash:    "hash",
	TypeNull:    "null",
	TypeReturn:  "return",
	TypeFn:      "function",
	TypeBuiltin: "builtin function",
	// TypeAny:     "any",
}

func ObjectTypeToString(ty ObjectType) string {
	if s, ok := types[ty]; ok {
		return s
	}
	return ""
}

func ObjectTypesToString(ty ObjectType) string {
	s := []string{}

	for k, v := range types {
		if ty&k != 0 {
			s = append(s, v)
		}
	}

	return strings.Join(s, ", ")
}
