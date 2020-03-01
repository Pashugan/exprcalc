# Expression calculator

## Overview

A simple embedded expression calculator.

## Usage

```go
type Person struct {
	dob     time.Time
	city    string
	married bool
}

func (p *Person) Age() int {
	return int(math.Floor(time.Since(p.dob).Hours() / 24 / 365))
}

// Implements exprcalc.Gettable
func (p *Person) GetByName(name string) (interface{}, error) {
	switch strings.ToLower(name) {
	case "age":
		return p.Age(), nil
	case "city":
		return p.city, nil
	case "married":
		return p.married, nil
	}
	return nil, fmt.Errorf("Getter '%v' not found", name)
}

expr := `(city == "Massachusetts" OR city == "Berkeley") AND age > 23 AND married == true`

unix := &Person{
	dob: time.Unix(0, 0),
	city: "Berkeley",
	married: true,
}
value, err := exprcalc.Eval(expr, unix)

// or preparsed flavour

parsed, err := exprcalc.Parse(expr)
value, err = exprcalc.EvalParsed(parsed, unix)
```

will return `true` as the value.

## Description

#### Operands

number (e.g. `1984`, `3.14`, `–273.15`, `6.62607004e-34`)
string (e.g. `"foo"`, `'bar'`)
bool (case-insensitive `true` and `false`)
identifier, see details below (case-sensitive, C-variable-like names, i.e. starting from an alphabetic symbol or underscore followed by alphabetic symbols, digits or underscore, e.g. `foo`, `BAR`, `_myvar`, `__my_var`, `Moon44`)

#### Operators

##### Comparison
number  `== != < > <= >=`
string  `== != < > <= >=`
boolean `== !=`

##### Logical
`AND OR` (case-insensitive)

##### Subexpression
Parentheses, as in maths
`( )`, e.g. `(true OR false) AND true`

#### Operator precedence

`( )` (highest)
`== != < > <= >=`
`AND`
`OR` (lowest)

Implicit type casting is not supported, i.e. both operands must be of the same type. Both operands of logical operators must be boolean.

##### Caveat
Short-circuit logical expression evaluation can result in correct evaluation of expressions with incompatible types. This behaviour should be expected as it trades performance against possible bug prone code. For instance, `true OR "String"` and `false AND 123` will not fail because only the first operand is required to evaluate the result of the logical expression.

#### Identifiers

When you pass a context object implementing `exprcalc.Gettable` interface, it is possible to use identifiers in expressions, e.g. `age > 23`. The identifier is just a string that is passed into the GetByName() method of the object, that usually returns object field values or call its methods. Obviously, it must return one of the supported data types, i.e. any numeric type, string or boolean. Since using the interface instead of reflection, the overhead of the call is very small.
