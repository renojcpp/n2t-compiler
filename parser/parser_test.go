package jack_parser

import (
	"fmt"
	"os"
	"strings"
	"testing"

	jack_tokenizer "github.com/renojcpp/n2t-compiler/tokenizer"
)

func TestParseGrammar(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases
		{"/home/jr/school/cs3650/nand2tetris/projects/10/ArrayTest"},
		{"/home/jr/school/cs3650/nand2tetris/projects/10/ExpressionLessSquare"},
		{"/home/jr/school/cs3650/nand2tetris/projects/10/Square"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, err := os.ReadDir(tt.name)
			if err != nil {
				t.Errorf("no such directory: %s", tt.name)
			}

			for _, entry := range dirs {
				if strings.HasSuffix(entry.Name(), ".jack") {
					file, err := os.Open(fmt.Sprintf("%s/%s", tt.name, entry.Name()))
					if err != nil {
						t.Errorf("failed to open file: %s", tt.name)
					}

					tokens, err := jack_tokenizer.Tokenize(file)

					if err != nil {
						t.Errorf("failed to tokenize: %s", err)
					}

					toOut := ParseGrammar(tokens)
					fmt.Println("outputting")
					err = toOut(os.Stdout)

					if err != nil {
						t.Errorf("%s", err.Error())
					}
				}
			}
		})
	}
}
