package main

import (
	"cc/cc"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	var content string
	if len(os.Args) == 2 {
		fileName := os.Args[1]
		f, err := os.Open(fileName)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}

		contentBytes, err := ioutil.ReadAll(f)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
		content = string(contentBytes)
	} else if len(os.Args) == 3 && (os.Args[1] == "-c" || os.Args[1] == "--code") {
		content = os.Args[2]
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "%s: invalid number of arguments\n", os.Args[0])
		return
	}

	err := cc.Compile(os.Stdout, []rune(content))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
