// Copyright 2020 Pavel Knoblokh. All rights reserved.
// Use of this source code is governed by MIT License
// that can be found in the LICENSE file.
// nolint: govet
package exprcalc

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/repr"
)

var Debug = false

type (
	Evaluable interface {
		Eval(ctx *Context) (interface{}, error)
	}

	Gettable interface {
		// Currently must return a number, string or bool
		GetByName(string) (interface{}, error)
	}

	Boolean bool
)

func (b *Boolean) Capture(values []string) error {
	*b = strings.ToLower(values[0]) == "true"
	return nil
}

type Expression struct {
	Pos lexer.Position

	Or []*OrCondition `@@ { "OR" @@ }`
}

type OrCondition struct {
	Pos lexer.Position

	And []*ConditionOperand `@@ { "AND" @@ }`
}

type ConditionOperand struct {
	Pos lexer.Position

	Term    *Term    `@@`
	Compare *Compare `[ @@ ]`
}

type Compare struct {
	Pos lexer.Position

	Operator string `@( "<=" | ">=" | "==" | "<" | ">" | "!=" )`
	Term     *Term  `( @@ )`
}

type Term struct {
	Pos lexer.Position

	Value         *Value      `@@`
	Identifier    *string     `| @Ident`
	SubExpression *Expression `| "(" @@ ")"`
}

type Value struct {
	Pos lexer.Position

	Number  *float64 `(  @Number`
	String  *string  ` | @String`
	Boolean *Boolean ` | @Boolean )`
}

type Context struct {
	Object Gettable
}

// Returns a float64, string or bool
func Eval(expr string, obj Gettable) (interface{}, error) {
	if len(expr) == 0 {
		return nil, nil
	}

	e, err := Parse(expr)
	if err != nil {
		return nil, err
	}

	if Debug {
		repr.Println(e, repr.Indent("      "), repr.OmitEmpty(true))
	}

	return EvalParsed(e, obj)
}

