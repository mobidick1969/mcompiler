package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"
	INT   = "INT"

	COMMA     = "," //
	SEMICOLON = ";" //

	ASSIGN = "=" //
	PLUS   = "+" //
	MINUS  = "-" //

	LPAREN = "(" //
	RPAREN = ")" //
	LBRACE = "{" //
	RBRACE = "}" //

	FUNCTION = "FUNCTION"
	LET      = "LET"
)
