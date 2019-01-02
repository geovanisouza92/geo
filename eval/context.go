package eval

import (
	"github.com/geovanisouza92/geo/ast"
	"github.com/geovanisouza92/geo/object"
)

type Context struct {
	registry *ModuleRegistry
	builtins map[string]*object.Builtin
	scope    *object.Scope
}

func NewContext(scope *object.Scope) *Context {
	c := &Context{registry: newModuleRegistry(), scope: scope}
	c.builtins = builtins
	c.builtins["import"] = &object.Builtin{
		Impl: func(args ...object.Object) object.Object {
			// TODO

			return nil
		},
	}
	return c
}

func (c *Context) Eval(m *ast.Module) object.Object {
	return c.internalEval(m, c.scope)
}
