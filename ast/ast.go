package ast

import (
	"strings"

	"github.com/chenjunwen186/sqlexpr/token"
)

type Expression interface {
	TokenLiteral() string
	String() string
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Value
}

type PrefixExpression struct {
	Token token.Token
	Right Expression
}

func (p *PrefixExpression) Operator() string {
	return p.Token.Literal
}

func (p *PrefixExpression) TokenLiteral() string {
	return p.Token.Literal
}

func (p *PrefixExpression) String() string {
	var space string
	switch p.Token.Type {
	case token.DISTINCT:
		space = " "
	}

	return "(" + p.Operator() + space + p.Right.String() + ")"
}

type InfixExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (i *InfixExpression) Operator() token.Type {
	return i.Token.Type
}

func (i *InfixExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *InfixExpression) String() string {
	return "(" + i.Left.String() + " " + string(i.Operator()) + " " + i.Right.String() + ")"
}

type NullLiteral struct {
	token.Token
}

func (n *NullLiteral) TokenLiteral() string {
	return n.Token.Literal
}
func (n *NullLiteral) String() string {
	return n.Token.Literal
}

type BooleanLiteral struct {
	token.Token
}

func (b *BooleanLiteral) TokenLiteral() string {
	return b.Token.Literal
}

func (b *BooleanLiteral) String() string {
	return b.Token.Literal
}

func (b *BooleanLiteral) Value() bool {
	return b.Token.Type == token.TRUE
}

type CallExpression struct {
	Token     token.Token
	Fn        Expression
	Arguments []Expression
}

func (c *CallExpression) TokenLiteral() string {
	return c.Token.Literal
}

func (c *CallExpression) String() string {
	args := make([]string, len(c.Arguments))
	for i, arg := range c.Arguments {
		args[i] = arg.String()
	}

	return c.Fn.String() + "(" + strings.Join(args, ", ") + ")"
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (t *StringLiteral) TokenLiteral() string {
	return t.Token.Literal
}

func (t *StringLiteral) String() string {
	return t.Token.Literal
}

type NumberLiteral struct {
	token.Token
}

func (t *NumberLiteral) TokenLiteral() string {
	return t.Literal
}

func (t *NumberLiteral) String() string {
	return t.Literal
}

type CaseWhenExpression struct {
	Token token.Token
	Whens []When
	Else  Expression
}

func (c *CaseWhenExpression) TokenLiteral() string {
	return c.Token.Literal
}

func (c *CaseWhenExpression) String() string {
	var whens []string
	for _, when := range c.Whens {
		whens = append(whens, when.String())
	}

	var elseStr string
	if c.Else != nil {
		elseStr = " ELSE " + c.Else.String()
	}

	return "CASE " + strings.Join(whens, " ") + elseStr + " END"
}

type When struct {
	Cond Expression
	Then Expression
}

func (c *When) String() string {
	return "WHEN " + c.Cond.String() + " THEN " + c.Then.String()
}

type BetweenExpression struct {
	Left  Expression
	Range Expression
}

func (b *BetweenExpression) TokenLiteral() string {
	return token.BETWEEN
}

func (b *BetweenExpression) String() string {
	return "(" + b.Left.String() + " " + token.BETWEEN + " " + b.Range.String() + ")"
}

type NotBetweenExpression struct {
	Left  Expression
	Range Expression
}

func (n *NotBetweenExpression) TokenLiteral() string {
	return token.NOT + " " + token.BETWEEN
}

func (n *NotBetweenExpression) String() string {
	return "(" + n.Left.String() + " " + token.NOT + " " + token.BETWEEN + " " + n.Range.String() + ")"
}

type TupleExpression struct {
	Expressions []Expression
}

func (t *TupleExpression) TokenLiteral() string {
	return token.LPAREN + token.RPAREN
}

func (t *TupleExpression) String() string {
	var exprs []string
	for _, expr := range t.Expressions {
		exprs = append(exprs, expr.String())
	}
	return token.LPAREN + strings.Join(exprs, ", ") + token.RPAREN
}
