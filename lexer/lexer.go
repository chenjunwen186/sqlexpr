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
	char         rune

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

func (l *Lexer) readDecimal() token.Token {
	var b bytes.Buffer

	var hasPeriod bool
	for isDigit(l.char) || (l.char == '.' && isDigit(l.peekChar())) {
		if l.char == '.' {
			if hasPeriod {
				return token.NewIllegalToken("invalid number literal")
			}
			hasPeriod = true
		}
		b.WriteRune(l.char)
		l.readChar()
	}

	return token.Token{Type: token.NUMBER, Literal: b.String()}
}

func (l *Lexer) readBinaryNumber() token.Token {
	var b bytes.Buffer
	b.WriteRune(l.char)
	l.readChar()
	b.WriteRune(l.char)
	l.readChar()

	var isIllegal bool
	for isDigit(l.char) {
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

func (l *Lexer) readOctalNumber() token.Token {
	var b bytes.Buffer
	b.WriteRune(l.char)
	l.readChar()
	b.WriteRune(l.char)
	l.readChar()

	var isIllegal bool
	for isDigit(l.char) {
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

func (l *Lexer) readHexadecimalNumber() token.Token {
	var b bytes.Buffer
	b.WriteRune(l.char)
	l.readChar()
	b.WriteRune(l.char)
	l.readChar()

	var isIllegal bool
	for isDigit(l.char) || unicode.IsLetter(l.char) {
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

func (l *Lexer) readScientificNotationNumber() token.Token {
	panic("not implemented")
}

func (l *Lexer) readString() token.Token {
	var b bytes.Buffer

	var hasComment bool
	for {
		l.readChar()

		if l.char == EOF {
			return token.NewIllegalToken(fmt.Sprintf("unexpected EOF: '%s", b.String()))
		}

		if l.char == '\'' {
			break
		}

		if l.char == '-' && l.peekChar() == '-' {
			hasComment = true
		}

		b.WriteRune(l.char)
	}

	if hasComment {
		return token.NewIllegalToken(fmt.Sprintf("not support SQL comment `--` in string literal: '%s'", b.String()))
	}

	return token.Token{Type: token.STRING, Literal: b.String()}
}

func (l *Lexer) readIdentifier() string {
	var b bytes.Buffer

	for isIdentifier(l.char) || isDigit(l.char) {
		b.WriteRune(l.char)
		l.readChar()
	}

	return b.String()
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

func (l *Lexer) isDecimalStart() bool {
	if isDigit(l.char) || (l.char == '.' && isDigit(l.peekChar())) {
		return true
	}
	return false
}

func isDigit(char rune) bool {
	return char >= '0' && char <= '9'
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
			tok = token.Token{Type: token.NOT_EQ1, Literal: "!="}
		} else {
			tok = token.NewIllegalToken("not support `!`")
		}

	case '(':
		tok = newToken(token.LPAREN, l.char)
	case ')':
		tok = newToken(token.RPAREN, l.char)
	case ',':
		tok = newToken(token.COMMA, l.char)
	case '+':
		tok = newToken(token.PLUS, l.char)
	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			// Not support `--``
			tok = token.NewIllegalToken("not support SQL comment `--`")
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
		tok = newToken(token.ASTERISK, l.char)
	case '/':
		if l.peekChar() == '*' {
			l.readChar()
			// Not support `/*`
			tok = token.NewIllegalToken("not support SQL comment `/*`")
		} else {
			tok = newToken(token.SLASH, l.char)
		}
	case '%':
		tok = newToken(token.MOD, l.char)
	case '~':
		tok = newToken(token.TILDE, l.char)
	case '&':
		tok = newToken(token.AMP, l.char)

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
			tok = token.Token{Type: token.NOT_EQ2, Literal: "<>"}
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

	case '\'':
		tok = l.readString()

	case '`':
		//TODO: IDENT_QUOTED
	case '"':
		//TODO
	case '[':
		//TODO

	case '0':
		next := l.peekChar()
		if next == 'b' || next == 'B' {
			// binary number
			tok = l.readBinaryNumber()
		} else if next == 'x' || next == 'X' {
			// hexadecimal number
			tok = l.readHexadecimalNumber()
		} else if isDigit(next) {
			// octal number
			tok = l.readOctalNumber()
		} else if next == '.' {
			// 0.xxx
			tok = l.readDecimal()
		} else {
			// 0
			tok = newToken(token.NUMBER, l.char)
		}

	case EOF:
		tok.Literal = ""
		tok.Type = token.EOF

	default:
		if l.isDecimalStart() {
			tok = l.readDecimal()
			return tok
		} else if l.char == '.' && !isDigit(l.peekChar()) {
			tok = newToken(token.PERIOD, l.char)
			l.readChar() // Move to next char
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
