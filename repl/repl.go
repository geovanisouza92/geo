package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/geovanisouza92/geo/eval"
	"github.com/geovanisouza92/geo/object"
)

const Prompt = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	scope := object.NewRootScope()

	for {
		fmt.Print(Prompt)
		// Read
		if !scanner.Scan() {
			return // EOF
		}
		input := scanner.Text()

		// Eval
		e, err := run(input, scope)
		if err != nil {
			io.WriteString(out, err.Error())
			continue
		}

		// Print
		if e != nil {
			io.WriteString(out, fmt.Sprintf("%v : %s", e.String(), e.Type()))
			io.WriteString(out, "\n")
		}

		// Loop
	}
}

func run(input string, scope *object.Scope) (object.Object, error) {
	s, err := eval.Compile(input)
	if err != nil {
		return nil, err
	}
	c := eval.NewContext(scope)
	return c.Eval(s), nil
}
