// This package compiles AST into a mysterious stack-based IR
// so that we don't end up calling tons of CGO libjit functions.
package ir

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/yesh0/gruel/internal/gruelparser"
)

type GoString [2]uint64

// Builds byte code
type IrBuilder struct {
	b     bytes.Buffer
	final bool
	argc  int
	// Maps from paramter names to parameter indices
	argv map[string]int
	// Paramter types
	args    []byte
	symbols map[string]byte
	// Keep those objects alive
	objects []string
	strings list.List
	// Stack space needed, in bytes
	maxStack     int
	currentStack int
}

func (b *IrBuilder) Push(value string, t gruelparser.TokenType, argc int) error {
	if b.final {
		return fmt.Errorf("code already finalized")
	}

	b.currentStack += 8
	if b.maxStack < b.currentStack {
		b.maxStack = b.currentStack
	}

	var output uint64
	tType := uint64(t)
	switch t {
	case gruelparser.TypeBool:
		if value == "true" {
			output = 1
		} else {
			output = 0
		}
	case gruelparser.TypeFloat:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		output = math.Float64bits(v)
	case gruelparser.TypeInt:
		v, err := strconv.ParseInt(value, 0, 64)
		if rangeErr := err; rangeErr != nil {
			output, err = strconv.ParseUint(value, 0, 64)
			if err != nil {
				return rangeErr
			}
		} else {
			output = uint64(v)
		}
	case gruelparser.TypeString:
		length := len(value)
		if length >= math.MaxInt32-2 {
			return fmt.Errorf("string too large")
		}
		b.objects = append(b.objects, value)
		hdr := (*reflect.StringHeader)(unsafe.Pointer(&value))
		s := &GoString{uint64(hdr.Data), uint64(length)}
		b.strings.PushBack(s)
		output = uint64(uintptr(unsafe.Pointer(&s[0])))
	case gruelparser.TypeSymbol:
		var ok bool
		_, ok = b.symbols[value]
		if ok {
			index, ok := b.argv[value]
			if ok {
				output = uint64(index)
			} else {
				b.argv[value] = b.argc
				output = uint64(b.argc)
				b.argc++
			}
		} else {
			return fmt.Errorf("symbol %s not found", value)
		}
	case gruelparser.TypeParenthesis:
		if op := findOperator(value, argc); op == nil {
			return fmt.Errorf("operator %s not found", value)
		} else {
			output = uint64(op.Opcode)
			// Handling `(+ 1 2 3 4 5 ...)`
			for argc > op.Argc {
				binary.Write(&b.b, binary.LittleEndian, &tType)
				binary.Write(&b.b, binary.LittleEndian, &output)
				b.currentStack -= 8
				argc--
			}
			b.currentStack -= 16
		}
	}
	binary.Write(&b.b, binary.LittleEndian, &tType)
	binary.Write(&b.b, binary.LittleEndian, &output)
	return nil
}

func findOperator(name string, argc int) *Operator {
	ops, ok := Operators[name]
	if ok {
		var bi_op *Operator
		for _, op := range ops {
			if op.Argc == argc {
				return &op
			}
			if op.Argc == 2 {
				bi_op = &op
			}
		}
		return bi_op
	} else {
		return nil
	}
}

func (b *IrBuilder) Finalize() {
	if !b.final {
		b.final = true

		args := make([]byte, len(b.argv))
		for name, index := range b.argv {
			args[index] = byte(b.symbols[name])
		}
		b.args = args
	}
}

// Returns the byte code
func (b *IrBuilder) Code() []byte {
	b.Finalize()
	return b.b.Bytes()
}

// Encodes the argument types, one per byte
func (b *IrBuilder) Args() []byte {
	b.Finalize()
	return b.args
}

// Maps named parameters to indices
func (b *IrBuilder) ArgMap() map[string]int {
	b.Finalize()
	return b.argv
}

// Objects used by the program.
//
// It's fortunate that Go's GC does not move objects.
func (b *IrBuilder) References() any {
	b.Finalize()
	if len(b.objects) == 0 && b.strings.Len() == 0 {
		return nil
	}
	return []any{b.objects, b.strings}
}

// Needed stack space, in bytes
func (b *IrBuilder) MaxStack() int {
	b.Finalize()
	return b.maxStack
}

func (b *IrBuilder) StringArgc() int {
	b.Finalize()
	count := 0
	for _, v := range b.args {
		if v == byte(gruelparser.TypeString) {
			count++
		}
	}
	return count
}

func (b *IrBuilder) Append(ast *gruelparser.GruelAstNode) error {
	if ast.Parameters != nil {
		for i := len(ast.Parameters) - 1; i >= 0; i-- {
			if err := b.Append(&ast.Parameters[i]); err != nil {
				return err
			}
		}
	}
	return b.Push(ast.Value, ast.Type, len(ast.Parameters))
}

type CompiledChunk struct {
	Code       []byte
	Parameters []gruelparser.TokenType
}

// Compiles the AST into byte codes
func Compile(ast *gruelparser.GruelAstNode, symbols map[string]byte) (*IrBuilder, error) {
	for k, vb := range symbols {
		v := gruelparser.TokenType(vb)
		if v != gruelparser.TypeBool && v != gruelparser.TypeInt &&
			v != gruelparser.TypeFloat && v != gruelparser.TypeString {
			return nil, fmt.Errorf("symbol %s must have a value type", k)
		}
	}

	b := IrBuilder{
		symbols: symbols,
		argv:    make(map[string]int, len(symbols)),
	}
	if err := b.Append(ast); err != nil {
		return nil, err
	}
	return &b, nil
}
