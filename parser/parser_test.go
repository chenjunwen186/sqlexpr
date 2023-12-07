package parser

import (
	"testing"

	"github.com/chenjunwen186/sqlexpr/ast"
	"github.com/chenjunwen186/sqlexpr/lexer"
	"github.com/chenjunwen186/sqlexpr/token"
)

func parseExpression(t *testing.T, input string) ast.Expression {
	l := lexer.New(input)
	p := New(l)
	r, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("parseExpression() failed: %s", err)
	}

	return r
}

func TestIdentifierExpression(t *testing.T) {
	input := `
	hello  `
	expr := parseExpression(t, input)
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("parseExpression() failed: expected *ast.Identifier, got %T", expr)
	}

	if ident.Value != "hello" {
		t.Errorf("parseExpression() failed: expected ident.Value=%q, got %q", "hello", ident.Value)
	}
}

func TestNumberLiteralExpression(t *testing.T) {
	type TestCase struct {
		input    string
		expected string
	}
	inputs := []TestCase{
		{"  123  ", "123"},
		{"  123.456 \r ", "123.456"},
		{" \t.123 \n \r", ".123"},
	}
	for _, v := range inputs {
		expr := parseExpression(t, v.input)
		num, ok := expr.(*ast.NumberLiteral)
		if !ok {
			t.Fatalf("parseExpression() failed: expected *ast.NumberLiteral, got %T", expr)
		}

		if num.Type != token.NUMBER {
			t.Errorf("parseExpression() failed: expected num.Token=%q, got %q", token.NUMBER, num.Token)
		}

		if num.Literal != v.expected {
			t.Errorf("parseExpression() failed: expected num.Literal=%q, got %q", v.expected, num.Literal)
		}
	}
}

func TestNullLiteral(t *testing.T) {
	input := `
	null  `
	expr := parseExpression(t, input)
	null, ok := expr.(*ast.NullLiteral)
	if !ok {
		t.Fatalf("parseExpression() failed: expected *ast.NullLiteral, got %T", expr)
	}

	if null.Type != token.NULL {
		t.Errorf("parseExpression() failed: expected null.Token=%q, got %q", token.NULL, null.Token)
	}

	if null.Literal != "null" {
		t.Errorf("parseExpression() failed: expected null.Literal=%q, got %q", "null", null.Literal)
	}
}

func TestBooleanLiteral(t *testing.T) {
	type TestCase struct {
		input    string
		expected string
		value    bool
	}

	inputs := []TestCase{
		{"  true  ", "true", true},
		{"  false \r ", "false", false},
		{" \tTrue \n \r", "True", true},
		{" \tFALSE \n \r", "FALSE", false},
	}
	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		boolean, ok := expr.(*ast.BooleanLiteral)
		if !ok {
			t.Fatalf("parseExpression() failed: expected *ast.BooleanLiteral, got %T", expr)
		}

		if input.value {
			if boolean.Type != token.TRUE {
				t.Errorf("parseExpression() failed: expected boolean.Token=%q, got %q", token.TRUE, boolean.Token)
			}
		} else {
			if boolean.Type != token.FALSE {
				t.Errorf("parseExpression() failed: expected boolean.Token=%q, got %q", token.FALSE, boolean.Token)
			}
		}

		if boolean.Literal != input.expected {
			t.Errorf("parseExpression() failed: expected boolean.Literal=%q, got %q", input.expected, boolean.Literal)
		}
	}
}
