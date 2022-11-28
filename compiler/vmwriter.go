package jack_compiler

import (
	"fmt"
	"io"
)

type VMWriter struct {
	io.WriteCloser
}

var sm = map[SegmentType]string{
	CONSTANT: "constant",
	ARGUMENT: "arg",
	LOCAL:    "local",
	STATIC_S: "static",
	THIS:     "this",
	THAT:     "that",
	POINTER:  "pointer",
	TEMP:     "temp",
}

var am = map[ArithmeticType]string{
	ADD: "add",
	SUB: "sub",
	NEG: "neg",
	EQ:  "eq",
	GT:  "gt",
	LT:  "lt",
	AND: "and",
	OR:  "or",
	NOT: "not",
}

func (s *VMWriter) WritePush(segment SegmentType, index int) {
	io.WriteString(s, fmt.Sprintf("push %s %d\n", sm[segment], index))
}

func (s *VMWriter) WritePop(segment SegmentType, index int) {
	io.WriteString(s, fmt.Sprintf("pop %s %d\n", sm[segment], index))
}

func (s *VMWriter) WriteArithmetic(arithmetic ArithmeticType) {
	io.WriteString(s, am[arithmetic]+"\n")
}

func (s *VMWriter) WriteLabel(label string) {
	io.WriteString(s, fmt.Sprintf("label %s\n", label))
}

func (s *VMWriter) WriteGoto(label string) {
	io.WriteString(s, fmt.Sprintf("goto %s\n", label))
}

func (s *VMWriter) WriteIf(label string) {
	io.WriteString(s, fmt.Sprintf("if-goto %s\n", label))
}

func (s *VMWriter) WriteCall(name string, nArgs int) {
	io.WriteString(s, fmt.Sprintf("call %s %d\n", name, nArgs))
}

func (s *VMWriter) WriteFunction(name Name, nvars int) {
	io.WriteString(s, fmt.Sprintf("function %s %s\n", name, nvars))
}

func (s *VMWriter) WriteReturn() {
	io.WriteString(s, "return\n")
}

func NewVMWriter(w io.WriteCloser) *VMWriter {
	return &VMWriter{w}
}
