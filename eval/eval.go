package eval

import (
	"fmt"

	"geo/ast"
	"geo/object"
)

var (
	True  = object.NewBool(true)
	False = object.NewBool(false)
	Null  = &object.Null{}
)

func (c *Context) internalEval(node ast.Node, scope *object.Scope) object.Object {
	switch node := node.(type) {
	case *ast.Number:
		return object.NewNumber(node.Value)

	case *ast.Bool:
		return c.nativeBoolToObject(node.Value)

	case *ast.String:
		return object.NewString(node.Value)

	case *ast.Array:
		elms := c.evalExpressions(node.Elements, scope)
		if len(elms) == 1 && isError(elms[0]) {
			return elms[0]
		}
		return &object.Array{elms}

	case *ast.Hash:
		return c.evalHashExpression(node, scope)

	case *ast.Index:
		left := c.internalEval(node.Left, scope)
		if isError(left) {
			return left
		}
		index := c.internalEval(node.Index, scope)
		if isError(index) {
			return index
		}
		return c.evalIndexExpression(left, index)

	case *ast.PrefixExpression:
		right := c.internalEval(node.Right, scope)
		if isError(right) {
			return right
		}
		return c.evalPrefix(node.Op, right)

	case *ast.InfixExpression:
		left := c.internalEval(node.Left, scope)
		if isError(left) {
			return left
		}
		right := c.internalEval(node.Right, scope)
		if isError(right) {
			return right
		}
		return c.evalInfix(node.Op, left, right)

	case *ast.BlockStatement:
		return c.evalBlock(node, scope)

	case *ast.IfExpression:
		return c.evalIfExpression(node, scope)

	case *ast.LetStatement:
		val := c.internalEval(node.Value, scope)
		if isError(val) {
			return val
		}
		scope.Set(node.Name.Value, val)

	case *ast.Id:
		return c.evalIdExpression(node, scope)

	case *ast.ReturnStatement:
		val := c.internalEval(node.Value, scope)
		if isError(val) {
			return val
		}
		return &object.Return{val}

	case *ast.ExpressionStatement:
		return c.internalEval(node.Expression, scope)

	case *ast.Fn:
		return &object.Fn{Params: node.Params, Body: node.Body, Scope: scope}

	case *ast.Call:
		fn := c.internalEval(node.Fn, scope)
		if isError(fn) {
			return fn
		}
		args := c.evalExpressions(node.Args, scope)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return c.applyFn(fn, args...)

	case *ast.Module:
		return c.evalModule(node.Statements, scope)
	}
	return nil
}

