package lexer

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/chenjunwen186/sqlexpr/token"
)

var EOF rune = 0

type Lexer struct {
	input        []rune
	position     int
	nextPosition int

	preChar rune
	char    rune

	nextToken token.Token
}

func New(input string) *Lexer {
	l := &Lexer{input: []rune(input)}
	l.readChar()

	l.nextToken = l.move()
	return l
}

func (l *Lexer) Len() int {
	return len(l.input)
}

func (l *Lexer) readChar() {
	l.preChar = l.char
	if l.nextPosition >= len(l.input) {
		l.char = EOF
	} else {
		l.char = l.input[l.nextPosition]
	}
	l.position = l.nextPosition
	l.nextPosition += 1
}

func (l *Lexer) peekChar() rune {
	if l.nextPosition >= len(l.input) {
		return 0
	}
	return l.input[l.nextPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.isWhitespace() {
		l.readChar()
	}
}

func (l *Lexer) isWhitespace() bool {
	return l.char == ' ' || l.char == '\t' || l.char == '\n' || l.char == '\r'
}

// Start with [\d]
// Support 0 100 1.0 2e2 1.23e3 0.23e-3 0.1e+3 12. 1.e3 0e+3, 0b01, 0x1af 0765
// Not support 1e 1e+ 1e- 1e1.2 1e1e2 .12
// 1e+3+3 => ((1e+3)+3)
func (l *Lexer) readNumber() token.Token {
	var b bytes.Buffer

	if l.char == '0' {
		peekChar := l.peekChar()
		if peekChar == 'b' || peekChar == 'B' {
			return l.readBinaryNumber()
		} else if peekChar == 'x' || peekChar == 'X' {
			return l.readHexadecimalNumber()
		} else if unicode.IsDigit(peekChar) {
			return l.readOctalNumber()
		}
	}

	var (
		hasPeriod   bool
		hasExponent bool
		hasSign     bool

		isInvalid bool
	)

	isExponent := func(char rune) bool {
		return char == 'e' || char == 'E'
	}

	for unicode.IsLetter(l.char) || unicode.IsDigit(l.char) || l.char == '.' || l.char == '+' || l.char == '-' {
		if l.char == '+' || l.char == '-' {
			if hasSign {
				// 12.e+3+3 => ((12.e+3)+3)
				// 12.e-3-3 => ((12.e-3)-3)
				break
			} else {
				hasSign = true
			}

			// 0e+ is invalid
			// 0e- is invalid
			// 0e.1 is invalid
			if !hasExponent || !unicode.IsDigit(l.peekChar()) {
				isInvalid = true
			}
		} else if l.char == '.' {
			if hasPeriod {
				isInvalid = true
			} else {
				hasPeriod = true
			}
		} else if isExponent(l.char) {
			if hasExponent {
				isInvalid = true
			} else {
				hasExponent = true
			}

			if l.peekChar() != '+' && l.peekChar() != '-' && !unicode.IsDigit(l.peekChar()) {
				isInvalid = true
			}
		} else if hasExponent {
			if !unicode.IsDigit(l.char) {
				isInvalid = true
			}
		} else if unicode.IsLetter(l.char) {
			isInvalid = true
		}

		b.WriteRune(l.char)
		l.readChar()
	}

	if isInvalid {
		return token.NewIllegalToken(fmt.Sprintf("invalid number literal: %q", b.String()))
	}

	return token.Token{Type: token.NUMBER, Literal: b.String()}
}

// Start with 0[bB]
func (l *Lexer) readBinaryNumber() token.Token {
	var b bytes.Buffer

	// Write `0`
	b.WriteRune(l.char)
	l.readChar()

	// Write `b` or `B`
	b.WriteRune(l.char)
	l.readChar()

	var isIllegal bool
	for unicode.IsDigit(l.char) {
		if l.char == '0' || l.char == '1' {
			b.WriteRune(l.char)
		} else {
			isIllegal = true
			b.WriteRune(l.char)
		}
		l.readChar()
	}

	if isIllegal {
		return token.NewIllegalToken(fmt.Sprintf("invalid binary number literal: %q", b.String()))
	}

	return token.Token{Type: token.NUMBER, Literal: b.String()}
}

// Start with 0[\d]
func (l *Lexer) readOctalNumber() token.Token {
	var b bytes.Buffer

	// Write `0`
	b.WriteRune(l.char)
	l.readChar()

	// Write `0` ~ `7`
	b.WriteRune(l.char)
	l.readChar()

	var isIllegal bool
	for unicode.IsDigit(l.char) {
		if l.char >= '0' && l.char <= '7' {
			b.WriteRune(l.char)
		} else {
			isIllegal = true
			b.WriteRune(l.char)
		}
		l.readChar()
	}

	if isIllegal {
		return token.NewIllegalToken(fmt.Sprintf("invalid octal number literal: %q", b.String()))
	}

	return token.Token{Type: token.NUMBER, Literal: b.String()}
}

// Start with 0[xX]
func (l *Lexer) readHexadecimalNumber() token.Token {
	var b bytes.Buffer

	// Write `0`
	b.WriteRune(l.char)
	l.readChar()
	// Write `x` or `X`
	b.WriteRune(l.char)
	l.readChar()

	var isIllegal bool
	for unicode.IsDigit(l.char) || unicode.IsLetter(l.char) {
		if (l.char >= '0' && l.char <= '9') || (l.char >= 'a' && l.char <= 'f') || (l.char >= 'A' && l.char <= 'F') {
			b.WriteRune(l.char)
		} else {
			isIllegal = true
			b.WriteRune(l.char)
		}
		l.readChar()
	}

	if isIllegal {
		return token.NewIllegalToken(fmt.Sprintf("invalid hexadecimal number literal: %q", b.String()))
	}

	return token.Token{Type: token.NUMBER, Literal: b.String()}
}

func (l *Lexer) readString() token.Token {
	var b bytes.Buffer

	// Write `'`
	b.WriteRune(l.char)

	var (
		isPreValidEscape bool
		isPreValidQuote  bool
	)
	for {
		l.readChar()

		if l.char == EOF {
			return token.NewIllegalToken(fmt.Sprintf("unexpected EOF: %s", b.String()))
		}

		if l.char == '\'' && !isPreValidEscape && !isPreValidQuote {
			if l.peekChar() != '\'' {
				// Write end `'`
				b.WriteRune(l.char)
				break
			} else {
				isPreValidQuote = true
			}
		} else {
			isPreValidQuote = false
		}

		if l.char == '\\' && !isPreValidEscape {
			isPreValidEscape = true
		} else {
			isPreValidEscape = false
		}

		b.WriteRune(l.char)
	}

	return token.Token{Type: token.STRING, Literal: b.String()}
}

func (l *Lexer) readBackQuoteIdentifier() token.Token {
	var b bytes.Buffer

	// Write '`'
	b.WriteRune(l.char)

	var (
		isPreValidEscape    bool
		isPreValidBackQuote bool
	)
	for {
		l.readChar()

		if l.char == EOF {
			return token.NewIllegalToken(fmt.Sprintf("unexpected EOF: %s", b.String()))
		}

		if l.char == '`' && !isPreValidEscape && !isPreValidBackQuote {
			if l.peekChar() != '`' {
				// Write end '`'
				b.WriteRune(l.char)
				break
			} else {
				isPreValidBackQuote = true
			}
		} else {
			isPreValidBackQuote = false
		}

		if l.char == '\\' && !isPreValidEscape {
			isPreValidEscape = true
		} else {
			isPreValidEscape = false
		}

		b.WriteRune(l.char)
	}

	return token.Token{Type: token.BACK_QUOTE_IDENT, Literal: "`" + b.String() + "`"}
}

func (l *Lexer) readDoubleQuoteIdentifier() token.Token {
	var b bytes.Buffer

	// Write `"`
	b.WriteRune(l.char)

	var (
		isPreValidEscape      bool
		isPreValidDoubleQuote bool
	)
	for {
		l.readChar()

		if l.char == EOF {
			return token.NewIllegalToken(fmt.Sprintf(`unexpected EOF: "%s`, b.String()))
		}

		if l.char == '"' && !isPreValidEscape && !isPreValidDoubleQuote {
			if l.peekChar() != '"' {
				// Write end `"`
				b.WriteRune(l.char)
				break
			} else {
				isPreValidDoubleQuote = true
			}
		} else {
			isPreValidDoubleQuote = false
		}

		if l.char == '\\' && !isPreValidEscape {
			isPreValidEscape = true
		} else {
			isPreValidEscape = false
		}

		b.WriteRune(l.char)
	}

	return token.Token{Type: token.DOUBLE_QUOTE_IDENT, Literal: `"` + b.String() + `"`}
}

func (l *Lexer) readIdentifier() string {
	var b bytes.Buffer

	for isIdentifier(l.char) || unicode.IsDigit(l.char) {
		b.WriteRune(l.char)
		l.readChar()
	}

	return b.String()
}

func (l *Lexer) readSingleLineComment() token.Token {
	var b bytes.Buffer

	// Write `-` or `#`
	b.WriteRune(l.char)

	for {
		l.readChar()

		if l.char == EOF {
			break
		}

		if l.char == '\r' && l.peekChar() == '\n' {
			l.readChar()
			break
		}

		if l.char == '\n' {
			break
		}

		b.WriteRune(l.char)
	}

	// Do not support `--` or `#` token to reduce SQL injection risk.
	return token.NewIllegalToken(fmt.Sprintf(`not support SQL comment: "%s"`, b.String()))
}

func (l *Lexer) readMultilineComment() token.Token {
	var b bytes.Buffer

	// Read `/*`
	b.WriteRune(l.char)
	l.readChar()
	b.WriteRune(l.char)

	for {
		l.readChar()

		if l.char == EOF {
			return token.NewIllegalToken(fmt.Sprintf(`unexpected EOF: "%s"`, b.String()))
		}

		if l.char == '*' && l.peekChar() == '/' {
			b.WriteRune(l.char)
			l.readChar()
			b.WriteRune(l.char)
			break
		}

		b.WriteRune(l.char)
	}

	// Do not support `/* */` token to reduce SQL injection risk.
	return token.NewIllegalToken(fmt.Sprintf(`not support SQL comment: "%s"`, b.String()))
}

func isIdentifier(char rune) bool {
	if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' {
		return true
	}

	return false
}

