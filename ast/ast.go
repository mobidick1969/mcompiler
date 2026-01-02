package ast

import (
	"bytes"
	"fmt"
	"mcompiler/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

func (es *ExpressionStatement) String() string {
	var out bytes.Buffer
	out.WriteString(es.Expression.String())
	out.WriteString(";")
	return out.String()
}

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode() {}

func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	} else {
		out.WriteString("nil")
	}
	out.WriteString(";")
	return out.String()
}

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}

func (bl *BooleanLiteral) TokenLiteral() string {
	return bl.Token.Literal
}

func (bl *BooleanLiteral) String() string { return bl.TokenLiteral() }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}

func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string { return i.Value }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}
func (il *IntegerLiteral) String() string { return fmt.Sprintf("%d", il.Value) }

type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (rs *ReturnStatement) statementNode() {}

func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.Value != nil {
		out.WriteString(rs.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type UnaryExpression struct {
	Token token.Token
	Right Expression
}

func (ue *UnaryExpression) expressionNode() {}
func (ue *UnaryExpression) TokenLiteral() string {
	return ue.Token.Literal
}
func (ue *UnaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(" + ue.TokenLiteral() + " ")
	out.WriteString(ue.Right.String())
	out.WriteString(")")
	return out.String()
}

type BinaryExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (be *BinaryExpression) expressionNode() {}
func (be *BinaryExpression) TokenLiteral() string {
	return be.Token.Literal
}

func (be *BinaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(be.Left.String())
	out.WriteString(" " + be.Token.Literal + " ")
	if be.Right != nil {
		out.WriteString(be.Right.String())
	} else {
		out.WriteString("nil")
	}
	out.WriteString(")")

	return out.String()
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}
func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	for _, stmt := range bs.Statements {
		out.WriteString(stmt.String())
	}
	out.WriteString("}")
	return out.String()
}

type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence Statement
	Alternative Statement
}

func (is *IfStatement) statementNode() {}
func (is *IfStatement) TokenLiteral() string {
	return is.Token.Literal
}
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString(is.TokenLiteral() + " ")
	out.WriteString(is.Condition.String())
	out.WriteString(" ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

type FunctionInvokeExpression struct {
	Token     token.Token
	Arguments []Expression
}

func (fie *FunctionInvokeExpression) expressionNode() {}
func (fie *FunctionInvokeExpression) TokenLiteral() string {
	return fie.Token.Literal
}
func (fie *FunctionInvokeExpression) String() string {
	var out bytes.Buffer
	out.WriteString(fie.TokenLiteral() + "(")
	for i, arg := range fie.Arguments {
		out.WriteString(arg.String())
		if i < len(fie.Arguments)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(")")
	return out.String()
}

type FunctionExpression struct {
	Token      token.Token
	Parameters []Identifier
	Body       Statement
}

func (fs *FunctionExpression) expressionNode() {}
func (fs *FunctionExpression) TokenLiteral() string {
	return fs.Token.Literal
}
func (fs *FunctionExpression) String() string {
	var out bytes.Buffer
	out.WriteString(fs.TokenLiteral() + "(")
	for i, param := range fs.Parameters {
		out.WriteString(param.String())
		if i < len(fs.Parameters)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(")")
	out.WriteString(fs.Body.String())
	return out.String()
}
