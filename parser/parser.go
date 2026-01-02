package parser

import (
	"fmt"
	"mcompiler/ast"
	"mcompiler/lexer"
	"mcompiler/token"
	"strconv"
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

type Precedence int

const (
	_ = iota
	LOWEST
	LOW
	LOWER
	MID
	HIGH
	HIGHEST
)

func getPrecedence(tokenType token.TokenType) Precedence {
	switch tokenType {
	case token.FUNCTION:
		return HIGH
	case token.ASTERISK, token.SLASH:
		return MID
	case token.PLUS, token.MINUS:
		return LOWER
	case token.EQUAL, token.NOTEQUAL:
		return LOW
	case token.GT, token.LT:
		return LOWER
	default:
		return LOWEST
	}
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseBinaryExpression)
	p.registerInfix(token.MINUS, p.parseBinaryExpression)
	p.registerInfix(token.ASTERISK, p.parseBinaryExpression)
	p.registerInfix(token.SLASH, p.parseBinaryExpression)
	p.registerInfix(token.EQUAL, p.parseBinaryExpression)
	p.registerInfix(token.NOTEQUAL, p.parseBinaryExpression)
	p.registerInfix(token.LT, p.parseBinaryExpression)
	p.registerInfix(token.GT, p.parseBinaryExpression)

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
	case token.LBRACE:
		return p.parseBlockStatement()
	case token.IF:
		return p.parseIfStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	stmt := &ast.FunctionExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = []ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	} else {
		p.nextToken()
		stmt.Parameters = append(stmt.Parameters, ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			stmt.Parameters = append(stmt.Parameters, ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Consequence = p.parseBlockStatement()
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		stmt.Alternative = p.parseBlockStatement()
	}
	return stmt
}

func (p *Parser) parseBlockStatement() ast.Statement {
	stmt := &ast.BlockStatement{Token: p.curToken}
	stmt.Statements = []ast.Statement{}

	p.nextToken() // skip the current LBRACE

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt.Statements = append(stmt.Statements, p.parseStatement())
		p.nextToken()
	}

	return stmt
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

	prefixFn := p.prefixParseFns[p.curToken.Type]
	if prefixFn == nil {
		p.prefixFnError(p.curToken.Type)
		return nil
	}

	left := prefixFn()

	for precedence < p.peekPrecedence() && !p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()

		infixFn := p.infixParseFns[p.curToken.Type]
		if infixFn == nil {
			p.infixFnError(p.curToken.Type)
			return nil
		}
		left = infixFn(left)
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

func (p *Parser) parseIdentifier() ast.Expression {
	if p.peekTokenIs(token.LPAREN) {
		return p.parseFunctionInvokeExpression()
	}
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseFunctionInvokeExpression() ast.Expression {
	expr := &ast.FunctionInvokeExpression{Token: p.curToken}
	expr.Arguments = []ast.Expression{}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	for !p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
		expr.Arguments = append(expr.Arguments, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expr
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.currError(token.INT)
		return nil
	}
	return &ast.IntegerLiteral{Token: p.curToken, Value: value}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Type == token.TRUE}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	tok := p.curToken
	p.nextToken()
	right := p.parseExpression(HIGHEST)
	return &ast.UnaryExpression{Token: tok, Right: right}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	left := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return left
}

func (p *Parser) prefixFnError(tokenType token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", tokenType)
	p.errors = append(p.errors, msg)
}

func (p *Parser) infixFnError(tokenType token.TokenType) {
	msg := fmt.Sprintf("no infix parse function for %s found", tokenType)
	p.errors = append(p.errors, msg)
}
