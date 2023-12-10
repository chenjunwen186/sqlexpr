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

	IDENT_QUOTED1 = "IDENT_QUOTED" // `ident` for MySQL, Sqlite, Clickhouse, ORACLE, SparkSQL
	IDENT_QUOTED2 = "IDENT_QUOTED" // "ident" for PgSQL
	IDENT_QUOTED3 = "IDENT_QUOTED" // [ident] for MSSQL

	STRING = "STRING"
	NUMBER = "NUMBER"

	NOT = "NOT"

	NOT_IN      = "NOT IN"
	NOT_LIKE    = "NOT LIKE"
	NOT_BETWEEN = "NOT BETWEEN"
	IS_NOT      = "IS NOT"

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
	PERIOD   = "."

	COMMA = ","

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
	PRT      = "->"
	PRT2     = "->>"

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

	DISTINCT = "DISTINCT"
	AS       = "AS"

	INTERVAL = "INTERVAL"
	DAY      = "DAY"
	HOUR     = "HOUR"
	MONTH    = "MONTH"
	MINUTE   = "MINUTE"
	WEEK     = "WEEK"
	YEAR     = "YEAR"
	QUARTER  = "QUARTER"
	SECOND   = "SECOND"
)

type Token struct {
	Type    Type
	Literal string
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %s)", t.Type, t.Literal)
}

func (t Token) IsError() error {
	if t.Type == ILLEGAL {
		return fmt.Errorf(t.Literal)
	}

	return nil
}

func (t Token) IsEOF() bool {
	return t.Type == EOF
}

func NewIllegalToken(errMsg string) Token {
	return Token{
		Type:    ILLEGAL,
		Literal: errMsg,
	}
}

var keywords = map[string]Type{
	"CASE": CASE,
	"END":  END,
	"WHEN": WHEN,
	"THEN": THEN,
	"ELSE": ELSE,

	"TRUE":  TRUE,
	"FALSE": FALSE,
	"NULL":  NULL,

	"NOT": NOT,

	"IN":      IN,
	"BETWEEN": BETWEEN,
	"IS":      IS,
	"LIKE":    LIKE,

	"AND": AND,
	"OR":  OR,

	"DISTINCT": DISTINCT,
	"AS":       AS,

	// time
	"INTERVAL": INTERVAL,
	"DAY":      DAY,
	"HOUR":     HOUR,
	"MONTH":    MONTH,
	"MINUTE":   MINUTE,
	"WEEK":     WEEK,
	"YEAR":     YEAR,
	"QUARTER":  QUARTER,
	"SECOND":   SECOND,
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

func (t Type) IsTimeUnit() bool {
	switch t {
	case DAY, HOUR, MONTH, MINUTE, WEEK, YEAR, QUARTER, SECOND:
		return true
	default:
		return false
	}
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