func Parse(expr string) (*Expression, error) {
	e := &Expression{}
	err := Parser.ParseString(expr, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func EvalParsed(expr *Expression, obj Gettable) (interface{}, error) {
	ctx := &Context{obj}

	value, err := expr.Eval(ctx)
	if err != nil {
		return nil, err
	}

	return castToExternal(value), nil
}

func (e *Expression) Eval(ctx *Context) (interface{}, error) {
	if len(e.Or) == 0 {
		return nil, nil
	}

	lhs, err := e.Or[0].Eval(ctx)
	if err != nil {
		return nil, err
	}

	if len(e.Or) == 1 {
		return lhs, nil
	}

	if len(e.Or) == 2 { // short-circuit OR
		if lhsBool, ok := lhs.(Boolean); ok && bool(lhsBool) {
			return lhs, nil
		}
	}

	for _, or := range e.Or[1:] {
		lhsBool, rhsBool, err := evaluateBooleans(ctx, lhs, or)
		if err != nil {
			return nil, lexer.Errorf(e.Pos, "%v", err)
		}
		if lhsBool { // short-circuit OR
			break
		}
		lhs = lhsBool || rhsBool
	}

	return lhs, nil
}

func (o *OrCondition) Eval(ctx *Context) (interface{}, error) {
	if len(o.And) == 0 {
		return nil, nil
	}

	lhs, err := o.And[0].Eval(ctx)
	if err != nil {
		return nil, err
	}

	if len(o.And) == 1 {
		return lhs, nil
	}

	if len(o.And) == 2 { // short-circuit AND
		if lhsBool, ok := lhs.(Boolean); ok && !bool(lhsBool) {
			return lhs, nil
		}
	}

	for _, and := range o.And[1:] {
		lhsBool, rhsBool, err := evaluateBooleans(ctx, lhs, and)
		if err != nil {
			return nil, lexer.Errorf(o.Pos, "%v", err)
		}
		if !lhsBool { // short-circuit AND
			break
		}
		lhs = lhsBool && rhsBool
	}

	return lhs, nil
}

func (c *ConditionOperand) Eval(ctx *Context) (interface{}, error) {
	lhs, err := c.Term.Eval(ctx)
	if err != nil {
		return nil, err
	}

	if c.Compare == nil {
		return lhs, nil
	}

	res, err := c.Compare.Eval(ctx, lhs)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Compare) Eval(ctx *Context, lhs interface{}) (interface{}, error) {
	rhs, err := c.Term.Eval(ctx)
	if err != nil {
		return nil, err
	}

	switch lhs := lhs.(type) {
	case float64:
		rhs, ok := rhs.(float64)
		if !ok {
			return nil, lexer.Errorf(c.Pos, "rhs of %s must be a number", c.Operator)
		}
		switch c.Operator {
		case "==":
			return Boolean(lhs == rhs), nil
		case "!=":
			return Boolean(lhs != rhs), nil
		case "<":
			return Boolean(lhs < rhs), nil
		case ">":
			return Boolean(lhs > rhs), nil
		case "<=":
			return Boolean(lhs <= rhs), nil
		case ">=":
			return Boolean(lhs >= rhs), nil
		default:
			return nil, lexer.Errorf(c.Pos, "unsupported number comparison operator %s", c.Operator)
		}
	case string:
		rhs, ok := rhs.(string)
		if !ok {
			return nil, lexer.Errorf(c.Pos, "rhs of %s must be a string", c.Operator)
		}
		switch c.Operator {
		case "==":
			return Boolean(lhs == rhs), nil
		case "!=":
			return Boolean(lhs != rhs), nil
		case "<":
			return Boolean(lhs < rhs), nil
		case ">":
			return Boolean(lhs > rhs), nil
		case "<=":
			return Boolean(lhs <= rhs), nil
		case ">=":
			return Boolean(lhs >= rhs), nil
		default:
			return nil, lexer.Errorf(c.Pos, "unsupported string comparison operator %s", c.Operator)
		}
	case Boolean:
		rhs, ok := rhs.(Boolean)
		if !ok {
			return nil, lexer.Errorf(c.Pos, "rhs of %s must be boolean", c.Operator)
		}
		switch c.Operator {
		case "==":
			return Boolean(lhs == rhs), nil
		case "!=":
			return Boolean(lhs != rhs), nil
		default:
			return nil, lexer.Errorf(c.Pos, "unsupported boolean comparison operator %s", c.Operator)
		}
	default:
		return nil, lexer.Errorf(c.Pos, "lhs of %s must be a number, string or boolean", c.Operator)
	}
	panic("unreachable")
}

func (t *Term) Eval(ctx *Context) (interface{}, error) {
	switch {
	case t.Value != nil:
		return t.Value.Eval(ctx)
	case t.Identifier != nil:
		if ctx.Object == nil {
			return nil, lexer.Errorf(t.Pos, "Identifier %v on nil object", t.Identifier)
		}
		value, err := ctx.Object.GetByName(*t.Identifier)
		if err != nil {
			return nil, lexer.Errorf(t.Pos, "%v", err)
		}
		return castToInternal(value), nil
	case t.SubExpression != nil:
		return t.SubExpression.Eval(ctx)
	}
	panic("unsupported term type" + repr.String(t))
}

func (v *Value) Eval(ctx *Context) (interface{}, error) {
	switch {
	case v.Number != nil:
		return *v.Number, nil
	case v.String != nil:
		return *v.String, nil
	case v.Boolean != nil:
		return *v.Boolean, nil
	}
	panic("unsupported value type" + repr.String(v))
}

func castToInternal(value interface{}) interface{} {
	switch value := value.(type) {
	case bool:
		return Boolean(value)
	case int:
		return float64(value)
	case int8:
		return float64(value)
	case int16:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case uint:
		return float64(value)
	case uint8:
		return float64(value)
	case uint16:
		return float64(value)
	case uint32:
		return float64(value)
	case uint64:
		return float64(value)
	case float32:
		return float64(value)
	}
	return value
}

func castToExternal(value interface{}) interface{} {
	switch value := value.(type) {
	case Boolean:
		return bool(value)
	}
	return value
}

func evaluateBooleans(ctx *Context, lhs interface{}, rhsExpr Evaluable) (Boolean, Boolean, error) {
	rhs, err := rhsExpr.Eval(ctx)
	if err != nil {
		return false, false, err
	}
	lhsBool, ok := lhs.(Boolean)
	if !ok {
		return false, false, fmt.Errorf("lhs must be boolean")
	}
	rhsBool, ok := rhs.(Boolean)
	if !ok {
		return false, false, fmt.Errorf("rhs must be boolean")
	}
	return lhsBool, rhsBool, nil
}

var (
	myLexer = lexer.Must(lexer.Regexp(`(\s+)` +
		`|(?P<LogicOp>(?i)AND|OR)` +
		`|(?P<Boolean>(?i)true|false)` +
		`|(?P<Ident>[a-zA-Z_][a-zA-Z0-9_]*)` +
		`|(?P<Number>[-+]?\d*\.?\d+([eE][-+]?\d+)?)` +
		`|(?P<String>'[^']*'|"[^"]*")` +
		`|(?P<CompareOp>!=|<=|>=|==|[()<>])`,
	))
	Parser = participle.MustBuild(
		&Expression{},
		participle.Lexer(myLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("LogicOp", "Boolean"),
	)
)
