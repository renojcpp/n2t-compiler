package jack_compiler

import (
	"errors"
	"fmt"
	"io"
	"strconv"
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
	index  int

	err          error
	subroutineSt *SymbolTable
	classSt      *SymbolTable
	vmWriter     VMWriter

	className   string
	labelNumber int
}

type symboldata struct {
	name   string
	symbol FieldType
	index  int
}

func NewParser(tokens []jack_tokenizer.Token, vmw VMWriter) *parser {
	return &parser{
		tokens,
		0,
		nil,
		NewSymbolTable(),
		NewSymbolTable(),
		vmw,
		"",
		0,
	}
}

func (s *parser) resolveSymbol(sym string) symboldata {
	res := s.subroutineSt.KindOf(Name(sym))

	if res == NONE {
		res = s.classSt.KindOf(Name(sym))
		if res == NONE {
			return symboldata{
				sym,
				NONE,
				0,
			}
		}

		return symboldata{
			sym,
			res,
			s.classSt.IndexOf(Name(sym)),
		}
	}

	return symboldata{
		sym,
		res,
		s.subroutineSt.IndexOf(Name(sym)),
	}
}

func (s *parser) process(pairs []tokenpair) (*jack_tokenizer.Token, error) {
	var err error
	if !s.matches(pairs) {
		var ss strings.Builder
		for _, p := range pairs {
			ss.WriteString(fmt.Sprintf("{ %s, %s }", p.tt.String(), p.st.String()))
		}

		s.err = fmt.Errorf("%s %s %s %s: grammar error: got %s, wanted %s @ %d", s.tokens[s.index-2].Lexeme, s.tokens[s.index-1].Lexeme, s.tokens[s.index].Lexeme, s.tokens[s.index+1].Lexeme, s.Current().Lexeme, ss.String(), s.index)
		err = s.err
	}
	ret := s.Current()
	s.Advance()

	return ret, err
}

func (s *parser) matches(pairs []tokenpair) bool {
	curr := s.Current()
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

func (s *parser) Current() *jack_tokenizer.Token {
	return &s.tokens[s.index]
}

func (s *parser) Advance() {
	s.index++
}

func (s *parser) Parse() error {
	s.Class()

	return s.err
}

func (s *parser) helper_type(additional []tokenpair) (*jack_tokenizer.Token, error) {
	tp := make([]tokenpair, 0)
	tp = append(tp, typePair...)
	tp = append(tp, additional...)

	return s.process(tp)
}

func (s *parser) symbolHelper(st jack_tokenizer.TokenSubtype) (*jack_tokenizer.Token, error) {
	return s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, st},
	})
}

func (s *parser) identifierHelper() (*jack_tokenizer.Token, error) {
	return s.process([]tokenpair{
		{jack_tokenizer.IDENTIFIER, jack_tokenizer.NONE},
	})
}

func (s *parser) keywordHelper(st jack_tokenizer.TokenSubtype) (*jack_tokenizer.Token, error) {
	return s.process([]tokenpair{
		{jack_tokenizer.KEYWORD, st},
	})
}

// compiles a Class
func (s *parser) Class() {
	s.classSt.Reset()
	s.subroutineSt.Reset()

	s.keywordHelper(jack_tokenizer.KW_CLASS)
	token, err := s.identifierHelper()
	if err == nil {
		s.className = token.Lexeme
	}
	s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)

	for s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_STATIC},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FIELD},
	}) {
		s.ClassVarDec()
	}

	for s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CONSTRUCTOR},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_METHOD},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FUNCTION},
	}) {
		s.Subroutine()
	}

	s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACE)
}

// Compiles a static variable declaration or a field declaration
func (s *parser) ClassVarDec() {
	var ft FieldType
	var typing string
	var varName string
	hasError := false
	token, err := s.process([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_STATIC},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FIELD},
	})

	if err == nil {
		ft = constructorTTtoFT[token.Subtype]
	} else {
		hasError = true
	}

	token, err = s.helper_type([]tokenpair{})

	if err == nil {
		typing = token.Lexeme
	} else {
		hasError = true
	}

	token, err = s.identifierHelper()

	if err == nil {
		varName = token.Lexeme
	} else {
		hasError = true
	}

	if !hasError {
		s.classSt.Define(Name(varName), typing, ft)
	}

	// (',' varName)
	for s.matches([]tokenpair{{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA}}) {
		s.symbolHelper(jack_tokenizer.SYM_COMMA)
		token, err = s.identifierHelper()

		if err == nil {
			varName = token.Lexeme
			s.classSt.Define(Name(varName), typing, ft)
		}
	}

	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)
}

