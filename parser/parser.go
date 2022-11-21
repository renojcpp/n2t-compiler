package jack_parser

import (
	"errors"
	"io"

	jack_tokenizer "github.com/renojcpp/n2t-compiler/tokenizer"
)

type tokenpair struct {
	tt jack_tokenizer.TokenType
	st jack_tokenizer.TokenSubtype
}

type parser struct {
	tokens []jack_tokenizer.Token
	writer io.Writer
	index  int
	err error
}

func NewParser(tokens []jack_tokenizer.Token, w io.Writer) *parser {
	return &parser{
		tokens,
		w,
		0,
		nil,
	}
}

func (s *parser) process(pairs []tokenpair) {
	curr := s.current()
	for _, p := range pairs {
		if p.tt == curr.Tokentype && p.st == curr.Subtype {
			s.writer.WriteString(curr.Lexeme)
			return
		}
	}
	err = errors.New("grammar error: " + curr.Lexeme)
}

func (s *parser) atEnd() bool {
	return s.index >= len(s.tokens)
}

func (s *parser) peek() (*jack_tokenizer.Token, error) {
	if s.atEnd() || s.index+1 >= len(s.tokens) {
		return nil, errors.New("error at end")
	}

	return &s.tokens[s.index+1], nil
}

func (s *parser) current() *jack_tokenizer.Token {
	return &s.tokens[s.index]
}

func (s *parser) parse() error {
	p.class()

	return s.err
}

func (s *parser) helper_type(additional ...tokenpair) {
	tokenpair = make([]tokenpair, 0)
	tokenpair = append(tokenpair, []tokenpair{		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_INT},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CHAR},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_BOOLEAN},
		{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
}...)
	tokenpair = append(tokenpair, additional...)
	process(tokenpair)
}

// compiles a class
func (s *parser) class() {
	w.writer.writerString("<class>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CLASS}})
	s.process([]tokenpair{tokenpair{jack_tokenizer.IDENTIFIER, NONE}})
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACE}})
	s.classVarDec()
	s.subroutine()
	w.writer.writerString("</class>")
}

// Compiles a static variable declaration or a field declaration
func (s *parser) classVarDec() {
	w.writer.writerString("<classVarDec>");
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_STATIC}, {jack_tokenizer.KEYWORD, jack_tokenizer.KW_FIELD}})
	s.process(jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACE)
	s.process(jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_BRACE)
	w.writer.writerString("</classVarDec>");
}

// Compiles a complete method, function or constructor
func (s *parser) subroutine() {

}

// Compiles a (possibly empty) parameters
// list. Does not handle the enclosing
// parantheses tokens (ands).
func (s *parser) parameters() {

}

// Compiles a subroutine's body
func (s *parser) subroutineBody() {

}

// Compiles a var declaration
func (s *parser) varDec() {

}

// Compiles a sequeneces of statemnents
// Does not handle the enclosing curly
// bracket tokens { and }.
func (s *parser) statements() {

}

// Compiles a let statement.
func (s *parser) letStatement() {

}

// Compiles an if statement
// possibly with a trailing else clause
func (s *parser) ifStatement() {

}

// Compiles a while statement
func (s *parser) while() {
	s.writer.writeString("<whileStatement>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_WHILE}})
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN}})
	s.expression()
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN}})
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_BRACE}})
	s.statements()
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_BRACE}})
	printf("</whileStatement>")
}

// Compiles a do statement
func (s *parser) do() {
	s.writer.writeString("<doStatement>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_DO}})
	s.subroutine()
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_PERIOD}})
	s.writer.writeString("</doStatement>")
}

// Compiles a return statement
func (s *parser) returnStatement() {
	s.writer.writeString("<returnStatement>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_RETURN}})
	peek, err = s.peek()
	if err != nil {
		s.err = err
	}


	s.writer.writeString("</returnStatement>")
}

// Compiles an expression
func (s *parser) expression() {

}

// Compiles a term. If the current token is an
// identifier, the routine must resolve it
// into a variable, an array element, or a
// subroutine call. A single lookahead token,
// which may be [, (, or ., suffices to distinguish
// between the possibilities.
// Any other token is not part of this term
// and should not be advanced over.
func (s *parser) term() {

}

// Compiles a (possibly empty) comma-
// separated list of expression. Returns
// the number of expressions in the list
func (s *parser) expressionList() int {

}

// x* : 0 or more
// ? one or more
// x y x followed by y
// x | y x or y
func ParseGrammar(tokens []jack_tokenizer.Token) {
	parser := NewParser(tokens)

	err =
}
