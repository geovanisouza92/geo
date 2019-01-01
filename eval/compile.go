package eval

import (
	"fmt"
	"strings"

	"github.com/geovanisouza92/geo/ast"
	"github.com/geovanisouza92/geo/lexer"
	"github.com/geovanisouza92/geo/parser"
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
