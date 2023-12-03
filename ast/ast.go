package ast

import (
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

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (i *IntegerLiteral) TokenLiteral() string {
	return i.Token.Literal
}

func (i *IntegerLiteral) String() string {
	return i.Token.Literal
}

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (f *FloatLiteral) TokenLiteral() string {
	return f.Token.Literal
}

func (f *FloatLiteral) String() string {
	return f.Token.Literal
}

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (p *PrefixExpression) TokenLiteral() string {
	return p.Token.Literal
}

func (p *PrefixExpression) String() string {
	return "(" + p.Operator + p.Right.String() + ")"
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (i *InfixExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *InfixExpression) String() string {
	return "(" + i.Left.String() + " " + i.Operator + " " + i.Right.String() + ")"
}

type NullLiteral struct {
	Token token.Token
}

func (n *NullLiteral) TokenLiteral() string {
	return n.Token.Literal
}
func (n *NullLiteral) String() string {
	return n.Token.Literal
}

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) TokenLiteral() string {
	return b.Token.Literal
}

func (b *BooleanLiteral) String() string {
	return b.Token.Literal
}

type CallExpression struct {
	Token token.Token
	// TODO
	Function  Expression
	Arguments []Expression
}

func (c *CallExpression) TokenLiteral() string {
	return c.Token.Literal
}

func (c *CallExpression) String() string {
	// TODO
	return c.Token.Literal
}

type TestStringLiteral struct {
	Token token.Token
	Value string
}

func (t *TestStringLiteral) TokenLiteral() string {
	return t.Token.Literal
}

func (t *TestStringLiteral) String() string {
	return t.Token.Literal
}

type TestNumberLiteral struct {
	Token token.Token
	Value string
}

type CaseExpression struct {
	Token token.Token
	// TODO
}

type InExpression struct {
	Token token.Token
}

type BetweenExpression struct {
	Token token.Token
}
