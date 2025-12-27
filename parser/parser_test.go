package parser

import (
	"mcompiler/ast"
	"mcompiler/lexer"
	"testing"
)

func TestParser_ParseReturnStatement(t *testing.T) {
	// input := `return a+b+c*d+e/f;
	// return 5;
	// return 10;
	// return a + b * c;
	// return a+b;
	input := `return 2*((a*c)+(((a+b)*c)+d)));`
	// return 2*((a*c)+(((a+b)*c)+d)));

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	expected := []string{
		// "return (((a + b) + (c * d)) + (e / f));",
		// "return 5;",
		// "return 10;",
		// "return (a + (b * c));",
		// "return (a + b);",
		"return ((a + b) * c);",
		// "return (2 * ((a * c) + (((a + b) * c) + d)))!;",
	}

	if len(p.Errors()) > 0 {
		t.Errorf("errors during parsing:%s", p.Errors())
	}

	for i, exp := range expected {
		if program.Statements[i].String() != exp {
			t.Errorf("stmt %d - wrong string. expected=%q, got=%q",
				i, exp, program.Statements[i].String())
		}
	}
}

func TestParser_ParseProgram(t *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar = 838383;
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct{ expectedIdentifier string }{
		{"x"},
		{"y"},
		{"foobar"},
	}
	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}

}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("stmt.TokenLiteral() != \"let\". got=%q", s.TokenLiteral())
		return false
	}
	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}
	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}
	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name not '%s'. got=%s", name, letStmt.Name)
		return false
	}
	return true
}
