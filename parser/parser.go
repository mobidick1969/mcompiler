package parser

import (
	"fmt"
	"mcompiler/ast"
	"mcompiler/lexer"
	"mcompiler/token"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
}

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
		return nil
	}
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

	stmt.Value = p.parseExpression(token.ILLEGAL)

	return stmt
}

func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken() //advance token for skipping return token

	stmt.Value = p.parseExpression(token.ILLEGAL)

	return stmt
}

func (p *Parser) parseExpression(op token.TokenType) ast.Expression {
	var left ast.Expression

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		left = p.parseExpression(token.ILLEGAL)
		fmt.Printf("left:%+v, op:%+v, curToken:%+v\n", left, op, p.curToken)
		if !p.expectPeek(token.RPAREN) {
			fmt.Printf("left:%+v, op:%+v, curToken:%+v\n", left, op, p.curToken)
			p.peekError(token.RPAREN)
			return nil
		}
		fmt.Printf("left:%+v, op:%+v, curToken:%+v\n", left, op, p.curToken)
	} else if p.curTokenIs(token.IDENT) {
		left = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else if p.curTokenIs(token.INT) {
		left = &ast.IntegerLiteral{Token: p.curToken, Value: p.curToken.Literal}
	} 

	for (getPrecedence(op) < getPrecedence(p.peekToken.Type)) && !p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		fmt.Printf("left:%+v, op:%+v, curToken:%+v\n", left, op, p.curToken)

		tok := p.curToken
		p.nextToken()

		left = &ast.BinaryExpression{
			Token: tok,
			Left:  left,
			Right: p.parseExpression(tok.Type),
		}
	}

	return left
}

func getPrecedence(tokenType token.TokenType) int {
	switch tokenType {
	case token.LPAREN, token.RPAREN:
		return 3
	case token.ASTERISK, token.SLASH:
		return 2
	case token.PLUS, token.MINUS:
		return 1
	default:
		return 0
	}
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
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