func (c *Context) evalModule(stmts []ast.Statement, scope *object.Scope) object.Object {
	var result object.Object

	for _, s := range stmts {
		result = c.internalEval(s, scope)

		switch result := result.(type) {
		case *object.Return:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func (c *Context) evalBlock(block *ast.BlockStatement, scope *object.Scope) object.Object {
	var result object.Object

	isolated := object.NewScope(scope)
	for _, s := range block.Statements {
		result = c.internalEval(s, isolated)

		if result != nil {
			ty := result.Type()
			if ty == object.TypeReturn || ty == object.TypeError {
				return result
			}
		}
	}

	return result
}

func (c *Context) evalExpressions(exps []ast.Expression, scope *object.Scope) []object.Object {
	var result []object.Object

	for _, exp := range exps {
		e := c.internalEval(exp, scope)
		if isError(e) {
			return []object.Object{e}
		}
		result = append(result, e)
	}

	return result
}

func (c *Context) applyFn(fn object.Object, args ...object.Object) object.Object {
	// NOTE: How to avoid infinite recursion / call stack?

	switch fn := fn.(type) {
	case *object.Fn:
		min := len(fn.Params)
		if len(args) < min {
			min = len(args)
		}

		scope := object.NewScope(fn.Scope)
		// NOTE: What to do when there are more args than params (...args?)
		for i := 0; i < min; i++ {
			scope.Set(fn.Params[i].Value, args[i])
		}

		// When there is less args than params, return a new function
		if min < len(fn.Params) {
			restParams := fn.Params[min:len(fn.Params)]
			return &object.Fn{Params: restParams, Body: fn.Body, Scope: scope}
		}

		result := c.internalEval(fn.Body, scope)
		if ret, ok := result.(*object.Return); ok {
			return ret.Value
		}
		return result

	case *object.Builtin:
		if len(args) < len(fn.Params) {
			return newError("wrong number of arguments. got=%d, want=1", len(args))
		}

		for i, a := range fn.Params {
			if a&args[i].Type() == 0 {
				return newError("argument to `%s` must be (%s), got %s", fn.Name, object.ObjectTypesToString(a), object.ObjectTypeToString(args[i].Type()))
			}
		}

		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func (c *Context) nativeBoolToObject(v bool) object.Object {
	if v {
		return True
	}
	return False
}

func (c *Context) evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.TypeArray && index.Type() == object.TypeNumber:
		return c.evalArrayIndexExpression(left, index)

	case left.Type() == object.TypeHash:
		return c.evalHashIndexExpression(left, index)

	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func (c *Context) evalArrayIndexExpression(left, index object.Object) object.Object {
	ary := left.(*object.Array)
	max := int64(len(ary.Elements) - 1)
	idx := int64(index.(*object.Number).Value)
	if idx < 0 || idx > max {
		return Null
	}
	return ary.Elements[idx]
}

func (c *Context) evalHashIndexExpression(left, index object.Object) object.Object {
	hash := left.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hash.Pairs[key.HashKey()]
	if ok {
		return pair.Value
	}

	return Null
}

func (c *Context) evalPrefix(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return c.evalNotExpression(right)

	case "-":
		return c.evalMinusExpression(right)

	default:
		return newError("unknown operator: %s%s", op, right.String())
	}
}

func (c *Context) evalNotExpression(right object.Object) object.Object {
	switch right {
	case True:
		return False

	case False:
		return True

	case Null:
		return True

	default:
		if isTruthy(right) {
			return False
		} else {
			return True
		}
	}
}

func (c *Context) evalMinusExpression(right object.Object) object.Object {
	if right.Type() != object.TypeNumber {
		return newError("unknown operator: -%s", right.Type())
	}

	val := right.(*object.Number).Value
	return object.NewNumber(-val)
}

func (c *Context) evalInfix(op string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.TypeNumber && right.Type() == object.TypeNumber:
		return c.evalNumberExpression(op, left, right)

	case left.Type() == object.TypeString && right.Type() == object.TypeString:
		return c.evalStringExpression(op, left, right)

	case op == "==":
		return c.nativeBoolToObject(left == right)

	case op == "!=":
		return c.nativeBoolToObject(left != right)

	case op == "|":
		return c.applyFn(right, left)

	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())

	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func (c *Context) evalNumberExpression(op string, left, right object.Object) object.Object {
	leftVal := left.(*object.Number).Value
	rightVal := right.(*object.Number).Value

	switch op {
	case "+":
		return object.NewNumber(leftVal + rightVal)

	case "-":
		return object.NewNumber(leftVal - rightVal)

	case "*":
		return object.NewNumber(leftVal * rightVal)

	case "/":
		return object.NewNumber(leftVal / rightVal)

	case ">":
		return c.nativeBoolToObject(leftVal > rightVal)

	case ">=":
		return c.nativeBoolToObject(leftVal >= rightVal)

	case "<":
		return c.nativeBoolToObject(leftVal < rightVal)

	case "<=":
		return c.nativeBoolToObject(leftVal <= rightVal)

	case "==":
		return c.nativeBoolToObject(leftVal == rightVal)

	case "!=":
		return c.nativeBoolToObject(leftVal != rightVal)

	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func (c *Context) evalStringExpression(op string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch op {
	case "+":
		return object.NewString(leftVal + rightVal)

	// NOTE: Maybe... case "*": return object.NewNumber(leftVal * rightVal)

	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func (c *Context) evalHashExpression(node *ast.Hash, scope *object.Scope) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for k, v := range node.Pairs {
		key := c.internalEval(k, scope)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := c.internalEval(v, scope)
		if isError(value) {
			return value
		}

		pairs[hashKey.HashKey()] = object.HashPair{key, value}
	}

	return &object.Hash{pairs}
}

func (c *Context) evalIdExpression(node *ast.Id, scope *object.Scope) object.Object {
	if val, ok := scope.Get(node.Value); ok {
		return val
	}
	if builtin, ok := c.builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
}

func (c *Context) evalIfExpression(node *ast.IfExpression, scope *object.Scope) object.Object {
	cond := c.internalEval(node.Condition, scope)
	if isError(cond) {
		return cond
	}
	if isTruthy(cond) {
		return c.internalEval(node.Consequence, scope)
	} else if node.Alternative != nil {
		return c.internalEval(node.Alternative, scope)
	} else {
		return Null
	}
}

func isTruthy(v object.Object) bool {
	switch {
	case v == True:
		return true

	case v == False || v == Null:
		return false

	default:
		if num, ok := v.(*object.Number); ok && num.Value == 0 {
			return false
		}
		if str, ok := v.(*object.String); ok && str.Value == "" {
			return false
		}
		if ary, ok := v.(*object.Array); ok && len(ary.Elements) == 0 {
			return false
		}
		if hsh, ok := v.(*object.Hash); ok && len(hsh.Pairs) == 0 {
			return false
		}
	}
	return true
}

func newError(msg string, a ...interface{}) object.Object {
	return &object.Error{fmt.Errorf(msg, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.TypeNull
	}
	return false
}
