package jack_compiler

import jack_tokenizer "github.com/renojcpp/n2t-compiler/tokenizer"

type Name string

type tableentry struct {
	index  int
	typing string
}

type FieldType int

const (
	STATIC_F FieldType = iota
	FIELD
	ARG
	VAR

	NONE
)

var fieldtoSegment = map[FieldType]SegmentType{
	STATIC_F: STATIC_S,
	FIELD:    THIS,
	VAR:      LOCAL,
	ARG:      ARGUMENT,
}

var constructorTTtoFT = map[jack_tokenizer.TokenSubtype]FieldType{
	jack_tokenizer.KW_STATIC: STATIC_F,
	jack_tokenizer.KW_FIELD:  VAR,
}

var subtypeToOp = map[jack_tokenizer.TokenSubtype]ArithmeticType{
	jack_tokenizer.SYM_PLUS:         ADD,
	jack_tokenizer.SYM_MINUS:        SUB,
	jack_tokenizer.SYM_AMPERSAND:    AND,
	jack_tokenizer.SYM_LESS_THAN:    LT,
	jack_tokenizer.SYM_GREATER_THAN: GT,
	jack_tokenizer.SYM_EQUALS:       EQ,
	jack_tokenizer.SYM_PIPE:         OR,
	jack_tokenizer.SYM_TILDE:        NOT,
}

type SymbolTable struct {
	staticTable map[Name]tableentry
	fieldTable  map[Name]tableentry
	argTable    map[Name]tableentry
	varTable    map[Name]tableentry

	staticIndex int
	fieldIndex  int
	argIndex    int
	varIndex    int
}

func (s *SymbolTable) getTable(kind FieldType) (*map[Name]tableentry, *int) {
	switch kind {
	case STATIC_F:
		return &s.staticTable, &s.staticIndex
	case FIELD:
		return &s.fieldTable, &s.fieldIndex
	case ARG:
		return &s.argTable, &s.argIndex
	case VAR:
		return &s.varTable, &s.varIndex
	default:
		panic("unknown segment")
	}
}

func (sym *SymbolTable) find(n Name) *map[Name]tableentry {
	if _, ok := sym.staticTable[n]; ok {
		return &sym.staticTable
	}

	if _, ok := sym.fieldTable[n]; ok {
		return &sym.fieldTable
	}

	if _, ok := sym.argTable[n]; ok {
		return &sym.argTable
	}

	if _, ok := sym.varTable[n]; ok {
		return &sym.varTable
	}

	return nil
}

func (sym *SymbolTable) Reset() {
	sym.staticTable = make(map[Name]tableentry)
	sym.fieldTable = make(map[Name]tableentry)
	sym.argTable = make(map[Name]tableentry)
	sym.varTable = make(map[Name]tableentry)

	sym.staticIndex = 0
	sym.fieldIndex = 0
	sym.argIndex = 0
	sym.varIndex = 0
}

func (sym *SymbolTable) Define(n Name, t string, kind FieldType) {
	table, index := sym.getTable(kind)

	(*table)[n] = tableentry{*index, t}
	*index += 1
}

func (sym *SymbolTable) VarCount(kind FieldType) int {
	table, _ := sym.getTable(kind)

	return len(*table)
}

func (sym *SymbolTable) KindOf(n Name) FieldType {
	if _, ok := sym.staticTable[n]; ok {
		return STATIC_F
	}

	if _, ok := sym.fieldTable[n]; ok {
		return FIELD
	}

	if _, ok := sym.argTable[n]; ok {
		return ARG
	}

	if _, ok := sym.varTable[n]; ok {
		return VAR
	}

	return NONE
}

func (sym *SymbolTable) TypeOf(n Name) string {
	table := sym.find(n)

	return (*table)[n].typing
}

func (sym *SymbolTable) IndexOf(n Name) int {
	table := sym.find(n)

	return (*table)[n].index
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{}
}

type SegmentType int

const (
	CONSTANT SegmentType = iota
	ARGUMENT
	LOCAL
	STATIC_S
	THIS
	THAT
	POINTER
	TEMP
)

type ArithmeticType int

const (
	ADD ArithmeticType = iota
	SUB
	NEG
	EQ
	GT
	LT
	AND
	OR
	NOT
)