func (l *Lexer) isIdentifierStart() bool {
	if unicode.IsLetter(l.char) || l.char == '_' {
		return true
	}

	return false
}

func newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) NextToken() token.Token {
	tok := l.nextToken
	l.nextToken = l.move()
	if tok.Type == token.IS && l.nextToken.Type == token.NOT {
		tok = token.Token{Type: token.IS_NOT, Literal: "IS NOT"}
		l.nextToken = l.move()
		return tok
	} else if tok.Type == token.NOT && l.nextToken.Type == token.IN {
		tok = token.Token{Type: token.NOT_IN, Literal: "NOT IN"}
		l.nextToken = l.move()
		return tok
	} else if tok.Type == token.NOT && l.nextToken.Type == token.BETWEEN {
		tok = token.Token{Type: token.NOT_BETWEEN, Literal: "NOT BETWEEN"}
		l.nextToken = l.move()
		return tok
	} else if tok.Type == token.NOT && l.nextToken.Type == token.LIKE {
		tok = token.Token{Type: token.NOT_LIKE, Literal: "NOT LIKE"}
		l.nextToken = l.move()
		return tok
	}

	return tok
}

func (l *Lexer) move() token.Token {
	var tok token.Token
	l.skipWhitespace()

	switch l.char {
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.PIPE2, Literal: "||"}
		} else {
			tok = newToken(token.PIPE, l.char)
		}

	case '=':
		tok = newToken(token.EQ, l.char)

	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.BANG_EQ, Literal: "!="}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.BANG_GT, Literal: "!>"}
		} else if l.peekChar() == '<' {
			l.readChar()
			tok = token.Token{Type: token.BANG_LT, Literal: "!<"}
		} else {
			tok = newToken(token.BANG, l.char)
		}

	case '(':
		tok = newToken(token.LPAREN, l.char)
	case ')':
		tok = newToken(token.RPAREN, l.char)
	case '[':
		tok = newToken(token.LBRACKET, l.char)
	case ']':
		tok = newToken(token.RBRACKET, l.char)

	case ',':
		tok = newToken(token.COMMA, l.char)
	case '+':
		tok = newToken(token.PLUS, l.char)

	case '#':
		tok = l.readSingleLineComment()

	// Do not support token `;` to reduce SQL injection risk.
	case ';':
		tok = token.NewIllegalToken("not support token `;`")
	case '-':
		if l.peekChar() == '-' {
			tok = l.readSingleLineComment()
		} else if l.peekChar() == '>' {
			l.readChar()
			if l.peekChar() == '>' {
				l.readChar()
				tok = token.Token{Type: token.PRT2, Literal: "->>"}
			} else {
				tok = token.Token{Type: token.PRT, Literal: "->"}
			}
		} else {
			tok = newToken(token.MINUS, l.char)
		}
	case '*':
		if l.peekChar() == '/' {
			l.readChar()
			// Not support `*/`
			tok = token.NewIllegalToken("not support SQL comment `*/`")
		} else {
			tok = newToken(token.ASTERISK, l.char)
		}
	case '/':
		if l.peekChar() == '*' {
			tok = l.readMultilineComment()
		} else {
			tok = newToken(token.SLASH, l.char)
		}
	case '%':
		tok = newToken(token.MOD, l.char)
	case '~':
		tok = newToken(token.TILDE, l.char)
	case '&':
		tok = newToken(token.AMP, l.char)
	case '^':
		tok = newToken(token.XOR, l.char)

	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			if l.peekChar() == '>' {
				l.readChar()
				tok = token.Token{Type: token.LT_EQ_GT, Literal: "<=>"}
			} else {
				tok = token.Token{Type: token.LT_EQ, Literal: "<="}
			}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "<>"}
		} else if l.peekChar() == '<' {
			l.readChar()
			tok = token.Token{Type: token.LT2, Literal: "<<"}
		} else {
			tok = newToken(token.LT, l.char)
		}

	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GT_EQ, Literal: ">="}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.RT2, Literal: ">>"}
		} else {
			tok = newToken(token.GT, l.char)
		}

	case '.':
		tok = newToken(token.PERIOD, l.char)

	case '\'':
		tok = l.readString()

	case '`':
		tok = l.readBackQuoteIdentifier()
	case '"':
		tok = l.readDoubleQuoteIdentifier()

	case '?':
		tok = newToken(token.QUESTION, l.char)

	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			tok = token.Token{Type: token.COLON2, Literal: "::"}
		} else {
			tok = newToken(token.COLON, l.char)
		}

	case EOF:
		tok.Literal = ""
		tok.Type = token.EOF

	default:
		if unicode.IsDigit(l.char) {
			tok = l.readNumber()
			return tok
		} else if l.isIdentifierStart() {
			ident := l.readIdentifier()
			tok = token.LookupIdent(ident)
			return tok
		}

		tok = token.Token{Type: token.ILLEGAL, Literal: string(l.char)}
	}

	l.readChar()
	return tok
}
