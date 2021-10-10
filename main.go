package main

import (
	"cc/cc"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid number of arguments\n", os.Args[0])
		return
	}

	fileName := os.Args[1]
	f, err := os.Open(fileName)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}

	err = cc.Compile(os.Stdout, []rune(string(content)))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
