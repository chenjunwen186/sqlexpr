package lexer

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/chenjunwen186/sqlexpr/token"
)

type Lexer struct {
	input        []rune
	position     int
	nextPosition int
	char         rune
}

func New(input string) *Lexer {
	l := &Lexer{input: []rune(input)}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.nextPosition >= len(l.input) {
		l.char = 0
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

// Currently only decimal literals are supported
// TODO: support binary, octal, hexadecimal, scientific notation literals
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
	panic("not implemented")
}

func (l *Lexer) readOctalNumber() token.Token {
	panic("not implemented")
}

func (l *Lexer) readHexadecimalNumber() token.Token {
	panic("not implemented")
}

func (l *Lexer) readScientificNotationNumber() token.Token {
	panic("not implemented")
}

func (l *Lexer) readString() (string, error) {
	var b bytes.Buffer

	for {
		l.readChar()

		if l.char == '0' {
			return "", fmt.Errorf("unterminated string")
		}

		if l.char == '\'' {
			break
		}

		if l.char == '-' && l.peekChar() == '-' {
			return "", fmt.Errorf("`--` is not supported in string")
		}

		b.WriteRune(l.char)
	}

	return b.String(), nil
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
		str, err := l.readString()
		if err != nil {
			tok = token.Token{Type: token.ILLEGAL, Literal: err.Error()}
		} else {
			tok = token.Token{Type: token.STRING, Literal: str}
		}

	case '`':
		//TODO: IDENT_QUOTED
	case '"':
		//TODO
	case '[':
		//TODO

	case '0':
		next := l.peekChar()
		if next == 'b' || next == 'B' {
			//TODO
		} else if next == 'x' || next == 'X' {
			//TODO
		} else {
			//TODO
		}

	case 0:
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
			return token.LookupIdent(ident)
		}

		tok = token.Token{Type: token.ILLEGAL, Literal: string(l.char)}
	}

	l.readChar()
	return tok
}
