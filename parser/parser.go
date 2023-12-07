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
	prefixParseFn func() (ast.Expression, error)
	infixParseFn  func(ast.Expression) (ast.Expression, error)
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

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.EOF, p.parseEOF)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.PLUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ1, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ2, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ_GT, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	return p
}

func (p *Parser) ParseExpression() (ast.Expression, error) {
	return p.parseExpression(LOWEST)
}

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil, fmt.Errorf("no prefix parse function for %s found", p.curToken.Type)
	}
	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}

	for !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return nil, fmt.Errorf("no infix parse function for %s found", p.peekToken.Type)
		}
		p.nextToken()
		leftExp, err = infix(leftExp)
		if err != nil {
			return nil, err
		}
	}
	return leftExp, nil
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

func (p *Parser) expectPeek(t token.Type) error {
	if p.peekToken.Type == t {
		p.nextToken()
		return nil
	}
	return fmt.Errorf("expected next token to be %s, got %s instead", t, p.curToken.Type)
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
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

func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	var err error
	expr.Right, err = p.parseExpression(PREFIX)
	return expr, err
}

func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	expr := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	var err error
	expr.Right, err = p.parseExpression(precedence)
	return expr, err
}

func (p *Parser) parseEOF() (ast.Expression, error) {
	return nil, nil
}

func (p *Parser) parseIdentifier() (ast.Expression, error) {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}, nil
}

func (p *Parser) parseBooleanLiteral() (ast.Expression, error) {
	return &ast.BooleanLiteral{Token: p.curToken}, nil
}

func (p *Parser) parseNullLiteral() (ast.Expression, error) {
	return &ast.NullLiteral{Token: p.curToken}, nil
}

func (p *Parser) parseStringLiteral() (ast.Expression, error) {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}, nil
}

func (p *Parser) parseNumberLiteral() (ast.Expression, error) {
	return &ast.NumberLiteral{Token: p.curToken}, nil
}

func (p *Parser) parseCaseWhenExpression() (ast.Expression, error) {
	// TODO
	return nil, nil
}

func (p *Parser) parseGroupedExpression() (ast.Expression, error) {
	p.nextToken()
	expr, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if err := p.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) parseCallExpression(fn ast.Expression) (ast.Expression, error) {
	// SQL only support identifier as callee
	ident, ok := fn.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier, got %s", fn.TokenLiteral())
	}

	expr := &ast.CallExpression{Token: p.curToken, FunctionIdent: *ident}
	var err error
	expr.Arguments, err = p.parseExpressionList(token.RPAREN)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *Parser) parseExpressionList(end token.Type) ([]ast.Expression, error) {
	var list []ast.Expression
	if p.peekTokenIs(end) {
		p.nextToken()
		return list, nil
	}

	p.nextToken()
	v, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	list = append(list, v)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		v, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	if err := p.expectPeek(end); err != nil {
		return nil, err
	}

	return list, nil
}
