package jack_parser

import (
	"errors"
	"fmt"
	"io"
	"strings"

	jack_tokenizer "github.com/renojcpp/n2t-compiler/tokenizer"
)

type tokenpair struct {
	tt jack_tokenizer.TokenType
	st jack_tokenizer.TokenSubtype
}

var typePair = []tokenpair{
	{jack_tokenizer.KEYWORD, jack_tokenizer.KW_INT},
	{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CHAR},
	{jack_tokenizer.KEYWORD, jack_tokenizer.KW_BOOLEAN},
	{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
}

type parser struct {
	tokens []jack_tokenizer.Token
	writer io.Writer
	index  int
	err    error
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
	if s.matches(pairs) {
		io.WriteString(s.writer, (s.current().Lexeme))
	} else {
		var ss strings.Builder
		for _, p := range pairs {
			ss.WriteString(fmt.Sprintf("{ %s, %s }", p.tt.String(), p.st.String()))
		}

		s.err = fmt.Errorf("%s %s %s %s: grammar error: got %s, wanted %s @ %d", s.tokens[s.index-2].Lexeme, s.tokens[s.index-1].Lexeme, s.tokens[s.index].Lexeme, s.tokens[s.index+1].Lexeme, s.current().Lexeme, ss.String(), s.index)
	}
	s.advance()
}

func (s *parser) matches(pairs []tokenpair) bool {
	curr := s.current()
	for _, p := range pairs {
		if p.tt == curr.Tokentype && p.st == curr.Subtype {
			return true
		}
	}

	return false
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

func (s *parser) advance() {
	s.index++
}

func (s *parser) parse() error {
	s.class()

	return s.err
}

func (s *parser) helper_type(additional []tokenpair) {
	tp := make([]tokenpair, 0)
	tp = append(tp, typePair...)
	tp = append(tp, additional...)
	s.process(tp)
}

func (s *parser) helper_typeVarName() {
	s.process(typePair)
	s.process([]tokenpair{
		{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
	})
}

func (s *parser) wrap(tag string, f func()) {
	io.WriteString(s.writer, "<"+tag+">")
	f()
	io.WriteString(s.writer, "</"+tag+">")
}

func (s *parser) symbolHelper(st jack_tokenizer.TokenSubtype) {
	io.WriteString(s.writer, "<symbol>")
	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, st},
	})
	io.WriteString(s.writer, "</symbol>")
}

func (s *parser) identifierHelper() {
	io.WriteString(s.writer, "<identifier>")
	s.process([]tokenpair{
		{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
	})
	io.WriteString(s.writer, "</identifier>")
}

func (s *parser) keywordHelper(st jack_tokenizer.TokenSubtype) {
	io.WriteString(s.writer, "<keyword>")
	s.process([]tokenpair{
		{jack_tokenizer.KEYWORD, st},
	})
	io.WriteString(s.writer, "</keyword>")
}

// compiles a class
func (s *parser) class() {
	io.WriteString(s.writer, "<class>")

	s.keywordHelper(jack_tokenizer.KW_CLASS)
	s.identifierHelper()
	s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)

	for s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_STATIC},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FIELD},
	}) {
		s.classVarDec()
	}

	for s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CONSTRUCTOR},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_METHOD},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FUNCTION},
	}) {
		s.subroutine()
	}

	s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACE)
	io.WriteString(s.writer, "</class>")
}

// Compiles a static variable declaration or a field declaration
func (s *parser) classVarDec() {
	io.WriteString(s.writer, "<classVarDec>")

	s.wrap("keyword", func() {
		s.process([]tokenpair{
			{jack_tokenizer.KEYWORD, jack_tokenizer.KW_STATIC},
			{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FIELD},
		})
	})

	s.helper_typeVarName()
	// (',' varName)
	for s.matches([]tokenpair{{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA}}) {
		s.symbolHelper(jack_tokenizer.SYM_COMMA)
		s.identifierHelper()
	}
	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)
	io.WriteString(s.writer, "</classVarDec>")
}

// Compiles a complete method, function or constructor
func (s *parser) subroutine() {
	io.WriteString(s.writer, "<subroutineDec>")
	s.wrap("keyword", func() {
		s.process([]tokenpair{
			{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CONSTRUCTOR},
			{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FUNCTION},
			{jack_tokenizer.KEYWORD, jack_tokenizer.KW_METHOD},
		})
	})

	s.helper_type([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_VOID},
	})

	s.identifierHelper()

	s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)

	s.parameters()

	s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)

	s.subroutineBody()
	io.WriteString(s.writer, "</subroutineDec>")
}

