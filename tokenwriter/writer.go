package jack_tokenwriter

import (
	"fmt"
	"io"
	"strings"

	jack_tokenizer "github.com/renojcpp/n2t-compiler/tokenizer"
)

func CreateXMLWriter(tokens []jack_tokenizer.Token) func(io.Writer) {
	var ss strings.Builder

	return func(w io.Writer) {
		for _, token := range tokens {
			tagname, ok := keyword2tag[token.Tokentype]
			if !ok {
				tagname = "unknown"
			}
			escaped, ok := escapeLexeme[token.Lexeme]
			if !ok {
				escaped = token.Lexeme
			}
			ss.WriteString(wrap(tagname, escaped, 1))
		}
		var root strings.Builder
		root.WriteString(wrap("tokens", "\n"+ss.String(), 0))

		io.WriteString(w, root.String())
	}
}

func wrap(tagName, content string, tabs int) string {
	return fmt.Sprintf("%s<%s>%s</%s>\n", strings.Repeat("\t", tabs), tagName, content, tagName)
}

var keyword2tag = map[jack_tokenizer.TokenType]string{
	jack_tokenizer.KEYWORD:         "keyword",
	jack_tokenizer.SYMBOL:          "symbol",
	jack_tokenizer.INT_CONSTANT:    "integerConstant",
	jack_tokenizer.STRING_CONSTANT: "stringConstant",
	jack_tokenizer.IDENTIFIER:      "identifier",
}

var escapeLexeme = map[string]string{
	"<":  "&lt;",
	">":  "&gt;",
	"\"": "&quot;",
	"&":  "&amp;",
}