// Compiles a complete method, function or constructor
func (s *parser) Subroutine() {
	s.subroutineSt.Reset()

	keyToken, err := s.process([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_CONSTRUCTOR},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_FUNCTION},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_METHOD},
	})

	if err == nil && keyToken.Subtype == jack_tokenizer.KW_METHOD {
		s.subroutineSt.Define("this", s.className, ARG)
	}

	s.helper_type([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_VOID},
	})

	token, _ := s.identifierHelper()
	s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)

	nargs := s.Parameters()

	s.vmWriter.WriteFunction(Name(fmt.Sprintf("%s.%s", s.className, token.Lexeme)), nargs)

	if keyToken.Subtype == jack_tokenizer.KW_METHOD {
		s.vmWriter.WritePush(ARGUMENT, 0)
		s.vmWriter.WritePop(POINTER, 0)
	} else if keyToken.Subtype == jack_tokenizer.KW_CONSTRUCTOR {
		n := s.subroutineSt.VarCount(FIELD)
		s.vmWriter.WritePush(CONSTANT, n)
		s.vmWriter.WriteCall("Memory.alloc", 1)
		s.vmWriter.WritePop(POINTER, 0)
	}

	s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)

	s.SubroutineBody()
}

// Compiles a (possibly empty) Parameters
// list. Does not handle the enclosing
// parantheses tokens (ands).
func (s *parser) Parameters() int {
	count := 0
	processTypeVarName := func() {
		var typing string
		var varName string
		token, err := s.process(typePair)
		if err == nil {
			typing = token.Lexeme
		}
		token, err = s.identifierHelper()

		if err == nil {
			varName = token.Lexeme
		}

		s.subroutineSt.Define(Name(varName), typing, ARG)
		count++
	}
	if s.matches(typePair) {
		processTypeVarName()

		for s.matches([]tokenpair{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
		}) {
			s.symbolHelper(jack_tokenizer.SYM_COMMA)
			processTypeVarName()
		}
	}

	return count
}

// Compiles a subroutine's body
func (s *parser) SubroutineBody() {
	s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)

	for s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_VAR},
	}) {
		s.VarDec()
	}
	s.Statements()
}

// Compiles a var declaration
func (s *parser) VarDec() {
	s.keywordHelper(jack_tokenizer.KW_VAR)
	var typing string
	var varName string

	token, err := s.helper_type(nil)
	if err == nil {
		typing = token.Lexeme
	}

	token, err = s.identifierHelper()

	if err == nil {
		varName = token.Lexeme
	}

	s.subroutineSt.Define(Name(varName), typing, FieldType(VAR))
	for s.matches([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
	}) {
		s.symbolHelper(jack_tokenizer.SYM_COMMA)
		s.identifierHelper()
	}

	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)
}

// Compiles a sequeneces of statemnents
// Does not handle the enclosing curly
// bracket tokens { and }.
func (s *parser) Statements() {
	states := []tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_LET},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_IF},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_WHILE},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_DO},
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_RETURN},
	}

	if s.matches(states) {
		for s.matches(states) {
			switch {
			case s.matches([]tokenpair{states[0]}):
				s.LetStatement()
			case s.matches([]tokenpair{states[1]}):
				s.IfStatement()
			case s.matches([]tokenpair{states[2]}):
				s.While()
			case s.matches([]tokenpair{states[3]}):
				s.Do()
			case s.matches([]tokenpair{states[4]}):
				s.ReturnStatement()
			}
		}
	} else {
		s.err = errors.New("unexpected lexeme " + s.Current().Lexeme)
	}
}

// Compiles a let statement.
func (s *parser) LetStatement() {
	s.keywordHelper(jack_tokenizer.KW_LET)
	token, _ := s.identifierHelper()
	res := s.resolveSymbol(token.Lexeme)
	isarr := false
	if s.matches([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_BRACK},
	}) {
		isarr = true
		s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACK)

		s.vmWriter.WritePush(fieldtoSegment[res.symbol], res.index)
		s.Expression()
		s.vmWriter.WriteArithmetic(ADD)
		s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACK)
	}

	s.symbolHelper(jack_tokenizer.SYM_EQUALS)

	s.Expression()
	if isarr {
		s.vmWriter.WritePop(TEMP, 0)
		s.vmWriter.WritePop(POINTER, 1)
		s.vmWriter.WritePush(TEMP, 0)
		s.vmWriter.WritePop(THAT, 0)
	} else {
		// pop symbolArgName index

		s.vmWriter.WritePop(fieldtoSegment[res.symbol], res.index)
	}
	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)
}

