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

func (i *InfixExpression) Operator() string {
	return i.Token.Literal
}

func (i *InfixExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *InfixExpression) String() string {
	return "(" + i.Left.String() + " " + i.Operator() + " " + i.Right.String() + ")"
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
	Token         token.Token
	FunctionIdent Identifier
	Arguments     []Expression
}

func (c *CallExpression) TokenLiteral() string {
	return c.Token.Literal
}

func (c *CallExpression) String() string {
	args := make([]string, len(c.Arguments))
	for i, arg := range c.Arguments {
		args[i] = arg.String()
	}

	return c.FunctionIdent.Value + "(" + strings.Join(args, ", ") + ")"
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

type CaseExpression struct {
	Token token.Token
	Whens []When
	Else  Expression
}

func (c *CaseExpression) TokenLiteral() string {
	return c.Token.Literal
}

func (c *CaseExpression) String() string {
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

type InExpression struct {
	Token token.Token
}

func (i *InExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *InExpression) String() string {
	return i.Token.Literal
}

type NotInExpression struct {
}

type BetweenExpression struct {
	From Expression
	To   Expression
}

func (b *BetweenExpression) TokenLiteral() string {
	return token.BETWEEN
}

type NotBetweenExpression struct {
	From Expression
	To   Expression
}

func (n *NotBetweenExpression) TokenLiteral() string {
	return token.NOT + " " + token.BETWEEN
}

type LikeExpression struct {
	Match Expression
}

func (l *LikeExpression) TokenLiteral() string {
	return token.LIKE
}

func (l *LikeExpression) String() string {
	return "LIKE " + l.Match.String()
}

type NotLikeExpression struct {
	Match Expression
}

func (n *NotLikeExpression) TokenLiteral() string {
	return token.NOT + " " + token.LIKE
}

func (n *NotLikeExpression) String() string {
	return "NOT LIKE " + n.Match.String()
}
