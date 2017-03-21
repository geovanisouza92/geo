package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"geo/ast"
)

type Object interface {
	Type() ObjectType
	String() string
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hashable interface {
	HashKey() HashKey
}

type Number struct {
	Value   float64
	hashKey HashKey
}

func NewNumber(value float64) *Number {
	n := &Number{Value: value}
	n.hashKey = HashKey{n.Type(), uint64(value)}
	return n
}

func (n *Number) Type() ObjectType {
	return TypeNumber
}

func (n *Number) String() string {
	return fmt.Sprintf("%v", n.Value)
}

func (n *Number) HashKey() HashKey {
	return n.hashKey
}

type Bool struct {
	Value   bool
	hashKey HashKey
}

func NewBool(value bool) *Bool {
	b := &Bool{Value: value}
	if value {
		b.hashKey = HashKey{b.Type(), 1}
	} else {
		b.hashKey = HashKey{b.Type(), 0}
	}
	return b
}

func (b *Bool) Type() ObjectType {
	return TypeBool
}

func (b *Bool) String() string {
	return fmt.Sprintf("%t", b.Value)
}

func (b *Bool) HashKey() HashKey {
	return b.hashKey
}

type String struct {
	Value   string
	hashKey HashKey
}

func NewString(value string) *String {
	s := &String{Value: value}
	h := fnv.New64a()
	h.Write([]byte(value))
	s.hashKey = HashKey{s.Type(), h.Sum64()}
	return s
}

func (s *String) Type() ObjectType {
	return TypeString
}

func (s *String) String() string {
	return s.Value
}

func (s *String) HashKey() HashKey {
	return s.hashKey
}

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType {
	return TypeArray
}

func (a *Array) String() string {
	var b bytes.Buffer

	elms := []string{}
	for _, elm := range a.Elements {
		elms = append(elms, elm.String())
	}

	b.WriteString("[")
	b.WriteString(strings.Join(elms, ", "))
	b.WriteString("]")

	return b.String()
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType {
	return TypeHash
}

func (h *Hash) String() string {
	var b bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, pair.Key.String()+": "+pair.Value.String())
	}

	b.WriteString("{")
	b.WriteString(strings.Join(pairs, ", "))
	b.WriteString("}")

	return b.String()
}

type Null struct {
}

func (n *Null) Type() ObjectType {
	return TypeNull
}

func (n *Null) String() string {
	return "null"
}

type Return struct {
	Value Object
}

func (r *Return) Type() ObjectType {
	return TypeReturn
}

func (r *Return) String() string {
	return r.Value.String()
}

type Error struct {
	Message error
}

func (e *Error) Type() ObjectType {
	return TypeError
}

func (e *Error) String() string {
	return e.Message.Error()
}

type Scope struct {
	internal map[string]Object
	parent   *Scope
}

func NewRootScope() *Scope {
	in := make(map[string]Object)
	return &Scope{internal: in}
}

func NewScope(parent *Scope) *Scope {
	s := NewRootScope()
	s.parent = parent
	return s
}

func (e *Scope) Get(name string) (Object, bool) {
	val, ok := e.internal[name]
	if !ok && e.parent != nil {
		val, ok = e.parent.Get(name)
	}
	return val, ok
}

func (e *Scope) Set(name string, val Object) Object {
	// TODO: First, check if value was already defined
	// NOTE: Should return defined value or simply emits an error?
	// NOTE: Should emit an error when name is conflicting with builtin?
	e.internal[name] = val
	return val
}

type Fn struct {
	Params []*ast.Id
	Body   *ast.BlockStatement
	Scope  *Scope // TODO: Should this be moved for each block?
}

func (f *Fn) Type() ObjectType {
	return TypeFn
}

func (f *Fn) String() string {
	var b bytes.Buffer

	params := []string{}
	for _, p := range f.Params {
		params = append(params, p.String())
	}

	b.WriteString("fn(")
	b.WriteString(strings.Join(params, ", "))
	b.WriteString(") {\n")
	b.WriteString(f.Body.String())
	b.WriteString("\n}")

	return b.String()
}

type BuiltinFunction func(...Object) Object

type Builtin struct {
	// NOTE: Maybe declare other fields, like Doc
	Name   string
	Params []ObjectType
	Fn     BuiltinFunction
}

func (r *Builtin) Type() ObjectType {
	return TypeBuiltin
}

func (r *Builtin) String() string {
	return "builtin function"
}
