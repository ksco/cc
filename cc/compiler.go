package cc

import (
	"io"
)

func Compile(w io.Writer, s []rune) error {
	scanner := NewScanner(s)
	tokens, err := scanner.Scan()
	if err != nil {
		return err
	}

	parser := NewParser(tokens)
	objects, err := parser.Parse()
	if err != nil {
		return err
	}

	gen := NewCodegen(w, objects)
	if err = gen.Gen(); err != nil {
		return err
	}

	return nil
}
