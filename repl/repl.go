package repl

import (
	"bufio"
	"fmt"
	"io"
	"mcompiler/lexer"
	"mcompiler/token"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		for {
			tok := l.NextToken()
			if tok.Type == token.EOF {
				break
			}
			fmt.Printf("%+v\n", tok)
		}
	}
}
