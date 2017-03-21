package eval

import (
	"fmt"
	"strings"

	"geo/ast"
	"geo/lexer"
	"geo/parser"
)

func Compile(input string) (*ast.Module, error) {
	l := lexer.New(strings.NewReader(input))
	p := parser.New(l)
	m, errs := p.Parse()
	if len(errs) > 0 {
		return nil, fmt.Errorf("%v", errs)
	}
	return m, nil
}
