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

const typePair = []tokenpair{
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
		return
	}
	s.err = errors.New("grammar error: " + s.current().Lexeme)
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
// compiles a class
func (s *parser) class() {
	io.WriteString(s.writer, "<class>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CLASS}})
	s.process([]tokenpair{tokenpair{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE}})
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACE}})
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

	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_BRACE}})
	io.WriteString(s.writer, "</class>")
}

// Compiles a static variable declaration or a field declaration
func (s *parser) classVarDec() {
	io.WriteString(s.writer, "<classVarDec>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_STATIC}, {jack_tokenizer.KEYWORD, jack_tokenizer.KW_FIELD}})
	s.helper_typeVarName()
	// (',' varName)
	for s.matches([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA}}) {
		s.process([]tokenpair{
			tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
		})
		s.process([]tokenpair{tokenpair{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE}})
	}
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_SEMICOLON}})
	io.WriteString(s.writer, "</classVarDec>")
}

// Compiles a complete method, function or constructor
func (s *parser) subroutine() {
	io.WriteString(s.writer, "<subroutineDec>")
	s.process([]tokenpair{
		tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CONSTRUCTOR},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FUNCTION},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_METHOD},
	})

	s.helper_type([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_VOID},
	})

	s.process([]tokenpair{
		{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
	})

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN},
	})

	s.parameters()

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_PAREN},
	})

	s.subroutineBody()
	io.WriteString(s.writer, "</subroutineDec>")
}

// Compiles a (possibly empty) parameters
// list. Does not handle the enclosing
// parantheses tokens (ands).
func (s *parser) parameters() {
	processTypeVarName := func() {
		s.process(typePair)
		s.process([]tokenpair{
			{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
		})
	}
	io.WriteString(s.writer, "<parameters>")
	if s.matches(typePair) {
		if s.matches([]tokenpair{
			{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
		}) {
			processTypeVarName()

			for s.matches([]tokenpair{
				{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
			}) {
				processTypeVarName();
			}
		}
	}
	io.WriteString(s.writer, "</parameters>")
}

// Compiles a subroutine's body
func (s *parser) subroutineBody() {
	io.WriteString(s.writer, "<subroutineBody>")
	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACE},
	})

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
	s.process([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_VAR},
	})

	s.helper_typeVarName()

	for s.matches([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
	}) {
		s.process([]tokenpair{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
		})
		s.process([]tokenpair{
			{jack_tokenizer.IDENTIFIER, NONE},
		})
	}

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_SEMICOLON},
	})
	
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
			case s.matches([]tokenpair(states[4])):
				s.do()
			case s.matches([]tokenpair(states[5])):
				s.returnStatement()
			}
		}
		io.WriteString(s.writer, "</statements>")
	} else {
		s.err = errors.New("unexpected lexeme " + s.current().Lexeme())
	}
}

// Compiles a let statement.
func (s *parser) letStatement() {
	io.writeString(s.writer, "<letStatement>")
	s.process([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_LET},
	})
	s.process([]tokenpair{
		{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
	})

	if s.matches([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACE},
	}) {
		s.process([]token{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACE},
		})

		s.expression()

		s.process([]token{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_BRACE},
		})
	}

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_EQUALS},
	})

	s.expression()

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_SEMICOLON},
	})

	io.writeString(s.writer, "</letStatement>")
}

// Compiles an if statement
// possibly with a trailing else clause
func (s *parser) ifStatement() {
	io.writeString(s.writer, "<ifStatement>")
	s.process([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_IF},
	})

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN},
	})

	s.expression()

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_PAREN},
	})

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACE},
	})

	s.statements()

	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_BRACE},
	})

	if s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_ELSE},
	}) {
		s.process([]tokenpair{
			{jack_tokenizer.KEYWORD, KW_ELSE},
		})

		s.process([]tokenpair{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACE},
		})

		s.statements()

		s.process([]tokenpair{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_PAREN},
		})
	}
	io.writeString(s.writer, "</ifStatement>")
}

// Compiles a while statement
func (s *parser) while() {
	io.WriteString(s.writer, "<whileStatement>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_WHILE}})
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN}})
	s.expression()
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN}})
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_BRACE}})
	s.statements()
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_RIGHT_BRACE}})
	io.WriteString(s.writer, "</whileStatement>")
}

// Compiles a do statement
func (s *parser) do() {
	io.WriteString(s.writer, "<doStatement>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_DO}})
	s.subroutine()
	s.process([]tokenpair{tokenpair{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_PERIOD}})
	io.WriteString(s.writer, "</doStatement>")
}

// Compiles a return statement
func (s *parser) returnStatement() {
	io.WriteString(s.writer, "<returnStatement>")
	s.process([]tokenpair{tokenpair{jack_tokenizer.KEYWORD, jack_tokenizer.KW_RETURN}})
	peek, err := s.peek()
	if err != nil {
		s.err = err
	}

	io.WriteString(s.writer, "</returnStatement>")
}

// Compiles an expression
func (s *parser) expression() {
	io.WriteString(s.writer, "<expression>")

	io.WriteString(s.writer, "</expression>")
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
// x | y x or yhelpe
func ParseGrammar(tokens []jack_tokenizer.Token) {
	parser := NewParser(tokens)

	return func()
}
