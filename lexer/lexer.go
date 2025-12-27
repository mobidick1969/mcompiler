package lexer

import (
	"mcompiler/token"
)

type Lexer struct {
	input        string
	position     int  //current pos
	readPosition int  //next pos
	ch           byte //char at current pos
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readIdentifier() string {
	currentPos := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[currentPos:l.position]
}

func (l *Lexer) readNumber() string {
	currentPos := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[currentPos:l.position]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			tok = token.Token{Type: token.EQUAL, Literal: "=="}
			l.readChar()
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			tok = token.Token{Type: token.NOTEQUAL, Literal: "!="}
			l.readChar()
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = l.lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

var keywords = map[string]token.TokenType{
	"let":    token.LET,
	"fn":     token.FUNCTION,
	"if":     token.IF,
	"else":   token.ELSE,
	"return": token.RETURN,
	"true":   token.TRUE,
	"false":  token.FALSE,
}

func (l *Lexer) lookupIdent(literal string) token.TokenType {
	if tok, ok := keywords[literal]; ok {
		return tok
	}
	return token.IDENT
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(b byte) bool {
	return 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || b == '_'
}

func newToken(t token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    t,
		Literal: string(ch),
	}
}
