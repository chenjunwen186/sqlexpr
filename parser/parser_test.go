package parser

import (
	"strconv"
	"testing"

	"github.com/chenjunwen186/sqlexpr/ast"
	"github.com/chenjunwen186/sqlexpr/lexer"
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
	type TestCase struct {
		input    string
		expected string
	}

	inputs := []TestCase{
		{"\r  hello\t\n", "hello"},
		{"\r  hello_world\t\n  ", "hello_world"},
		{"\r  hello_world123\t\n  ", "hello_world123"},
	}
	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		testIdentifier(t, expr, input.expected)
	}
}

func TestNumberLiteralExpression(t *testing.T) {
	type TestCase struct {
		input    string
		expected any
	}
	inputs := []TestCase{
		{"  123  ", 123},
		{"  123.456 \r ", 123.456},
		{" \t.123 \n \r", .123},
	}
	for _, v := range inputs {
		expr := parseExpression(t, v.input)
		testLiteralExpression(t, expr, v.expected)
	}
}

func TestNullLiteral(t *testing.T) {
	input := `
	null  `
	expr := parseExpression(t, input)
	testLiteralExpression(t, expr, nil)
}

func TestBooleanLiteral(t *testing.T) {
	type TestCase struct {
		input    string
		expected bool
	}

	inputs := []TestCase{
		{"true  ", true},
		{"  false \r ", false},
		{" \tTrue \n \r", true},
		{" \tFALSE \n \r", false},
	}
	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		testLiteralExpression(t, expr, input.expected)
	}
}

func TestEmptyInput(t *testing.T) {
	input := ``
	expr := parseExpression(t, input)
	if expr != nil {
		t.Errorf("parseExpression() failed: expected nil, got %T", expr)
	}
}

func TestPrefixExpression(t *testing.T) {
	type TestCase struct {
		input    string
		operator string
		right    any
		str      string
	}

	inputs := []TestCase{
		{"-123", "-", 123, "(-123)"},
		{"+123.456", "+", 123.456, "(+123.456)"},
		{"DISTINCT hello", "DISTINCT", "hello", "(DISTINCT hello)"},
	}
	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		testPrefixExpression(t, expr, input.operator, input.right)
		if expr.String() != input.str {
			t.Errorf("expr.String() not %q, got %q", input.str, expr.String())
		}
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier, got %T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %q, got %q", value, ident.Value)
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	v, ok := exp.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("exp not *ast.BooleanLiteral, got %T", exp)
		return false
	}
	if v.Value() != value {
		t.Errorf("v.Value() not %t, got %t", value, v.Value())
		return false
	}
	return true
}

func testNumberLiteral(t *testing.T, exp ast.Expression, expected any) bool {
	v, ok := exp.(*ast.NumberLiteral)
	if !ok {
		t.Errorf("exp not *ast.NumberLiteral, got %T", exp)
		return false
	}
	switch exp := expected.(type) {
	case int:
		i, err := strconv.ParseInt(v.Literal, 10, 64)
		if err != nil {
			t.Errorf("strconv.ParseInt(%q) failed: %s", v.Literal, err)
			return false
		}
		if i != int64(exp) {
			t.Errorf("i not %d, got %d", exp, i)
			return false
		}

		return true
	case float64:
		f, err := strconv.ParseFloat(v.Literal, 64)
		if err != nil {
			t.Errorf("strconv.ParseFloat(%q) failed: %s", v.Literal, err)
			return false
		}
		if f != exp {
			t.Errorf("f not %f, got %f", exp, f)
			return false
		}

		return true
	}

	t.Errorf("type of expected not handled, got %T", expected)
	return false
}

func testNullLiteral(t *testing.T, exp ast.Expression) bool {
	_, ok := exp.(*ast.NullLiteral)
	if !ok {
		t.Errorf("exp not *ast.NullLiteral, got %T", exp)
		return false
	}
	return true
}

func testLiteralExpression(t *testing.T, expr ast.Expression, expected any) bool {
	switch v := expected.(type) {
	case int:
		return testNumberLiteral(t, expr, v)
	case float64:
		return testNumberLiteral(t, expr, v)
	case string:
		return testIdentifier(t, expr, v)
	case bool:
		return testBooleanLiteral(t, expr, v)
	case nil:
		return testNullLiteral(t, expr)
	}
	t.Errorf("type of exp not handled, got %T", expr)
	return false
}

func testPrefixExpression(t *testing.T, expr ast.Expression, operator string, right any) bool {
	prefix, ok := expr.(*ast.PrefixExpression)
	if !ok {
		t.Errorf("expr not *ast.PrefixExpression, got %T", expr)
		return false
	}
	if prefix.Operator() != operator {
		t.Errorf("prefix.Operator not %q, got %q", operator, prefix.Operator())
		return false
	}
	if !testLiteralExpression(t, prefix.Right, right) {
		return false
	}
	return true
}
