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

	BACK_QUOTE_IDENT   = "BACK_QUOTE_IDENT"   // `ident` for MySQL, Sqlite, Clickhouse, ORACLE, SparkSQL
	DOUBLE_QUOTE_IDENT = "DOUBLE_QUOTE_IDENT" // "ident" for PgSQL, Clickhouse

	// Currently not support
	// Because it conflicts with Clickhouse's Array Literal
	// BRACKET_IDENT = "BRACKET_IDENT" // [ident] for MSSQL

	STRING = "STRING"
	NUMBER = "NUMBER"

	NOT_IN      = "NOT IN"
	NOT_LIKE    = "NOT LIKE"
	NOT_BETWEEN = "NOT BETWEEN"
	IS_NOT      = "IS NOT"

	PIPE = "|"
	AMP  = "&"
	XOR  = "^"

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

	QUESTION = "?"
	COLON    = ":"

	COLON2 = "::" // type case: select 1::int

	COMMA = ","

	LPAREN   = "("
	RPAREN   = ")"
	LBRACKET = "["
	RBRACKET = "]"

	NOT = "NOT"

	BANG    = "!"
	BANG_GT = "!>"
	BANG_LT = "!<"

	EQ       = "="
	BANG_EQ  = "!="
	NOT_EQ   = "<>"
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

	FROM = "FROM"

	ASC    = "ASC"
	DESC   = "DESC"
	ROWNUM = "ROWNUM" // for Oracle

	TRUE  = "TRUE"
	FALSE = "FALSE"
	NULL  = "NULL"

	IN      = "IN"
	LIKE    = "LIKE"
	IS      = "IS"
	BETWEEN = "BETWEEN"

	ANY    = "ANY"
	EXISTS = "EXISTS"

	DISTINCT = "DISTINCT"
	AS       = "AS"
	TOP      = "TOP" // for Oracle

	INTERVAL = "INTERVAL"
	SECOND   = "SECOND"
	MINUTE   = "MINUTE"
	HOUR     = "HOUR"
	DAY      = "DAY"
	WEEK     = "WEEK"
	MONTH    = "MONTH"
	QUARTER  = "QUARTER"
	YEAR     = "YEAR"
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
	"FROM": FROM,

	"ASC":    ASC,
	"DESC":   DESC,
	"ROWNUM": ROWNUM, // For Oracle

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
	"TOP":      TOP,
	"ANY":      ANY,
	"EXISTS":   EXISTS,

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
		"CREATE",
		"DROP",
		"TABLE",
		"INDEX",
		"VALUES",
		"UPDATE",
		"VIEW",
		"PRIMARY",
		"KEY",
		"EXEC",
		"ADD",
		"CONSTRAINT",
		"ALTER",
		"COLUMN",
		"BACKUP",
		"DATABASE",
		"CHECK",
		"REPLACE",
		"DELETE",
		"INSERT",
		"INTO",
		"PROCEDURE",
		"UNIQUE",
		"DEFAULT",
		"FOREIGN",
		"TRUNCATE",
		"WITH",
		"SET",
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
