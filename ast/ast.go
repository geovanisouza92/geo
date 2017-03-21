package ast

import (
	"bytes"
	"fmt"
	"strings"

	"geo/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	s()
}

type Expression interface {
	Node
	e()
}

type Module struct {
	Statements []Statement
}

func (m *Module) TokenLiteral() string {
	return ""
}

func (m *Module) String() string {
	var b bytes.Buffer

	for _, s := range m.Statements {
		b.WriteString(s.String())
	}

	return b.String()
}

type LetStatement struct {
	Token token.Token
	Name  *Id
	Value Expression
}

func (l *LetStatement) s() {}

func (l *LetStatement) TokenLiteral() string {
	return l.Token.Literal
}

func (l *LetStatement) String() string {
	var b bytes.Buffer

	b.WriteString(l.TokenLiteral() + " ")
	b.WriteString(l.Name.String())
	b.WriteString(" = ")
	if l.Value != nil {
		b.WriteString(l.Value.String())
	}
	b.WriteString(";")

	return b.String()
}

type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (r *ReturnStatement) s() {}

func (r *ReturnStatement) TokenLiteral() string {
	return r.Token.Literal
}

func (r *ReturnStatement) String() string {
	var b bytes.Buffer

	b.WriteString(r.TokenLiteral() + " ")
	if r.Value != nil {
		b.WriteString(r.Value.String())
	}
	b.WriteString(";")

	return b.String()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (e *ExpressionStatement) s() {}

func (e *ExpressionStatement) TokenLiteral() string {
	return e.Token.Literal
}

func (e *ExpressionStatement) String() string {
	if e.Expression != nil {
		return e.Expression.String()
	}
	return ""
}

type Id struct {
	Token token.Token
	Value string
}

func (i *Id) e() {}

func (i *Id) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Id) String() string {
	return i.Value
}

type Number struct {
	Token token.Token
	Value float64
}

func (n *Number) e() {}

func (n *Number) TokenLiteral() string {
	return n.Token.Literal
}

func (n *Number) String() string {
	return fmt.Sprintf("%v", n.Value)
}

type Bool struct {
	Token token.Token
	Value bool
}

func (b *Bool) e() {}

func (b *Bool) TokenLiteral() string {
	return b.Token.Literal
}

func (b *Bool) String() string {
	return fmt.Sprintf("%t", b.Value)
}

type String struct {
	Token token.Token
	Value string
}

func (s *String) e() {}

func (s *String) TokenLiteral() string {
	return s.Token.Literal
}

func (s *String) String() string {
	return s.Value
}

type Array struct {
	Token    token.Token
	Elements []Expression
}

func (a *Array) e() {}

func (a *Array) TokenLiteral() string {
	return a.Token.Literal
}

func (a *Array) String() string {
	var b bytes.Buffer

	elms := []string{}
	for _, e := range a.Elements {
		elms = append(elms, e.String())
	}

	b.WriteString("[")
	b.WriteString(strings.Join(elms, ", "))
	b.WriteString("]")

	return b.String()
}

type PrefixExpression struct {
	Token token.Token
	Op    string
	Right Expression
}

func (p *PrefixExpression) e() {}

func (p *PrefixExpression) TokenLiteral() string {
	return p.Token.Literal
}

func (p *PrefixExpression) String() string {
	var b bytes.Buffer

	b.WriteString("(")
	b.WriteString(p.Op)
	b.WriteString(p.Right.String())
	b.WriteString(")")

	return b.String()
}

type InfixExpression struct {
	Token token.Token
	Left  Expression
	Op    string
	Right Expression
}

func (i *InfixExpression) e() {}

func (i *InfixExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *InfixExpression) String() string {
	var b bytes.Buffer

	b.WriteString("(")
	b.WriteString(i.Left.String())
	b.WriteString(" " + i.Op + " ")
	b.WriteString(i.Right.String())
	b.WriteString(")")

	return b.String()
}

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (i *IfExpression) e() {}

func (i *IfExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *IfExpression) String() string {
	var b bytes.Buffer

	b.WriteString("if ")
	b.WriteString(i.Condition.String())
	b.WriteString(" ")
	b.WriteString(i.Consequence.String())

	if i.Alternative != nil {
		b.WriteString(" else ")
		b.WriteString(i.Alternative.String())
	}

	return b.String()
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (b *BlockStatement) s() {}

func (b *BlockStatement) TokenLiteral() string {
	return b.Token.Literal
}

func (bs *BlockStatement) String() string {
	var b bytes.Buffer

	for _, s := range bs.Statements {
		b.WriteString(s.String())
	}

	return b.String()
}

type Fn struct {
	Token  token.Token
	Params []*Id
	Body   *BlockStatement
}

func (f *Fn) e() {}

func (f *Fn) TokenLiteral() string {
	return f.Token.Literal
}

func (f *Fn) String() string {
	var b bytes.Buffer

	params := []string{}
	for _, p := range f.Params {
		params = append(params, p.String())
	}

	b.WriteString(f.TokenLiteral())
	b.WriteString("(")
	b.WriteString(strings.Join(params, ", "))
	b.WriteString(")")
	b.WriteString(f.Body.String())

	return b.String()
}

type Call struct {
	Token token.Token
	Fn    Expression
	Args  []Expression
}

func (c *Call) e() {}

func (c *Call) TokenLiteral() string {
	return c.Token.Literal
}

func (c *Call) String() string {
	var b bytes.Buffer

	args := []string{}
	for _, a := range c.Args {
		args = append(args, a.String())
	}

	b.WriteString(c.Fn.String())
	b.WriteString("(")
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(")")

	return b.String()
}

type Index struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (c *Index) e() {}

func (c *Index) TokenLiteral() string {
	return c.Token.Literal
}

func (c *Index) String() string {
	var b bytes.Buffer

	b.WriteString("(")
	b.WriteString(c.Left.String())
	b.WriteString("[")
	b.WriteString(c.Index.String())
	b.WriteString("])")

	return b.String()
}

type Hash struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (h *Hash) e() {}

func (h *Hash) TokenLiteral() string {
	return h.Token.Literal
}

func (h *Hash) String() string {
	var b bytes.Buffer

	pairs := []string{}
	for k, v := range h.Pairs {
		pairs = append(pairs, k.String()+": "+v.String())
	}

	b.WriteString("{")
	b.WriteString(strings.Join(pairs, ", "))
	b.WriteString("}")

	return b.String()
}
