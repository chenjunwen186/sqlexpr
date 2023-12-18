package lexer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chenjunwen186/sqlexpr/token"
)

type ExpectedLiteral struct {
	expectedType    token.Type
	expectedLiteral string
}

type ExpectedLiterals []ExpectedLiteral

func (ei ExpectedLiterals) testAll(t *testing.T, name string, l *Lexer) {
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

type TokenCase struct {
	input           string
	expectedType    token.Type
	expectedLiteral string
}

type TokenCases []TokenCase

func (tc TokenCases) testAll(t *testing.T, name string) {
	for _, v := range tc {
		l := New(v.input)
		tok := l.NextToken()
		if tok.Type != v.expectedType {
			t.Errorf("%s: token.Type wrong. expected=%q, got=%q", name, v.expectedType, tok.Type)
		}
		if tok.Literal != v.expectedLiteral {
			t.Errorf("%s: token.Literal wrong. expected=%q, got=%q", name, v.expectedLiteral, tok.Literal)
		}
	}
}

type IllegalCase struct {
	input string
	err   string
}

type IllegalCases []IllegalCase

func (il IllegalCases) testAll(t *testing.T, name string) {
	for _, v := range il {
		l := New(v.input)
		var tok token.Token
		for {
			tok = l.NextToken()
			if tok.Type == token.EOF {
				break
			}

			if tok.Type == token.ILLEGAL {
				break
			}
		}

		if tok.Type != token.ILLEGAL {
			t.Errorf("%s: tok.Type wrong. expected=%q, got=%q", name, token.ILLEGAL, tok.Type)
		} else {
			if tok.Literal != v.err {
				t.Errorf("%s: tok.Literal wrong. expected=%q, got=%q", name, v.err, tok.Literal)
			}
		}
	}
}

func TestStringLiteral(t *testing.T) {

	tokenCases := TokenCases{
		{`''`, token.STRING, `''`},
		{`'hello world'`, token.STRING, "'hello world'"},
		{"'hello world", token.ILLEGAL, `unexpected EOF: 'hello world`},
		{`'hello -- world'`, token.STRING, "'hello -- world'"},
		{`'hello # world'`, token.STRING, "'hello # world'"},
		{`'hello \' world'`, token.STRING, `'hello \' world'`},
		{`'hello \'\'\' world'`, token.STRING, `'hello \'\'\' world'`},
		{`'hello \'''\'''\' \' world'''`, token.STRING, `'hello \'''\'''\' \' world'''`},
		{`'hello \'`, token.ILLEGAL, `unexpected EOF: 'hello \'`},
		{`'hello ''`, token.ILLEGAL, `unexpected EOF: 'hello ''`},
		{`'hello \'\'\'`, token.ILLEGAL, `unexpected EOF: 'hello \'\'\'`},
		{`'hello ''''`, token.ILLEGAL, `unexpected EOF: 'hello ''''`},
		{`'hello '' world'`, token.STRING, "'hello '' world'"},
		{`'hello '''' world'`, token.STRING, "'hello '''' world'"},
		{`' 你好世界! '`, token.STRING, "' 你好世界! '"},
		{`' こんにちは世界! '`, token.STRING, "' こんにちは世界! '"},
		{`' 안녕하세요 세계! '`, token.STRING, "' 안녕하세요 세계! '"},
		{`' สวัสดีชาวโลก! '`, token.STRING, "' สวัสดีชาวโลก! '"},
		{`' Γειά σου Κόσμε! '`, token.STRING, "' Γειά σου Κόσμε! '"},
	}

	tokenCases.testAll(t, "TestStringLiteral")

	illegalCases := IllegalCases{
		{
			`'hello \''; deleTe from test where test.a = 1; -- '`,
			"not support token `;`",
		},
		{
			`'\''; select * from test --'`,
			"not support token `;`",
		},
	}

	illegalCases.testAll(t, "TestStringLiteral")
}

func TestBooleanLiteral(t *testing.T) {
	input := `true false True False TRUE FaLSE`
	expected := ExpectedLiterals{
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
	expected := ExpectedLiterals{
		{token.NULL, "null"},
		{token.NULL, "NULL"},
		{token.NULL, "Null"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestNullLiteral", l)
}

func TestNumberLiteral(t *testing.T) {
	input := `. 123
	. 123.456
	0.456 . 2e2
	0.2e+3 1.23e-2 12.
	0 . .
	0e+3 . 0e-3
	0e 0.e+
	0e+3+3 12.e-3+3
	0X123g 0b01010 0b01230 01234567 018 0xae12c34af
	`
	expected := ExpectedLiterals{
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
		{token.ILLEGAL, "invalid number literal: \"0.e+\""},
		{token.NUMBER, "0e+3"},
		{token.PLUS, "+"},
		{token.NUMBER, "3"},
		{token.NUMBER, "12.e-3"},
		{token.PLUS, "+"},
		{token.NUMBER, "3"},
		{token.ILLEGAL, `invalid hexadecimal number literal: "0X123g"`},
		{token.NUMBER, "0b01010"},
		{token.ILLEGAL, `invalid binary number literal: "0b01230"`},
		{token.NUMBER, "01234567"},
		{token.ILLEGAL, `invalid octal number literal: "018"`},
		{token.NUMBER, "0xae12c34af"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestNumberPeriodLiteral", l)
}

func TestIdentifiers(t *testing.T) {
	input := `hello _world world2_ _world_ _world_0
        HELLO_WORLD HelloWorld helloWorld
    `
	expected := ExpectedLiterals{
		{token.IDENT, "hello"},
		{token.IDENT, "_world"},
		{token.IDENT, "world2_"},
		{token.IDENT, "_world_"},
		{token.IDENT, "_world_0"},
		{token.IDENT, "HELLO_WORLD"},
		{token.IDENT, "HelloWorld"},
		{token.IDENT, "helloWorld"},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestIdentifiers", l)
}

func TestBackQuoteIdentifiers(t *testing.T) {
	literalsInput := "`Hello:@` `hello world` `hello ` `hello -- world` `hello "
	literalCases := ExpectedLiterals{
		{token.BACK_QUOTE_IDENT, "`Hello:@`"},
		{token.BACK_QUOTE_IDENT, "`hello world`"},
		{token.BACK_QUOTE_IDENT, "`hello `"},
		{token.BACK_QUOTE_IDENT, "`hello -- world`"},
		{token.ILLEGAL, "unexpected EOF: `hello "},
		{token.EOF, ""},
	}

	l := New(literalsInput)

	literalCases.testAll(t, "TestDoubleQuoteIdentifiers", l)

}

func TestDoubleQuoteIdentifiers(t *testing.T) {
	input := `"Hello:@" "hello world" "hello " "hello -- world" "hello `
	expected := ExpectedLiterals{
		{token.DOUBLE_QUOTE_IDENT, `"Hello:@"`},
		{token.DOUBLE_QUOTE_IDENT, `"hello world"`},
		{token.DOUBLE_QUOTE_IDENT, `"hello "`},
		{token.DOUBLE_QUOTE_IDENT, "\"hello -- world\""},
		{token.ILLEGAL, `unexpected EOF: "hello `},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestDoubleQuoteIdentifiers", l)
}

func TestOperators(t *testing.T) {
	input := `
	+
	- * / %
	& | ^ -> ->>
	|| << >> ~
	IS IS NOT
	BETWEEN NOT
	BETWEEN
	NOT LIKE LIKE -- hello : world ~
	/*
    hello
    world
    */
	# CASE
	! != !< !>
	>= <= <=> <> < > -> ->> --
	CASE WHEN x > 1 Then 1 ELSE 0 END # hello@world
	? : ,: 1::int
    /* hello
`
	expected := ExpectedLiterals{
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.MOD, "%"},
		{token.AMP, "&"},
		{token.PIPE, "|"},
		{token.XOR, "^"},
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
		{token.ILLEGAL, `not support SQL comment: "-- hello : world ~"`},
		{token.ILLEGAL, "not support SQL comment: \"/*\n    hello\n    world\n    */\""},
		{token.ILLEGAL, `not support SQL comment: "# CASE"`},
		{token.BANG, "!"},
		{token.BANG_EQ, "!="},
		{token.BANG_LT, "!<"},
		{token.BANG_GT, "!>"},
		{token.GT_EQ, ">="},
		{token.LT_EQ, "<="},
		{token.LT_EQ_GT, "<=>"},
		{token.NOT_EQ, "<>"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.PRT, "->"},
		{token.PRT2, "->>"},
		{token.ILLEGAL, `not support SQL comment: "--"`},
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
		{token.ILLEGAL, `not support SQL comment: "# hello@world"`},
		{token.QUESTION, "?"},
		{token.COLON, ":"},
		{token.COMMA, ","},
		{token.COLON, ":"},
		{token.NUMBER, "1"},
		{token.COLON2, "::"},
		{token.IDENT, "int"},
		{token.ILLEGAL, "unexpected EOF: \"/* hello\n\""},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestOperators", l)
}

func TestPairs(t *testing.T) {
	input := `
	(
	)

	[ ) ] (
	`
	expected := ExpectedLiterals{
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACKET, "["},
		{token.RPAREN, ")"},
		{token.RBRACKET, "]"},
		{token.LPAREN, "("},
		{token.EOF, ""},
	}

	l := New(input)

	expected.testAll(t, "TestPairs", l)
}

func TestExpressions(t *testing.T) {
	type TestCase struct {
		input   string
		literal string
		errMsg  string
	}

	inputs := []TestCase{
		{`1 + 2`, `1 + 2`, ""},
		{`1 + 2 * 3`, `1 + 2 * 3`, ""},
		{`(1 + 2) * 3`, `( 1 + 2 ) * 3`, ""},
		{`(1 ,2)`, `( 1 , 2 )`, ""},
		{
			`arrayFilter(x -> x > 1, [1, 2, 3])[1]`,
			`arrayFilter ( x -> x > 1 , [ 1 , 2 , 3 ] ) [ 1 ]`,
			"",
		},
		{"1 ? 2 : 3", "1 ? 2 : 3", ""},
		{`sumIf(1, 1)`, `sumIf ( 1 , 1 )`, ""},
		{
			`COUNT(*) as c, "world", True as t`,
			`COUNT ( * ) as c , "world" , True as t`,
			"",
		},
		{
			`CASE WHEN x > 1 Then 1 ELSE 0 END`,
			`CASE WHEN x > 1 Then 1 ELSE 0 END`,
			"",
		},
		{
			`[1, 02, 0.3, 4., 0b01010, 0XAbC, 1.e+3 , 123e-3, -1, 0]`,
			`[ 1 , 02 , 0.3 , 4. , 0b01010 , 0XAbC , 1.e+3 , 123e-3 , - 1 , 0 ]`,
			"",
		},
		{`1::int`, `1 :: int`, ""},
		{`1::int::int`, `1 :: int :: int`, ""},
		{
			`CAST(order_amount AS DECIMAL(10, 2))`,
			`CAST ( order_amount AS DECIMAL ( 10 , 2 ) )`,
			"",
		},
		{
			`DATE_SUB('2023-01-15', INTERVAL 3 MONTH)`,
			`DATE_SUB ( '2023-01-15' , INTERVAL 3 MONTH )`,
			"",
		},
		{
			`EXTRACT(YEAR FROM '2023-05-15 14:30:00')`,
			`EXTRACT ( YEAR FROM '2023-05-15 14:30:00' )`,
			"",
		},
		{
			`'hello world' select * from hello; -- '`,
			"",
			"not support keyword: \"select\"",
		},
	}

	for _, input := range inputs {
		var (
			tokens []token.Token
			errMsg string
		)
		l := New(input.input)
		for {
			tok := l.NextToken()

			if tok.Type == token.EOF {
				break
			}

			if tok.Type == token.ILLEGAL {
				errMsg = tok.Literal
				break
			}

			tokens = append(tokens, tok)
		}

		if errMsg != "" {
			if input.errMsg != errMsg {
				t.Errorf("errMsg wrong. expected=%q, got=%q", input.errMsg, errMsg)
			}
		} else {
			literal := tokensToString(tokens)
			if literal != input.literal {
				t.Errorf("literal wrong. expected=%q, got=%q", input.literal, literal)
			}
		}
	}
}

func tokensToString(tokens []token.Token) string {
	var literals []string
	for _, tok := range tokens {
		literals = append(literals, tok.Literal)
	}

	return strings.Join(literals, " ")
}

func testBenchmark(input string) error {
	l := New(input)
	for {
		tok := l.NextToken()

		if tok.Type == token.EOF {
			break
		}

		if tok.Type == token.ILLEGAL {
			return fmt.Errorf("illegal token: %s\n", tok.Literal)
		}
	}

	return nil
}

func BenchmarkLexerParse(b *testing.B) {
	input := `
	() [ ) ] (
	arrayFilter(x -> x > 1, [1, 2, 3])[1]
	1 ? 2 : 3
	sumIf(1, 1)
	COUNT(*) as c, "world", True as t
	CASE WHEN x > 1 Then 1 When x = 0 THEN 2 WHEN x < 0 THEN ELSE 0 END
	[1, 02, 0.3, 4., 0b01010, 0XAbC, 1.e+3 , 123e-3, -1, 0]
	1::int
	1::int::int
	CAST(order_amount AS DECIMAL(10, 2))
	DATE_SUB('2023-01-15', INTERVAL 3 MONTH)
	EXTRACT(YEAR FROM '2023-05-15 14:30:00')
	+ - * / %
	& | ^ -> ->>
	|| << >> ~
	IS IS NOT
	BETWEEN NOT
	BETWEEN
	NOT LIKE LIKE
	! != !< !>
	>= <= <=> <> < > -> ->>
	CASE WHEN x > 1 Then 1 ELSE 0 END
	? : ,: 1::int
	hello _world world2_ _world_ _world_0
    HELLO_WORLD HelloWorld helloWorld
	0 <0 >0 . 123
	. 123.456
	0.456 . 2e2
	0.2e+3 1.23e-2 12.
	0 . .
	0e+3 . 0e-3
	0e+3+3 12.e-3+3
	0b01010 01234567 0xae12cdef
	"Hello:@" "hello world" "hello " "hello -- world"
	null NULL Null true false True False TRUE FaLSE
	'' 'hello world' 'hello ' 'hello -- world' 'hello '
	'hello # world' 'hello \' world' 'hello \'\'\' world'
	'hello \'''\'''\' \' world''' 'hello \'' 'hello '''
	'hello \'\'\' ' 'hello '' world' 'hello '''' world'
	' 你好世界! ' ' こんにちは世界! ' ' 안녕하세요 세계! '
	' สวัสดีชาวโลก! ' ' Γειά σου Κόσμε! '
`

	input += "`Hello:@` `hello world` `hello ` `hello -- world`"

	for i := 0; i < b.N; i++ {
		if err := testBenchmark(input); err != nil {
			b.Fatal(err)
		}
	}
}