// Compiles a (possibly empty) parameters
// list. Does not handle the enclosing
// parantheses tokens (ands).
func (s *parser) parameters() {
	processTypeVarName := func() {
		s.process(typePair)
		s.identifierHelper()
	}
	io.WriteString(s.writer, "<parameters>")
	if s.matches(typePair) {

		processTypeVarName()

		for s.matches([]tokenpair{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
		}) {
			s.symbolHelper(jack_tokenizer.SYM_COMMA)
			processTypeVarName()
		}
	}
	io.WriteString(s.writer, "</parameters>")
}

// Compiles a subroutine's body
func (s *parser) subroutineBody() {
	io.WriteString(s.writer, "<subroutineBody>")
	s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)

	for s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_VAR},
	}) {
		s.varDec()
	}

	s.statements()

	io.WriteString(s.writer, "</subroutineBody>")
}

// Compiles a var declaration
func (s *parser) varDec() {
	io.WriteString(s.writer, "<varDec>")
	s.keywordHelper(jack_tokenizer.KW_VAR)

	s.helper_typeVarName()

	for s.matches([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
	}) {
		s.symbolHelper(jack_tokenizer.SYM_COMMA)
		s.identifierHelper()
	}

	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)

	io.WriteString(s.writer, "</varDec>")
}

// Compiles a sequeneces of statemnents
// Does not handle the enclosing curly
// bracket tokens { and }.
func (s *parser) statements() {
	states := []tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_LET},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_IF},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_WHILE},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_DO},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_RETURN},
	}

	if s.matches(states) {
		io.WriteString(s.writer, "<statements>")
		for s.matches(states) {
			switch {
			case s.matches([]tokenpair{states[0]}):
				s.letStatement()
			case s.matches([]tokenpair{states[1]}):
				s.ifStatement()
			case s.matches([]tokenpair{states[2]}):
				s.while()
			case s.matches([]tokenpair{states[3]}):
				s.do()
			case s.matches([]tokenpair{states[4]}):
				s.returnStatement()
			}
		}
		io.WriteString(s.writer, "</statements>")
	} else {
		s.err = errors.New("unexpected lexeme " + s.current().Lexeme)
	}
}

// Compiles a let statement.
func (s *parser) letStatement() {
	io.WriteString(s.writer, "<letStatement>")
	s.keywordHelper(jack_tokenizer.KW_LET)
	s.identifierHelper()

	if s.matches([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACK},
	}) {
		s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACK)

		s.expression()

		s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACK)
	}

	s.symbolHelper(jack_tokenizer.SYM_EQUALS)

	s.expression()
	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)

	io.WriteString(s.writer, "</letStatement>")
}

// Compiles an if statement
// possibly with a trailing else clause
func (s *parser) ifStatement() {
	io.WriteString(s.writer, "<ifStatement>")

	s.keywordHelper(jack_tokenizer.KW_IF)

	s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)

	s.expression()

	s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)

	s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)

	s.statements()

	s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACE)

	if s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_ELSE},
	}) {

		s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)

		s.statements()

		s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACE)
	}
	io.WriteString(s.writer, "</ifStatement>")
}

// Compiles a while statement
func (s *parser) while() {
	io.WriteString(s.writer, "<whileStatement>")
	s.keywordHelper(jack_tokenizer.KW_WHILE)
	s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)
	s.expression()
	s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)
	s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)
	s.statements()
	s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACE)
	io.WriteString(s.writer, "</whileStatement>")
}

// Compiles a do statement
func (s *parser) do() {
	io.WriteString(s.writer, "<doStatement>")
	s.keywordHelper(jack_tokenizer.KW_DO)
	s.subroutineCall()
	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)
	io.WriteString(s.writer, "</doStatement>")
}

// Compiles a return statement
func (s *parser) returnStatement() {
	begTerm := []tokenpair{
		{jack_tokenizer.STRING_CONSTANT, jack_tokenizer.NONE},
		{jack_tokenizer.INT_CONSTANT, jack_tokenizer.NONE},
		{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_PLUS},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_MINUS},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_TILDE},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_TRUE},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FALSE},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_NULL},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_THIS},
	}

	io.WriteString(s.writer, "<returnStatement>")
	s.keywordHelper(jack_tokenizer.KW_RETURN)
	if s.matches(begTerm) {
		s.expression()
	}
	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)
	io.WriteString(s.writer, "</returnStatement>")
}

