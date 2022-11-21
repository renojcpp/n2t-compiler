package jack_tokenwriter

import (
	"fmt"
	"os"
	"strings"
	"testing"

	jack_tokenizer "github.com/renojcpp/n2t-compiler/tokenizer"
)

func TestCreateXMLWriter(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases
		{"E:\\Documents\\csarch\\nand2tetris\\projects\\10\\ArrayTest"},
		{"E:\\Documents\\csarch\\nand2tetris\\projects\\10\\ExpressionLessSquare"},
		{"E:\\Documents\\csarch\\nand2tetris\\projects\\10\\Square"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirs, err := os.ReadDir(tt.name)
			if err != nil {
				t.Errorf("no such directory: %s", tt.name)
			}

			for _, entry := range dirs {
				if strings.HasSuffix(entry.Name(), ".jack") {
					file, err := os.Open(fmt.Sprintf("%s\\%s", tt.name, entry.Name()))
					if err != nil {
						t.Errorf("failed to open file: %s", tt.name)
					}

					tokens, err := jack_tokenizer.Tokenize(file)

					if err != nil {
						t.Errorf("failed to tokenize: %s", err)
					}

					toOut := CreateXMLWriter(tokens)
					fmt.Println("outputting")
					toOut(os.Stdout)
				}
			}
		})
	}
}
