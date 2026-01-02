package parser

import (
	"mcompiler/ast"
	"mcompiler/lexer"
	"testing"
)

func TestParser_ParseExpressionStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x+y;", "( x + y );"},
		{"5;", "5;"},
		{"10;", "10;"},
		{"a+b;", "a+b;"},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("errors during parsing:%s", p.Errors())
			// t.FailNow()
		}
		for i, stmt := range program.Statements {
			if stmt.String() != tt.expected {
				t.Errorf("stmt %d - wrong string. expected=%q, got=%q",
					i, tt.expected, stmt.String())
			} else {
				t.Logf("stmt %d - correct string. expected=%q, got=%q",
					i, tt.expected, stmt.String())
			}
		}
	}
}
func TestParser_ParseReturnStatement(t *testing.T) {
	input := `
	return a+b;
	return 5;
	return 10;
	return a + b * c;
	return a+b;
	return (a+b)*c+d+e;
	return -5+3;
	return -(2*3+(3*(7+3)))+5;
	return ((5))+3;
	return (5+3);
	`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	expected := []string{
		"return (a + b);",
		"return 5;",
		"return 10;",
		"return (a + (b * c));",
		"return (a + b);",
		"return ((((a + b) * c) + d) + e);",
		"return ((- 5) + 3);",
		"return ((- ((2 * 3) + (3 * (7 + 3)))) + 5);",
		"return (5 + 3);",
		"return (5 + 3);",
	}

	if len(p.Errors()) > 0 {
		t.Errorf("errors during parsing:%s", p.Errors())
		// t.FailNow()
	}

	for i, exp := range expected {
		if program.Statements[i].String() != exp {
			t.Errorf("stmt %d - wrong string. expected=%q, got=%q",
				i, exp, program.Statements[i].String())
		} else {
			t.Logf("stmt %d - correct string. expected=%q, got=%q",
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

	tests := []struct{ expectedIdentifier string }{
		{"let x = 5;"},
		{"let y = 10;"},
		{"let foobar = 838383;"},
		// {"let func = fn(a,b){ return a+b;};"},
	}
	for i, tt := range tests {
		stmt := program.Statements[i]
		if stmt.String() != tt.expectedIdentifier {
			t.Errorf("stmt %d - wrong string. expected=%q, got=%q",
				i, tt.expectedIdentifier, stmt.String())
		} else {
			t.Logf("stmt %d - correct string. expected=%q, got=%q",
				i, tt.expectedIdentifier, stmt.String())
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
