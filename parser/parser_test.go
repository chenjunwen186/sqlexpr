package parser

import (
	"strconv"
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

func parseExpressionWithError(t *testing.T, input string) (ast.Expression, error) {
	l := lexer.New(input)
	p := New(l)
	return p.ParseExpression()
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

func TestGroupedExpression(t *testing.T) {
	input := `(hello)`
	expr := parseExpression(t, input)
	testIdentifier(t, expr, "hello")

	inputEmpty := `()`
	_, err := parseExpressionWithError(t, inputEmpty)
	if err == nil {
		t.Errorf("should parsed error, but not")
	} else if err.Error() != "empty `()` is not supported" {
		t.Errorf("err.Error() should be: empty `()` is not supported")
	}
}

func TestTupleExpression(t *testing.T) {
	input := `(hello, 123, 123.456, .456)`
	expr := parseExpression(t, input)
	tuple, ok := expr.(*ast.TupleExpression)
	if !ok {
		t.Errorf("expr not *ast.TupleExpression, got %T", expr)
	}

	if len(tuple.Expressions) != 4 {
		t.Errorf("len(tuple.Expressions) not 4, got %d", len(tuple.Expressions))
	}

	testIdentifier(t, tuple.Expressions[0], "hello")
	testNumberLiteral(t, tuple.Expressions[1], 123)
	testNumberLiteral(t, tuple.Expressions[2], 123.456)
	testNumberLiteral(t, tuple.Expressions[3], .456)
}

func TestInfixExpression(t *testing.T) {
	type TestCase struct {
		input    string
		left     any
		operator token.Type
		right    any
		str      string
	}

	inputs := []TestCase{
		{"123 + 456", 123, token.PLUS, 456, "(123 + 456)"},
		{"123.456 - 456.789", 123.456, token.MINUS, 456.789, "(123.456 - 456.789)"},
		{"x * y", "x", token.ASTERISK, "y", "(x * y)"},
		{"x / y", "x", token.SLASH, "y", "(x / y)"},
		{"x % y", "x", token.MOD, "y", "(x % y)"},
		{"x Or y", "x", token.OR, "y", "(x OR y)"},
		{"x aNd y", "x", token.AND, "y", "(x AND y)"},
		{"x > y", "x", token.GT, "y", "(x > y)"},
		{"x >= y", "x", token.GT_EQ, "y", "(x >= y)"},
		{"x < y", "x", token.LT, "y", "(x < y)"},
		{"x <= y", "x", token.LT_EQ, "y", "(x <= y)"},
		{"x <=> y", "x", token.LT_EQ_GT, "y", "(x <=> y)"},
		{"x != y", "x", token.NOT_EQ1, "y", "(x != y)"},
		{"x <> y", "x", token.NOT_EQ2, "y", "(x <> y)"},
		{"x iN y", "x", token.IN, "y", "(x IN y)"},
		{"x nOt iN y", "x", token.NOT_IN, "y", "(x NOT IN y)"},
		{"x is y", "x", token.IS, "y", "(x IS y)"},
		{"x is Not y", "x", token.IS_NOT, "y", "(x IS NOT y)"},
		{"x lIkE y", "x", token.LIKE, "y", "(x LIKE y)"},
		{"x nOt lIkE y", "x", token.NOT_LIKE, "y", "(x NOT LIKE y)"},
	}
	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		testInfixExpression(t, expr, input.left, input.operator, input.right)
		if expr.String() != input.str {
			t.Errorf("expr.String() not %q, got %q", input.str, expr.String())
		}
	}
}

func TestBetweenExpression(t *testing.T) {
	type TestCase struct {
		input string
		left  any
		expr  string
	}

	inputs := []TestCase{
		{"123 between 456 and 789", 123, "(456 AND 789)"},
	}
	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		v, ok := expr.(*ast.BetweenExpression)
		if !ok {
			t.Errorf("expr not *ast.BetweenExpression, got %T", expr)
			continue
		}
		testLiteralExpression(t, v.Left, input.left)
		if v.Range.String() != input.expr {
			t.Errorf("v.Range.String() not %q, got %q", input.expr, v.Range.String())
		}
	}
}

