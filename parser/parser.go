package parser

import (
	"fmt"
	"mcompiler/ast"
	"mcompiler/lexer"
	"mcompiler/token"
)

type Parser struct {
	l              *lexer.Lexer
	curToken       token.Token
	peekToken      token.Token
	errors         []string
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	stmts := []ast.Statement{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		p.nextToken()
	}

	return &ast.Program{
		Statements: stmts,
	}
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseLetStatement() ast.Statement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken() //advance token for skipping return token

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	var left ast.Expression

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		left = p.parseExpression(LOWEST)
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	} else if p.curTokenIs(token.IDENT) {
		left = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.curTokenIs(token.INT) {
		left = &ast.IntegerLiteral{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.curTokenIs(token.MINUS) {
		tok := p.curToken
		p.nextToken()
		exp := p.parseExpression(HIGHEST)
		left = &ast.UnaryExpression{Token: tok, Right: exp}
	} else {
		p.errors = append(p.errors, fmt.Sprintf("expected next token to be Prefix, got %s instead", p.curToken.Type))
		return nil
	}

	for precedence < p.peekPrecedence() {
		p.nextToken()
		left = p.parseBinaryExpression(left)
	}

	return left
}

func (p *Parser) parseBinaryExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		Token: p.curToken,
		Left:  left,
	}
	precedence := p.currPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) currPrecedence() Precedence {
	return getPrecedence(p.curToken.Type)
}

func (p *Parser) peekPrecedence() Precedence {
	return getPrecedence(p.peekToken.Type)
}

type Precedence int

const (
	_ = iota
	LOWEST
	LOWER
	MID
	HIGH
	HIGHEST
)

func getPrecedence(tokenType token.TokenType) Precedence {
	switch tokenType {
	case token.ASTERISK, token.SLASH:
		return MID
	case token.PLUS, token.MINUS:
		return LOWER
	default:
		return LOWEST
	}
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectCurr(t token.TokenType) bool {
	if p.curTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.currError(t)
		return false
	}
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) currError(t token.TokenType) {
	msg := fmt.Sprintf("expected current token to be %s, got %s instead", t, p.curToken.Type)
	p.errors = append(p.errors, msg)
}
