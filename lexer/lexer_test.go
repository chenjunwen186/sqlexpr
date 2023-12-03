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
	input := `'hello world'`
	expected := ExpectedList{
		{token.STRING, "hello world"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestStringLiteral", l)
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

func TestNumberLiteral(t *testing.T) {
	input := `123 123.456 .456`
	expected := ExpectedList{
		{token.NUMBER, "123"},
		{token.NUMBER, "123.456"},
		{token.NUMBER, ".456"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestNumberLiteral", l)
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

func TestOperators(t *testing.T) {
	input := `
	+
	- * / %
	& |
	|| << >> ~
	>= <= <=> <> < >`
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
		{token.GT_EQ, ">="},
		{token.LT_EQ, "<="},
		{token.LT_EQ_GT, "<=>"},
		{token.NOT_EQ2, "<>"},
		{token.LT, "<"},
		{token.GT, ">"},
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
