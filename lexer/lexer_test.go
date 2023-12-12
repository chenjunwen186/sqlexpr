package lexer

import (
	"testing"

	"github.com/chenjunwen186/sqlexpr/token"
)

type ExpectedItem struct {
	expectedType    token.Type
	expectedLiteral string
}

type ExpectedList []ExpectedItem

func (ei ExpectedList) testAll(t *testing.T, name string, l *Lexer) {
	for _, e := range ei {
		tok := l.NextToken()
		if tok.Type != e.expectedType {
			t.Errorf("%s: tok.Type wrong. expected=%q, got=%q", name, e.expectedType, tok.Type)
		}
		if tok.Literal != e.expectedLiteral {
			t.Errorf("%s: tok.Literal wrong. expected=%q, got=%q", name, e.expectedLiteral, tok.Literal)
		}
	}
}

func TestStringLiteral(t *testing.T) {
	type TestCase struct {
		input string
		tok   token.Token
	}

	newToken := func(t token.Type, l string) token.Token {
		return token.Token{
			Type:    t,
			Literal: l,
		}
	}

	inputs := []TestCase{
		{`'hello world'`, newToken(token.STRING, "hello world")},
		{"'hello world", newToken(token.ILLEGAL, `unexpected EOF: 'hello world`)},
		{`'hello -- world'`, newToken(token.ILLEGAL, "not support SQL comment `--` in string literal: 'hello -- world'")},
	}

	for _, input := range inputs {
		l := New(input.input)
		tok := l.NextToken()
		if tok.Type != input.tok.Type {
			t.Errorf("tok.Type wrong. expected=%q, got=%q", input.tok.Type, tok.Type)
		}
		if tok.Literal != input.tok.Literal {
			t.Errorf("tok.Literal wrong. expected=%q, got=%q", input.tok.Literal, tok.Literal)
		}
	}
}

func TestBooleanLiteral(t *testing.T) {
	input := `true false True False TRUE FaLSE`
	expected := ExpectedList{
		{token.TRUE, "true"},
		{token.FALSE, "false"},
		{token.TRUE, "True"},
		{token.FALSE, "False"},
		{token.TRUE, "TRUE"},
		{token.FALSE, "FaLSE"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestBooleanLiteral", l)
}

func TestNullLiteral(t *testing.T) {
	input := `null NULL Null`
	expected := ExpectedList{
		{token.NULL, "null"},
		{token.NULL, "NULL"},
		{token.NULL, "Null"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestNullLiteral", l)
}

func TestNumberPeriodLiteral(t *testing.T) {
	input := `. 123
	. 123.456
	0.456 . 2e2
	0.2e+3 1.23e-2 12.
	0 . .
	0e+3 . 0e-3
	0e
	`
	expected := ExpectedList{
		{token.PERIOD, "."},
		{token.NUMBER, "123"},
		{token.PERIOD, "."},
		{token.NUMBER, "123.456"},
		{token.NUMBER, "0.456"},
		{token.PERIOD, "."},
		{token.NUMBER, "2e2"},
		{token.NUMBER, "0.2e+3"},
		{token.NUMBER, "1.23e-2"},
		{token.NUMBER, "12."},
		{token.NUMBER, "0"},
		{token.PERIOD, "."},
		{token.PERIOD, "."},
		{token.NUMBER, "0e+3"},
		{token.PERIOD, "."},
		{token.NUMBER, "0e-3"},
		{token.ILLEGAL, "invalid number literal: \"0e\""},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestNumberPeriodLiteral", l)
}

func TestIdentifiers(t *testing.T) {
	input := `hello _world world2_ _world_ _world_0`
	expected := ExpectedList{
		{token.IDENT, "hello"},
		{token.IDENT, "_world"},
		{token.IDENT, "world2_"},
		{token.IDENT, "_world_"},
		{token.IDENT, "_world_0"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestIdentifiers", l)
}

func TestBackQuoteIdentifiers(t *testing.T) {
	input := "`Hello:@` `hello world` `hello ` `hello -- world` `hello "
	expected := ExpectedList{
		{token.BACK_QUOTE_IDENT, "`Hello:@`"},
		{token.BACK_QUOTE_IDENT, "`hello world`"},
		{token.BACK_QUOTE_IDENT, "`hello `"},
		{token.ILLEGAL, "not support SQL comment `--` in back quote identifier: `hello -- world`"},
		{token.ILLEGAL, "unexpected EOF: `hello "},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestBackQuoteIdentifiers", l)
}

func TestDoubleQuoteIdentifiers(t *testing.T) {
	input := `"Hello:@" "hello world" "hello " "hello -- world" "hello `
	expected := ExpectedList{
		{token.DOUBLE_QUOTE_IDENT, `"Hello:@"`},
		{token.DOUBLE_QUOTE_IDENT, `"hello world"`},
		{token.DOUBLE_QUOTE_IDENT, `"hello "`},
		{token.ILLEGAL, "not support SQL comment `--` in double quote identifier: \"hello -- world\""},
		{token.ILLEGAL, `unexpected EOF: "hello `},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestBackQuoteIdentifiers", l)
}

func TestOperators(t *testing.T) {
	input := `
	+
	- * / %
	& |
	|| << >> ~
	IS IS NOT
	BETWEEN NOT
	BETWEEN
	NOT LIKE LIKE
	>= <= <=> <> < >
	CASE WHEN x > 1 Then 1 ELSE 0 END
	0b01010 0b01230 01234567 018
	0x1234af 0X123g
`
	expected := ExpectedList{
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.MOD, "%"},
		{token.AMP, "&"},
		{token.PIPE, "|"},
		{token.PIPE2, "||"},
		{token.LT2, "<<"},
		{token.RT2, ">>"},
		{token.TILDE, "~"},
		{token.IS, "IS"},
		{token.IS_NOT, "IS NOT"},
		{token.BETWEEN, "BETWEEN"},
		{token.NOT_BETWEEN, "NOT BETWEEN"},
		{token.NOT_LIKE, "NOT LIKE"},
		{token.LIKE, "LIKE"},
		{token.GT_EQ, ">="},
		{token.LT_EQ, "<="},
		{token.LT_EQ_GT, "<=>"},
		{token.NOT_EQ2, "<>"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.CASE, "CASE"},
		{token.WHEN, "WHEN"},
		{token.IDENT, "x"},
		{token.GT, ">"},
		{token.NUMBER, "1"},
		{token.THEN, "Then"},
		{token.NUMBER, "1"},
		{token.ELSE, "ELSE"},
		{token.NUMBER, "0"},
		{token.END, "END"},
		{token.NUMBER, "0b01010"},
		{token.ILLEGAL, `invalid binary number literal: "0b01230"`},
		{token.NUMBER, "01234567"},
		{token.ILLEGAL, `invalid octal number literal: "018"`},
		{token.NUMBER, "0x1234af"},
		{token.ILLEGAL, `invalid hexadecimal number literal: "0X123g"`},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestOperators", l)
}

func TestPairs(t *testing.T) {
	input := `
	(
	)
	`
	expected := ExpectedList{
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestPairs", l)
}