// Compiles an if statement
// possibly with a trailing else clause
func (s *parser) IfStatement() {
	s.keywordHelper(jack_tokenizer.KW_IF)

	s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)

	s.Expression()
	// not
	s.vmWriter.WriteArithmetic(NOT)
	// if-goto label1
	s.vmWriter.WriteIf(fmt.Sprintf("%s.IF-%d", s.className, s.labelNumber))
	s.labelNumber++
	s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)

	s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)

	s.Statements()
	// goto label2
	s.vmWriter.WriteGoto(fmt.Sprintf("%s.IF-%d", s.className, s.labelNumber))
	s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACE)

	if s.matches([]tokenpair{
		{jack_tokenizer.KEYWORD, jack_tokenizer.KW_ELSE},
	}) {
		// label l1
		s.vmWriter.WriteLabel(fmt.Sprintf("%s.IF-%d", s.className, s.labelNumber-1))
		s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)

		s.Statements()

		s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACE)
	}
	// label l2
	s.vmWriter.WriteLabel(fmt.Sprintf("%s.IF-%d", s.className, s.labelNumber))
}

// Compiles a While statement
func (s *parser) While() {
	s.keywordHelper(jack_tokenizer.KW_WHILE)
	s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)
	s.vmWriter.WriteLabel(fmt.Sprintf("%s-IF-%d", s.className, s.labelNumber))
	s.Expression()
	// not
	s.vmWriter.WriteArithmetic(NOT)
	s.labelNumber++
	// if-goto l2
	s.vmWriter.WriteIf(fmt.Sprintf("%s-IF-%d", s.className, s.labelNumber))
	s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)
	s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACE)
	s.Statements()
	// goto l1
	s.vmWriter.WriteGoto(fmt.Sprintf("%s-IF-%d", s.className, s.labelNumber-1))
	s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACE)
	// label l2
	s.vmWriter.WriteLabel(fmt.Sprintf("%s-IF-%d", s.className, s.labelNumber))
}

// Compiles a Do statement
func (s *parser) Do() {
	s.keywordHelper(jack_tokenizer.KW_DO)
	s.Expression()
	// pop something 0
	s.vmWriter.WritePop(TEMP, 0)
	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)
}

// Compiles a return statement
func (s *parser) ReturnStatement() {
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

	s.keywordHelper(jack_tokenizer.KW_RETURN)
	if s.matches(begTerm) {
		s.Expression()
	}
	// return
	s.vmWriter.WriteReturn()
	s.symbolHelper(jack_tokenizer.SYM_SEMICOLON)
}

// Compiles an Expression
func (s *parser) Expression() {
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
	s.Term()

	for s.matches(op) {
		token, _ := s.process(op)
		s.Term()

		switch token.Subtype {
		case jack_tokenizer.SYM_SLASH:
			s.vmWriter.WriteCall("Math.divide", 2)
		case jack_tokenizer.SYM_ASTERISK:
			s.vmWriter.WriteCall("Math.multiply", 2)
		default:
			s.vmWriter.WriteArithmetic(subtypeToOp[token.Subtype])
		}

	}
}

