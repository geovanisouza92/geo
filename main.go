package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"geo/ast"
	"geo/eval"
	"geo/object"
	"geo/repl"
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		repl.Start(os.Stdin, os.Stdout)
		return
	}

	c := eval.NewContext(object.NewRootScope())
	m := compileFile(flag.Arg(0))
	if m != nil {
		ev := c.Eval(m)
		if ev != eval.Null {
			fmt.Print(ev.String())
		}
	}
}

func compileFile(path string) *ast.Module {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err.Error())
		return nil
	}
	input := string(b)
	s, err := eval.Compile(input)
	if err != nil {
		fmt.Print(err.Error())
		return nil
	}
	return s
}
