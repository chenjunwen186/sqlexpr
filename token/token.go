package token

import (
	"fmt"
	"strings"
)

type Type string

const (
	ILLEGAL = "ILLEGAL"

	EOF = "EOF"

	IDENT = "IDENT"

	IDENT_QUOTED = "IDENT_QUOTED"

	// Float, Integer...
	NUMBER = "NUMBER"
	// INT    = "INT"
	// FLOAT  = "FLOAT"
	STRING = "STRING"

	NOT = "NOT"

	PIPE = "|"
	AMP  = "&"

	PLUS     = "+"
	MINUS    = "-"
	SLASH    = "/"
	ASTERISK = "*"
	MOD      = "%"
	PIPE2    = "||"
	LT2      = "<<"
	RT2      = ">>"
	TILDE    = "~"

	SEMICOLON = ";"
	COMMA     = ","

	LPAREN = "("
	RPAREN = ")"

	EQ       = "="
	NOT_EQ1  = "!="
	NOT_EQ2  = "<>"
	LT       = "<"
	LT_EQ    = "<="
	GT       = ">"
	GT_EQ    = ">="
	LT_EQ_GT = "<=>"

	AND = "AND"
	OR  = "OR"

	CASE = "CASE"
	END  = "END"
	WHEN = "WHEN"
	THEN = "THEN"
	ELSE = "ELSE"

	TRUE  = "TRUE"
	FALSE = "FALSE"
	NULL  = "NULL"

	IN      = "IN"
	LIKE    = "LIKE"
	IS      = "IS"
	BETWEEN = "BETWEEN"
)

type Token struct {
	Type    Type
	Literal string
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %s)", t.Type, t.Literal)
}

var keywords = map[string]Type{
	"CASE":    CASE,
	"END":     END,
	"WHEN":    WHEN,
	"TRUE":    TRUE,
	"FALSE":   FALSE,
	"NULL":    NULL,
	"IN":      IN,
	"IS":      IS,
	"AND":     AND,
	"OR":      OR,
	"THEN":    THEN,
	"ELSE":    ELSE,
	"LIKE":    LIKE,
	"BETWEEN": BETWEEN,
}

var notSupportKeywords = map[string]Type{}

func registerNotSupportKeyword(keywords ...string) {
	for _, keyword := range keywords {
		notSupportKeywords[keyword] = ILLEGAL
	}
}

func init() {
	registerNotSupportKeyword(
		"SELECT",
		"FROM",
		"WHERE",
		"GROUP",
		"BY",
		"HAVING",
		"ORDER",
		"LIMIT",
		"OFFSET",
		"UNION",
		"ALL",
		"DISTINCT",
		"AS",
		"ON",
		"USING",
		"INNER",
		"LEFT",
		"RIGHT",
		"FULL",
		"OUTER",
		"JOIN",
		"CROSS",
		"NATURAL",
		"ASC",
		"DESC",
		"UNION",
	)
}

func LookupIdent(ident string) Token {
	v := strings.ToUpper(ident)
	if typ, ok := notSupportKeywords[v]; ok {
		return Token{
			Type:    typ,
			Literal: fmt.Sprintf("not support keyword: %s", ident),
		}
	}

	if typ, ok := keywords[v]; ok {
		return Token{
			Type:    typ,
			Literal: ident,
		}
	}

	return Token{
		Type:    IDENT,
		Literal: ident,
	}
}