func TestNotBetweenExpression(t *testing.T) {
	type TestCase struct {
		input string
		left  any
		expr  string
	}

	inputs := []TestCase{
		{"123 not between 456 and 789", 123, "(456 AND 789)"},
	}
	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		v, ok := expr.(*ast.NotBetweenExpression)
		if !ok {
			t.Errorf("expr not *ast.NotBetweenExpression, got %T", expr)
			continue
		}
		testLiteralExpression(t, v.Left, input.left)
		if v.Range.String() != input.expr {
			t.Errorf("v.Range.String() not %q, got %q", input.expr, v.Range.String())
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

func testInfixExpression(t *testing.T, expr ast.Expression, left any, operator token.Type, right any) bool {
	switch infix := expr.(type) {
	case *ast.InfixExpression:
		{
			if infix.Operator() != operator {
				t.Errorf("infix.Operator not %q, got %q", operator, infix.Operator())
				return false
			}
			if !testLiteralExpression(t, infix.Left, left) {
				return false
			}
			if !testLiteralExpression(t, infix.Right, right) {
				return false
			}
			return true
		}
	}

	t.Errorf("expr not *ast.InfixExpression, got %T", expr)
	return false
}

func TestCallExpression(t *testing.T) {
	type TestCase struct {
		input  string
		fnName string
		args   []string
	}

	inputs := []TestCase{
		{"hello()", "hello", []string{}},
		{"hello(123)", "hello", []string{"123"}},
		{`hello(123, .456)`, "hello", []string{"123", ".456"}},
		{`hello(123, x + y, x * y)`, "hello", []string{"123", "(x + y)", "(x * y)"}},
	}
	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		testCallExpression(t, expr, input.fnName, input.args)
	}
}

func testCallExpression(t *testing.T, expr ast.Expression, fnName string, args []string) bool {
	call, ok := expr.(*ast.CallExpression)
	if !ok {
		t.Errorf("expr not *ast.CallExpression, got %T", expr)
		return false
	}
	if call.FnName.Value != fnName {
		t.Errorf("call.Function.Value not %q, got %q", fnName, call.FnName.Value)
		return false
	}

	if len(call.Arguments) != len(args) {
		t.Errorf("len(call.Arguments) not %d, got %d", len(args), len(call.Arguments))
		return false
	}
	for i, arg := range args {
		expected := call.Arguments[i].String()
		if arg != expected {
			t.Errorf("call.Arguments[%d].String() not %q, got %q", i, arg, expected)
			return false
		}
	}
	return true
}

func TestCaseWhenExpression(t *testing.T) {
	type WhenCase struct {
		condition string
		result    string
	}

	type TestCase struct {
		input      string
		whens      []WhenCase
		elseResult string
	}

	inputs := []TestCase{
		{
			"CASE WHEN x > 0 THEN 1 WHEN x < 0 THEN -1 ELSE 0 END",
			[]WhenCase{
				{"(x > 0)", "1"},
				{"(x < 0)", "(-1)"},
			},
			"0",
		},
	}

	for _, input := range inputs {
		expr := parseExpression(t, input.input)
		v, ok := expr.(*ast.CaseWhenExpression)
		if !ok {
			t.Errorf("expr not *ast.CaseWhenExpression, got %T", expr)
			continue
		}
		if len(v.Whens) != len(input.whens) {
			t.Errorf("len(v.Whens) not %d, got %d", len(input.whens), len(v.Whens))
			continue
		}

		for i, when := range input.whens {
			cond := v.Whens[i].Cond.String()
			if cond != when.condition {
				t.Errorf("v.Whens[%d].Cond.String() not %q, got %q", i, when.condition, cond)
			}
			then := v.Whens[i].Then.String()
			if then != when.result {
				t.Errorf("v.Whens[%d].Then.String() not %q, got %q", i, when.result, then)
			}
		}

		if input.elseResult != "" {
			elseStr := v.Else.String()
			if elseStr != input.elseResult {
				t.Errorf("v.Else.String() not %q, got %q", input.elseResult, elseStr)
			}
		}
	}
}