// Compiles an expression
func (s *parser) expression() {
	op := []tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_PLUS},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_MINUS},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_ASTERISK},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_SLASH},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_ASTERISK},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_PIPE},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LESS_THAN},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_GREATER_THAN},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_EQUALS},
	}
	io.WriteString(s.writer, "<expression>")
	s.term()
	for s.matches(op) {
		s.process(op)
		s.term()
	}
	io.WriteString(s.writer, "</expression>")
}

// Compiles a term. If the current token is an
// identifier, the routine must resolve it
// into a variable, an array element, or a
// subroutine call. A single lookahead tokezn,
// which may be [, (, or ., suffices to distinguish
// between the possibilities.
// Any other token is not part of this term
// and should not be advanced over.
func (s *parser) term() {
	io.WriteString(s.writer, "<term>")
	switch s.current().Tokentype {
	case jack_tokenizer.IDENTIFIER:
		// variable, array element or subroutine
		peek, err := s.peek()
		if err != nil {
			s.err = nil
		} else {
			switch peek.Subtype {
			case jack_tokenizer.SYM_LEFT_PAREN:
			case jack_tokenizer.SYM_PERIOD:
				s.subroutineCall()
			case jack_tokenizer.SYM_LEFT_BRACK:
				// varname[expression]
				s.identifierHelper()
				s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACK)
				s.expression()
				s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACK)
			default:
				s.identifierHelper()
			}

		}
	case jack_tokenizer.INT_CONSTANT:
		s.process([]tokenpair{
			{jack_tokenizer.INT_CONSTANT, jack_tokenizer.NONE},
		})
	case jack_tokenizer.STRING_CONSTANT:
		s.process([]tokenpair{
			{jack_tokenizer.STRING_CONSTANT, jack_tokenizer.NONE},
		})
	case jack_tokenizer.SYMBOL:
		switch s.current().Subtype {
		case jack_tokenizer.SYM_LEFT_PAREN:
			s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)
			s.expression()
			s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)
		case jack_tokenizer.SYM_MINUS:
		case jack_tokenizer.SYM_TILDE:
			s.symbolHelper(s.current().Subtype)
			s.term()
		}
	case jack_tokenizer.KEYWORD:
		switch s.current().Subtype {
		case jack_tokenizer.KW_TRUE:
		case jack_tokenizer.KW_FALSE:
		case jack_tokenizer.KW_NULL:
		case jack_tokenizer.KW_THIS:
			s.keywordHelper(s.current().Subtype)
		}
	}

	io.WriteString(s.writer, "</term>")
}

func (s *parser) subroutineCall() {
	io.WriteString(s.writer, "<subroutineCall>")
	s.identifierHelper()

	t := s.current()
	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_PERIOD},
	})

	switch t.Subtype {
	case jack_tokenizer.SYM_PERIOD:
		s.identifierHelper()
		s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)

		fallthrough
	case jack_tokenizer.SYM_LEFT_PAREN:
		s.expressionList()
		s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)
	}
	io.WriteString(s.writer, "</subroutineCall>")
}

// Compiles a (possibly empty) comma-
// separated list of expression. Returns
// the number of expressions in the list
func (s *parser) expressionList() int {
	begTerm := []tokenpair{
		{jack_tokenizer.STRING_CONSTANT, jack_tokenizer.NONE},
		{jack_tokenizer.INT_CONSTANT, jack_tokenizer.NONE},
		{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_PLUS},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_MINUS},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_TILDE},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_TRUE},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FALSE},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_NULL},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_THIS},
	}
	count := 0
	io.WriteString(s.writer, "<expressionList>")
	if s.matches(begTerm) {
		count++
		s.expression()
		for s.matches([]tokenpair{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
		}) {
			count++
			s.symbolHelper(jack_tokenizer.SYM_COMMA)
			s.expression()
		}
	}
	io.WriteString(s.writer, "</expressionList>")

	return count
}

// x* : 0 or more
// ? one or more
// x y x followed by y
// x | y x or y
func ParseGrammar(tokens []jack_tokenizer.Token) func(io.Writer) error {
	return func(w io.Writer) error {
		parser := NewParser(tokens, w)

		err := parser.parse()
		return err
	}
}
