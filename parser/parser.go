package parser

import (
	"fmt"

	"github.com/chenjunwen186/sqlexpr/ast"
	"github.com/chenjunwen186/sqlexpr/lexer"
	"github.com/chenjunwen186/sqlexpr/token"
)

// Precedence
const (
	_ int = iota
	LOWEST
	AS          // AS
	IN          // IN
	COND        // OR or AND
	BETWEEN     // BETWEEN
	NOT         // NOT
	IS          // IS
	EQUALS      // = <> <=>
	LESSGREATER // > or < <= >=
	SUM         // + or -
	PRODUCT     // * or /
	MOD         // %
	PREFIX      // -X or +X or ~X or DISTINCT
	CALL
	HIGHEST
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Each token precedence
var precedences = map[token.Type]int{
	token.EQ:       EQUALS,
	token.NOT_EQ1:  EQUALS,
	token.NOT_EQ2:  EQUALS,
	token.LT_EQ_GT: EQUALS,
	token.LT:       LESSGREATER,
	token.LT_EQ:    LESSGREATER,
	token.GT:       LESSGREATER,
	token.GT_EQ:    LESSGREATER,

	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.MOD:      MOD,
	token.TILDE:    PREFIX,
}

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token

	errors []error

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []error{}}
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)

	p.infixParseFns = make(map[token.Type]infixParseFn)

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) Errors() []error {
	return p.errors
}

func (p *Parser) peekError(t token.Token) {
	errMsg := fmt.Sprintf("expected next token to be %s, got %s instead", t.Type, p.peekToken.Type)
	p.errors = append(p.errors, fmt.Errorf(errMsg))
}

func (p *Parser) ParseExpression() ast.Expression {
	return p.parseExpression(LOWEST)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	for precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			//TODO
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) noPrefixParseFnError(t token.Type) {
	errMsg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, fmt.Errorf(errMsg))
}

// Looks up the precedence of the next token
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// Looks up the precedence of the current token
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expr.Right = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	return expr
}
