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
		{"x+y;", "(x + y);"},
		{"5;", "5;"},
		{"10;", "10;"},
		{"a+b;", "(a + b);"},
		{"!true;", "(! true);"},
		{"!false;", "(! false);"},
		{"80+3>y;", "((80 + 3) > y);"},
		{"f(y+3, z,10)*2;", "(f((y + 3), z, 10) * 2);"},
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

func TestParser_ParseFunctionStatement(t *testing.T) {
	input := `fn(x, y) { x + y; return 25; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Errorf("errors during parsing: %s", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statements[0] is not *ast.FunctionStatement. got=%T",
			program.Statements[0])
	}

	expr, ok := stmt.Expression.(*ast.FunctionExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.FunctionExpression. got=%T",
			stmt.Expression)
	}

	if len(expr.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d",
			len(expr.Parameters))
	}

	if expr.Parameters[0].Value != "x" {
		t.Fatalf("parameter 0 is not 'x'. got=%q", expr.Parameters[0].Value)
	}

	if expr.Parameters[1].Value != "y" {
		t.Fatalf("parameter 1 is not 'y'. got=%q", expr.Parameters[1].Value)
	}

	expectedString := "fn(x, y){(x + y);return 25;};"
	if stmt.String() != expectedString {
		t.Errorf("stmt.String() wrong. expected=%q, got=%q", expectedString, stmt.String())
	}
}

func TestParser_ParseIfStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"if (x > y) { x; }", "if (x > y) {x;}"},
		{"if (x > y) { x; } else { y; }", "if (x > y) {x;} else {y;}"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Errorf("errors during parsing: %s", p.Errors())
		}

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.IfStatement)
		if !ok {
			t.Fatalf("statements[0] is not *ast.IfStatement. got=%T",
				program.Statements[0])
		}

		if stmt.String() != tt.expected {
			t.Errorf("stmt.String() wrong. expected=%q, got=%q", tt.expected, stmt.String())
		}
	}
}

func TestParser_ParseBlockStatement(t *testing.T) {
	input := `{ let x = 5; let y = 10; let f = fn(x, y) { x + y; }; let y=f(a,10);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Errorf("errors during parsing: %s", p.Errors())
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("statements[0] is not *ast.BlockStatement. got=%T",
			program.Statements[0])
	}

	if len(stmt.Statements) != 4 {
		t.Fatalf("block statement does not contain 4 statements. got=%d",
			len(stmt.Statements))
	}

	expectedString := "{let x = 5;let y = 10;let f = fn(x, y){(x + y);};let y = f(a, 10);}"
	if stmt.String() != expectedString {
		t.Errorf("stmt.String() wrong. expected=%q, got=%q", expectedString, stmt.String())
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
