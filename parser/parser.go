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
	AS   // AS
	COND // OR or AND
	IN   // IN
	// BETWEEN     // BETWEEN
	NOT         // NOT
	EQUALS      // = <> <=>
	LESSGREATER // > or < <= >=
	SUM         // + or -
	PRODUCT     // * or /
	MOD         // %
	IS          // IS
	PREFIX      // -X or +X or ~X or DISTINCT
	CALL
	HIGHEST
)

const PASS = LOWEST

type (
	prefixParseFn func() (ast.Expression, error)
	infixParseFn  func(ast.Expression) (ast.Expression, error)
)

// Each token precedence
var precedences = map[token.Type]int{
	token.EOF:    LOWEST,
	token.COMMA:  LOWEST,
	token.RPAREN: LOWEST,
	token.WHEN:   LOWEST,
	token.THEN:   LOWEST,
	token.ELSE:   LOWEST,
	token.END:    LOWEST,

	token.IN:          IN,
	token.NOT_IN:      IN,
	token.LIKE:        IN,
	token.NOT_LIKE:    IN,
	token.BETWEEN:     IN,
	token.NOT_BETWEEN: IN,

	token.IS:     IS,
	token.IS_NOT: IS,

	token.EQ:      EQUALS,
	token.NOT_EQ1: EQUALS,
	token.NOT_EQ2: EQUALS,

	token.LT_EQ_GT: LESSGREATER, // TODO
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

	token.AND: COND,
	token.OR:  COND,

	token.LPAREN: CALL,
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
	p.registerPrefix(token.EOF, p.parseUnexpectedEOF)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.PLUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedOrTupleExpression)
	p.registerPrefix(token.DISTINCT, p.parsePrefixExpression)
	p.registerPrefix(token.CASE, p.parseCaseWhenExpression)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	// p.registerInfix(token.AS, p.parseInfixExpression)
	p.registerInfix(token.IN, p.parseInfixExpression)
	p.registerInfix(token.NOT_IN, p.parseInfixExpression)
	p.registerInfix(token.BETWEEN, p.parseBetweenExpression)
	p.registerInfix(token.NOT_BETWEEN, p.parseNotBetweenExpression)
	p.registerInfix(token.IS, p.parseInfixExpression)
	p.registerInfix(token.IS_NOT, p.parseInfixExpression)
	p.registerInfix(token.LIKE, p.parseInfixExpression)
	p.registerInfix(token.NOT_LIKE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ1, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ2, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ_GT, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	return p
}

func (p *Parser) ParseExpression() (ast.Expression, error) {
	if p.l.Len() == 0 {
		return nil, nil
	}

	return p.parseExpression(LOWEST)
}

func (p *Parser) parseExpression(precedence int) (ast.Expression, error) {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil, fmt.Errorf("no prefix parse function for %q found", p.curToken.Type)
	}

	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}

	for {
		peekPrecedence, err := p.peekPrecedence()
		if err != nil {
			return nil, err
		}
		if precedence >= peekPrecedence {
			break
		}

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
	return fmt.Errorf("expected next token to be %q, got %q instead", t, p.curToken.Type)
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

// Looks up the precedence of the next token
func (p *Parser) peekPrecedence() (int, error) {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p, nil
	}

	return 0, fmt.Errorf("peekPrecedence(): no precedence found for %q, literal: %q", p.peekToken.Type, p.peekToken.Literal)
}

// Looks up the precedence of the current token
func (p *Parser) curPrecedence() (int, error) {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p, nil
	}

	return 0, fmt.Errorf("curPrecedence(): no precedence found for %s, literal: %s", p.curToken.Type, p.curToken.Literal)
}

func (p *Parser) parsePrefixExpression() (ast.Expression, error) {
	expr := &ast.PrefixExpression{
		Token: p.curToken,
	}
	p.nextToken()
	var err error
	expr.Right, err = p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}

	return expr, err
}

func (p *Parser) parseInfixExpression(left ast.Expression) (ast.Expression, error) {
	expr := &ast.InfixExpression{
		Token: p.curToken,
		Left:  left,
	}

	precedence, err := p.curPrecedence()
	if err != nil {
		return nil, err
	}

	p.nextToken()
	expr.Right, err = p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

var EOFErr = fmt.Errorf("unexpected EOF error")

func (p *Parser) parseUnexpectedEOF() (ast.Expression, error) {
	return nil, EOFErr
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
	if !p.peekTokenIs(token.WHEN) {
		return nil, fmt.Errorf("CASE must have at least one WHEN")
	}

	var whens []ast.When
	for p.peekTokenIs(token.WHEN) {
		p.nextToken()
		p.nextToken()
		cond, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}

		if err := p.expectPeek(token.THEN); err != nil {
			return nil, err
		}
		p.nextToken()

		then, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}

		whens = append(whens, ast.When{Cond: cond, Then: then})
	}
	if len(whens) == 0 {
		return nil, fmt.Errorf("CASE must have at least one WHEN")
	}

	var elseExpr ast.Expression
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		p.nextToken()
		var err error
		elseExpr, err = p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
	}

	if err := p.expectPeek(token.END); err != nil {
		return nil, err
	}

	return &ast.CaseWhenExpression{Token: p.curToken, Whens: whens, Else: elseExpr}, nil
}

func (p *Parser) parseGroupedOrTupleExpression() (ast.Expression, error) {
	if p.peekToken.Type == token.RPAREN {
		return nil, fmt.Errorf("empty `()` is not supported")
	}

	p.nextToken()
	expr, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return expr, nil
	}

	if p.peekToken.Type != token.COMMA {
		return nil, fmt.Errorf("expected `)` or `,`, got %s", p.peekToken.Type)
	}

	var list []ast.Expression
	list = append(list, expr)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		v, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}

		list = append(list, v)
	}
	if err := p.expectPeek(token.RPAREN); err != nil {
		return nil, err
	}

	return &ast.TupleExpression{Expressions: list}, nil
}

func (p *Parser) parseCallExpression(fn ast.Expression) (ast.Expression, error) {
	// SQL only support identifier as callee
	ident, ok := fn.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier, got %s", fn.TokenLiteral())
	}

	expr := &ast.CallExpression{Token: p.curToken, FnName: *ident}
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

func (p *Parser) parseBetweenExpression(left ast.Expression) (ast.Expression, error) {
	p.nextToken()
	r, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	v, ok := r.(*ast.InfixExpression)
	if !ok {
		return nil, fmt.Errorf("expected infix expression, got %s", r.TokenLiteral())
	}
	if v.Operator() != token.AND {
		return nil, fmt.Errorf("expected AND, got %s", v.Operator())
	}

	expr := &ast.BetweenExpression{
		Left:  left,
		Range: v,
	}

	return expr, nil
}

func (p *Parser) parseNotBetweenExpression(left ast.Expression) (ast.Expression, error) {
	p.nextToken()
	r, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	v, ok := r.(*ast.InfixExpression)
	if !ok {
		return nil, fmt.Errorf("expected infix expression, got %s", r.TokenLiteral())
	}
	if v.Operator() != token.AND {
		return nil, fmt.Errorf("expected AND, got %s", v.Operator())
	}

	expr := &ast.NotBetweenExpression{
		Left:  left,
		Range: v,
	}

	return expr, nil
}
