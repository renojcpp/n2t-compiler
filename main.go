package main

import (
	"fmt"
	"os"
	"strings"

	jack_tokenizer "github.com/renojcpp/n2t-compiler/tokenizer"
)

func main() {
	args := os.Args[1:]

	for _, arg := range args {
		dirs, err := os.ReadDir(arg)
		if err != nil {
			fmt.Errorf("no such directory: %s", arg)
		}

		for _, entry := range dirs {
			if strings.HasSuffix(entry.Name(), ".jack") {
				file, err := os.Open(fmt.Sprintf("%s/%s", arg, entry.Name()))
				if err != nil {
					fmt.Errorf("failed to open file: %s", arg)
				}

				tokens, err := jack_tokenizer.Tokenize(file)

				if err != nil {
					fmt.Errorf("failed to tokenize: %s", err)
				}

				toOut := jack_compiler.ParseGrammar(tokens)
				fmt.Println("outputting")
				err = toOut(os.Stdout)

				if err != nil {
					fmt.Errorf("%s", err.Error())
				}
			}
		}
	}
}
