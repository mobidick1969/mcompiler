package main

import (
	"fmt"
	"mcompiler/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Hello %s\n", user.Name)
	repl.Start(os.Stdin, os.Stdout)
}
