// Copyright 2020 Pavel Knoblokh. All rights reserved.
// Use of this source code is governed by MIT License
// that can be found in the LICENSE file.
package exprcalc

import (
	"fmt"
	"testing"
)

type testPerson struct {
	gender  string
	age     int
	married bool
}

func (o *testPerson) GetByName(name string) (interface{}, error) {
	switch name {
	case "gender":
		return o.gender, nil
	case "age":
		return o.age, nil
	case "married":
		return o.married, nil
	}
	return nil, fmt.Errorf("Invalid identifier")
}

func TestParserErr(t *testing.T) {
	tests := []struct {
		name string
		expr string
		obj  Gettable
	}{
		{
			"Invalid tokens",
			`true || false && true`,
			nil,
		},
		{
			"Missing bracket",
			`true AND ( false OR true`,
			nil,
		},
		{
			"Empty subexpression",
			`()`,
			nil,
		},
		{
			"Invalid type comparison",
			`"asdf" > 1234`,
			nil,
		},
		{
			"Identifier on nil object",
			`gender > 18`,
			nil,
		},
		{
			"Invalid identifier",
			`height == 182`,
			&testPerson{"male", 22, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Eval(tt.expr, tt.obj)

			if err == nil {
				t.Error("Must fail")
			}
		})
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want interface{}
	}{
		// Literals
		{
			"Empty expression is nil",
			``,
			nil,
		},
		{
			"Number literal is itself",
			`0`,
			0.0,
		},
		{
			"String literal in double quotes is itself",
			`"asdf"`,
			"asdf",
		},
		{
			"String literal in single quotes is itself",
			`'asdf'`,
			"asdf",
		},
		{
			"True literal is true",
			`true`,
			true,
		},
		{
			"False literal is false",
			`false`,
			false,
		},
		// Number comparison
		{
			"Number comparison #1",
			`0 == 0`,
			true,
		},
		{
			"Number comparison #2",
			`0 != 0`,
			false,
		},
		{
			"Number comparison #3",
			`1 == 0`,
			false,
		},
		{
			"Number comparison #4",
			`1 != 0`,
			true,
		},
		{
			"Number comparison #5",
			`1234 > 234`,
			true,
		},
		{
			"Number comparison #6",
			`234 < 1234`,
			true,
		},
		{
			"Number comparison #7",
			`1234 >= 234`,
			true,
		},
		{
			"Number comparison #8",
			`1234 >= 1234`,
			true,
		},
		{
			"Number comparison #9",
			`234 <= 1234`,
			true,
		},
		{
			"Number comparison #10",
			`1234 <= 1234`,
			true,
		},
		{
			"Number comparison #11",
			`234 > 1234`,
			false,
		},
		{
			"Number comparison #12",
			`1234 < 234`,
			false,
		},
		{
			"Number comparison #13",
			`234 >= 1234`,
			false,
		},
		{
			"Number comparison #14",
			`1234 <= 234`,
			false,
		},
		{
			"Number comparison #15",
			`-1 > -2`,
			true,
		},
		{
			"Number comparison #16",
			`-1.25 <= -1.100000`,
			true,
		},
		{
			"Number comparison #17",
			`+10 > +2`,
			true,
		},
		{
			"Number comparison #18",
			`12.34 >= 1.234`,
			true,
		},
		{
			"Number comparison #19",
			`-1.234e9 == -1234000000`,
			true,
		},
		// String comparison
		{
			"String comparison #1",
			`"asdf" == "asdf"`,
			true,
		},
		{
			"String comparison #2",
			`"asdf" != "ASDF"`,
			true,
		},
		{
			"String comparison #3",
			`"a" < "bcd"`,
			true,
		},
		{
			"String comparison #4",
			`"2020-01-01" >= "2018-09-22"`,
			true,
		},
		// Boolean comparison
		{
			"Boolean comparison #1",
			`true == true`,
			true,
		},
		{
			"Boolean comparison #2",
			`false == false`,
			true,
		},
		{
			"Boolean comparison #3",
			`true != false`,
			true,
		},
		// Basic logic operations
		{
			"True OR",
			`true OR true`,
			true,
		},
		{
			"True left OR",
			`true OR false`,
			true,
		},
		{
			"True right OR",
			`false OR true`,
			true,
		},
		{
			"False OR",
			`false OR false`,
			false,
		},
		{
			"True AND",
			`true AND true`,
			true,
		},
		{
			"False AND",
			`false AND false`,
			false,
		},
		{
			"False left AND",
			`false AND true`,
			false,
		},
		{
			"False right AND",
			`true AND false`,
			false,
		},
		// Logic operator precedence
		{
			"Logic operator precedence #1",
			`false OR true AND true`,
			true,
		},
		{
			"Logic operator precedence #2",
			`true OR false AND true`,
			true,
		},
		{
			"Logic operator precedence #3",
			`true AND true OR false`,
			true,
		},
		{
			"Logic operator precedence #4",
			`false AND true OR true`,
			true,
		},
		{
			"Logic operator precedence #5",
			`false AND (true OR false)`,
			false,
		},
		{
			"Logic operator precedence #6",
			`false AND true OR true`,
			true,
		},
		{
			"Logic operator precedence #7",
			`false AND (true OR true)`,
			false,
		},
		// Subexpressions
		{
			"Value in subexpression is itself",
			`(3.14)`,
			3.14,
		},
		{
			"Valid subexpression",
			`(false OR false)`,
			false,
		},
		{
			"Valid subexpression",
			`(true AND true)`,
			true,
		},
		// Case insensitive operators
		{
			"Case insensitive operators #1",
			`TRUE and trUE oR FalsE`,
			true,
		},
		{
			"Case insensitive operators #2",
			`(TRue Or fALSe) and FALSE`,
			false,
		},
		// Short-circuit evaluation
		{
			"Short-circuit OR evaluation",
			`true OR "String"`,
			true,
		},
		{
			"Short-circuit AND evaluation",
			`false AND 123`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := Eval(tt.expr, nil)

			if err != nil {
				t.Errorf("Eval error '%v'", err)
			}

			if value != tt.want {
				t.Errorf("Value error: want '%v', got '%v'", tt.want, value)
			}
		})
	}
}

func TestParserWithContext(t *testing.T) {
	tests := []struct {
		name string
		expr string
		obj  Gettable
		want interface{}
	}{
		{
			"Valid identifier",
			`gender == "male"`,
			&testPerson{"male", 22, false},
			true,
		},
		{
			"Boolean identifier",
			`false OR married`,
			&testPerson{"male", 22, true},
			true,
		},
		{
			"Valid complex expression",
			`( ( gender == "male" OR (gender != "female") ) AND age < 18 AND age <= 18 AND age == 18 AND age >= 18 AND age > 18 AND married == true OR (25.25 == age OR married == FALSE) )`,
			&testPerson{"male", 22, true},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := Eval(tt.expr, tt.obj)

			if err != nil {
				t.Errorf("Eval error '%v'", err)
			}

			if value != tt.want {
				t.Errorf("Value error: want '%v', got '%v'", tt.want, value)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	expr := `(1 == 1 OR 1 == 0 OR 1 == 0 OR 1 == 0 OR 1 == 0) AND 1 == 0 AND 1 == 0 AND 1 == 0 AND 1 == 0`
	e, _ := Parse(expr)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EvalParsed(e, nil)
	}
}
