package eval

import (
	"errors"

	"geo/ast"
)

var errNotFound = errors.New("module not found")

type ModuleRegistry struct {
	cache     map[string]*ast.Module
	resolvers []ModuleResolver
}

func newModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		cache: make(map[string]*ast.Module),
		resolvers: []ModuleResolver{
			&FileResolver{},
		},
	}
}

func (mr *ModuleRegistry) Load(name string) (*ast.Module, error) {
	m, ok := mr.cache[name]
	if ok {
		return m, nil
	}

	for _, r := range mr.resolvers {
		if in := r.Resolve(name); in != "" {
			m, err := Compile(in)
			if err != nil {
				return nil, err
			}

			mr.cache[name] = m
			return m, nil
		}
	}

	return nil, errNotFound
}
