package jack_tokenizer

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

type scanner struct {
	bytes []byte
	index int
}

func (s *scanner) atEnd() bool {
	return s.index >= len(s.bytes)
}

func (s *scanner) advance() {
	s.index++
}

func (s *scanner) peek() (byte, error) {
	if s.atEnd() || s.index+1 >= len(s.bytes) {
		return 0, errors.New("peek beyond")
	}

	return s.bytes[s.index+1], nil
}

func (s *scanner) current() byte {
	return s.bytes[s.index]
}

func NewScanner(b []byte) scanner {
	return scanner{b, 0}
}

func isNumber(b byte) bool {
	_, err := strconv.ParseInt(string(b), 10, 8)
	return err == nil
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func Tokenize(r io.Reader) ([]Token, error) {
	chars, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var parseError error
	parseError = nil
	scan := NewScanner(chars)
	tokens := make([]Token, 0)
	line := 0

	writeInt := func() (string, error) {
		var ss strings.Builder
		for !scan.atEnd() && isNumber(scan.current()) {
			ss.WriteByte(scan.current())
			scan.advance()
		}

		if !scan.atEnd() && isLetter(scan.current()) {
			return "", errors.New("letter found " + string(scan.current()))
		}

		return ss.String(), nil
	}

	for !scan.atEnd() {
		ch := scan.current()
		pair, ok := mp[string(ch)]
		switch {
		// symbols
		case ch == '/':
			peek, err := scan.peek()
			if err != nil {
				parseError = err
			} else {
				if peek == '/' { // single comment
					for !scan.atEnd() && scan.current() != '\n' {
						scan.advance()
					}
				} else if peek == '*' { // multi line comment
					scan.advance()
					peek, _ = scan.peek()
					for !scan.atEnd() && (scan.current() != '*' || peek != '/') {
						scan.advance()
						peek, _ = scan.peek()
					}
					scan.advance()
				} else {
					tokens = append(tokens, NewToken(string(ch), line, SYMBOL, SYM_SLASH))
				}
			}
			scan.advance()
		case ch == '"':
			var ss strings.Builder
			scan.advance()
			for !scan.atEnd() && (scan.current() != '\n' && scan.current() != '"') {
				ss.WriteByte(scan.current())
				scan.advance()
			}

			sres := ss.String()
			if !scan.atEnd() && scan.current() == '"' {
				tokens = append(tokens, NewToken(sres, line, STRING_CONSTANT, NONE))
				scan.advance()
			} else {
				parseError = errors.New("unterminated string")
				tokens = append(tokens, NewToken(sres, line, ERROR, NONE))
			}
		case ok && pair.tt == SYMBOL:
			tokens = append(tokens, NewToken(string(ch), line, pair.tt, pair.st))
			scan.advance()
		case isNumber(ch):
			tt := INT_CONSTANT
			st := NONE
			numberStr, err := writeInt()
			if err != nil {
				tt = ERROR
				parseError = err
			}
			tokens = append(tokens, NewToken(numberStr, line, TokenType(tt), TokenSubtype(st)))

		case isLetter(ch) || ch == '_': // identifier, or keyword
			var ss strings.Builder
			for !scan.atEnd() && isLetter(scan.current()) {
				ss.WriteByte(scan.current())
				scan.advance()
			}
			sres := ss.String()
			pair, ok = mp[sres]
			if ok { // keyword
				tokens = append(tokens, NewToken(sres, line, pair.tt, pair.st))
			} else { // identifier
				tokens = append(tokens, NewToken(sres, line, IDENTIFIER, NONE))
			}
		default: // whitespace or unrecognized
			scan.advance()
		}

		// scan.advance()
		line++
	}

	return tokens, parseError
}