// Compiles a Term. If the current token is an
// identifier, the routine must resolve it
// into a variable, an array element, or a
// subroutine call. A single lookahead tokezn,
// which may be [, (, or ., suffices to distinguish
// between the possibilities.
// Any other token is not part of this Term
// and should not be advanced over.
func (s *parser) Term() {
	switch s.Current().Tokentype {
	case jack_tokenizer.IDENTIFIER:
		// variable, array element or subroutine
		peek, err := s.peek()
		if err != nil {
			s.err = nil
		} else {
			switch peek.Subtype {
			case jack_tokenizer.SYM_LEFT_PAREN:
			case jack_tokenizer.SYM_PERIOD:
				s.SubroutineCall()
			case jack_tokenizer.SYM_LEFT_BRACK:
				// varname[expression]
				token, _ := s.identifierHelper()
				resolved := s.resolveSymbol(token.Lexeme)
				s.vmWriter.WritePush(fieldtoSegment[resolved.symbol], resolved.index)
				s.symbolHelper(jack_tokenizer.SYM_LEFT_BRACK)
				s.Expression()
				s.vmWriter.WriteArithmetic(ADD)
				s.symbolHelper(jack_tokenizer.SYM_RIGHT_BRACK)
				s.vmWriter.WritePop(POINTER, 1)
				s.vmWriter.WritePush(THAT, 0)
			default:
				token, _ := s.identifierHelper()
				resolved := s.resolveSymbol(token.Lexeme)

				s.vmWriter.WritePush(fieldtoSegment[resolved.symbol], resolved.index)
			}

		}
	case jack_tokenizer.INT_CONSTANT:
		token, _ := s.process([]tokenpair{
			{jack_tokenizer.INT_CONSTANT, jack_tokenizer.NONE},
		})
		i, _ := strconv.Atoi(token.Lexeme)
		s.vmWriter.WritePush(CONSTANT, i)
	case jack_tokenizer.STRING_CONSTANT:
		token, _ := s.process([]tokenpair{
			{jack_tokenizer.STRING_CONSTANT, jack_tokenizer.NONE},
		})
		s.vmWriter.WritePush(CONSTANT, len(token.Lexeme))
		s.vmWriter.WriteCall("String.new", 1)
		for _, c := range []byte(token.Lexeme) {
			s.vmWriter.WritePush(CONSTANT, int(c))
			s.vmWriter.WriteCall("String.appendChar", 1)
		}
		// push c
	case jack_tokenizer.SYMBOL:
		switch s.Current().Subtype {
		case jack_tokenizer.SYM_LEFT_PAREN:
			s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)
			s.Expression()
			s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)
		case jack_tokenizer.SYM_MINUS:
		case jack_tokenizer.SYM_TILDE:
			s.symbolHelper(s.Current().Subtype)
			s.Term()
			// output op
			tokenName := SUB
			if s.Current().Subtype == jack_tokenizer.SYM_TILDE {
				tokenName = NOT
			}
			s.vmWriter.WriteArithmetic(tokenName)
		}
	case jack_tokenizer.KEYWORD:
		switch s.Current().Subtype {
		case jack_tokenizer.KW_FALSE:
		case jack_tokenizer.KW_NULL:
			s.vmWriter.WritePush(CONSTANT, 0)
		case jack_tokenizer.KW_TRUE:
			s.vmWriter.WritePush(CONSTANT, 1)
			s.vmWriter.WriteArithmetic(NEG)
		case jack_tokenizer.KW_THIS:
			s.vmWriter.WritePush(POINTER, 0)
		}
		s.keywordHelper(s.Current().Subtype)
	}

}

func (s *parser) SubroutineCall() {
	mainToken, _ := s.identifierHelper() // class's name or subroutine name, depending on if theres a .
	mainName := mainToken.Lexeme
	secondaryName := ""

	t := s.Current()
	s.process([]tokenpair{
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_LEFT_PAREN},
		{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_PERIOD},
	})

	switch t.Subtype {
	case jack_tokenizer.SYM_PERIOD:
		fn, _ := s.identifierHelper()
		secondaryName = "." + fn.Lexeme
		s.symbolHelper(jack_tokenizer.SYM_LEFT_PAREN)

		fallthrough
	case jack_tokenizer.SYM_LEFT_PAREN:
		n := s.ExpressionList()
		s.symbolHelper(jack_tokenizer.SYM_RIGHT_PAREN)
		s.vmWriter.WriteCall(fmt.Sprintf("%s%s", mainName, secondaryName), n)
	}
}

// Compiles a (possibly empty) comma-
// separated list of expression. Returns
// the number of expressions in the list
func (s *parser) ExpressionList() int {
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
	if s.matches(begTerm) {
		count++
		s.Expression()
		for s.matches([]tokenpair{
			{jack_tokenizer.SYMBOL, jack_tokenizer.SYM_COMMA},
		}) {
			count++
			s.symbolHelper(jack_tokenizer.SYM_COMMA)
			s.Expression()
		}
	}

	return count
}

// x* : 0 or more
// ? one or more
// x y x followed by y
// x | y x or y

func ParseGrammar(tokens []jack_tokenizer.Token) func(io.WriteCloser) error {
	return func(w io.WriteCloser) error {
		parser := NewParser(tokens, *NewVMWriter(w))

		err := parser.Parse()
		return err
	}
}
