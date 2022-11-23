package jack_symboltable

type Name string

type tableentry struct {
	index  int
	typing string
}

type SegmentType int

const (
	STATIC = iota
	FIELD
	ARG
	VAR

	NONE
)

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

func (s *SymbolTable) getTable(kind SegmentType) (*map[Name]tableentry, *int) {
	switch kind {
	case STATIC:
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
	sym.staticIndex = 0
	sym.fieldIndex = 0
	sym.argIndex = 0
	sym.varIndex = 0
}

func (sym *SymbolTable) Define(n Name, t string, kind SegmentType) {
	table, index := sym.getTable(kind)

	(*table)[n] = tableentry{*index, t}
}

func (sym *SymbolTable) VarCount(kind SegmentType) int {
	table, _ := sym.getTable(kind)

	return len(*table)
}

func (sym *SymbolTable) KindOf(n Name) SegmentType {
	if _, ok := sym.staticTable[n]; ok {
		return STATIC
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
