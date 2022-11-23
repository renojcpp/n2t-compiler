package jack_tokenizer

type Token struct {
	Lexeme    string
	Line      int
	Tokentype TokenType
	Subtype   TokenSubtype
}

func NewToken(lexeme string, line int, tokentype TokenType, subtype TokenSubtype) Token {
	return Token{
		lexeme,
		line,
		tokentype,
		subtype,
	}
}

type TokenType int

const (
	ERROR = -1
)

//go:generate string -type=TokenType
const (
	KEYWORD TokenType = iota
	SYMBOL
	INT_CONSTANT
	STRING_CONSTANT
	IDENTIFIER
)

type TokenSubtype int

//go:generate stringer -type=TokenSubtype
const (
	UNKNOWN TokenSubtype = -1
	NONE    TokenSubtype = iota
	// Keywords
	KW_CLASS
	KW_CONSTRUCTOR
	KW_FUNCTION
	KW_METHOD
	KW_FIELD
	KW_STATIC
	KW_VAR
	KW_INT
	KW_CHAR
	KW_BOOLEAN
	KW_VOID
	KW_TRUE
	KW_FALSE
	KW_NULL
	KW_THIS
	KW_LET
	KW_DO
	KW_IF
	KW_ELSE
	KW_WHILE
	KW_RETURN

	// symbols
	SYM_LEFT_BRACE
	SYM_RIGHT_BRACE
	SYM_LEFT_PAREN
	SYM_RIGHT_PAREN
	SYM_LEFT_BRACK
	SYM_RIGHT_BRACK
	SYM_PERIOD
	SYM_COMMA
	SYM_SEMICOLON
	SYM_PLUS
	SYM_MINUS
	SYM_ASTERISK
	SYM_SLASH
	SYM_AMPERSAND
	SYM_PIPE
	SYM_LESS_THAN
	SYM_GREATER_THAN
	SYM_EQUALS
	SYM_TILDE
)

type tokenpair struct {
	tt TokenType
	st TokenSubtype
}

type luxmap map[string]tokenpair

var mp = luxmap{
	"class":       {KEYWORD, KW_CLASS},
	"constructor": {KEYWORD, KW_CONSTRUCTOR},
	"function":    {KEYWORD, KW_FUNCTION},
	"method":      {KEYWORD, KW_METHOD},
	"field":       {KEYWORD, KW_FIELD},
	"static":      {KEYWORD, KW_STATIC},
	"var":         {KEYWORD, KW_VAR},
	"int":         {KEYWORD, KW_INT},
	"char":        {KEYWORD, KW_CHAR},
	"boolean":     {KEYWORD, KW_BOOLEAN},
	"void":        {KEYWORD, KW_VOID},
	"true":        {KEYWORD, KW_TRUE},
	"false":       {KEYWORD, KW_FALSE},
	"null":        {KEYWORD, KW_NULL},
	"this":        {KEYWORD, KW_THIS},
	"let":         {KEYWORD, KW_LET},
	"do":          {KEYWORD, KW_DO},
	"if":          {KEYWORD, KW_IF},
	"else":        {KEYWORD, KW_ELSE},
	"while":       {KEYWORD, KW_WHILE},
	"return":      {KEYWORD, KW_RETURN},

	"{": {SYMBOL, SYM_LEFT_BRACE},
	"}": {SYMBOL, SYM_RIGHT_BRACE},
	"(": {SYMBOL, SYM_LEFT_PAREN},
	")": {SYMBOL, SYM_RIGHT_PAREN},
	"[": {SYMBOL, SYM_LEFT_BRACK},
	"]": {SYMBOL, SYM_RIGHT_BRACK},
	".": {SYMBOL, SYM_PERIOD},
	",": {SYMBOL, SYM_COMMA},
	";": {SYMBOL, SYM_SEMICOLON},
	"+": {SYMBOL, SYM_PLUS},
	"-": {SYMBOL, SYM_MINUS},
	"*": {SYMBOL, SYM_ASTERISK},
	"/": {SYMBOL, SYM_SLASH},
	"&": {SYMBOL, SYM_AMPERSAND},
	"|": {SYMBOL, SYM_PIPE},
	"<": {SYMBOL, SYM_LESS_THAN},
	">": {SYMBOL, SYM_GREATER_THAN},
	"=": {SYMBOL, SYM_EQUALS},
	"~": {SYMBOL, SYM_TILDE},
}
