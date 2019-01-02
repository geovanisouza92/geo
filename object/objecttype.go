package object

//go:generate stringer -type=ObjectType

import (
	"sort"
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

type ByObjectType []ObjectType

func (a ByObjectType) Len() int           { return len(a) }
func (a ByObjectType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByObjectType) Less(i, j int) bool { return a[i] < a[j] }

var objectTypes = []ObjectType{
	TypeError,
	TypeNumber,
	TypeBool,
	TypeString,
	TypeArray,
	TypeHash,
	TypeNull,
	TypeReturn,
	TypeFn,
	TypeBuiltin,
	// TypeAny,
}

func ObjectTypesToString(t ObjectType) string {
	s := []string{}

	sort.Stable(ByObjectType(objectTypes))

	for _, it := range objectTypes {
		if t&it != 0 {
			s = append(s, it.String())
		}
	}

	return strings.Join(s, ", ")
}
