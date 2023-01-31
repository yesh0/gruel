// This package compiles AST into a mysterious stack-based IR
// so that we don't end up calling tons of CGO libjit functions.
package ir

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/yesh0/gruel/internal/gruelparser"
)

// Builds byte code
type IrBuilder struct {
	b       bytes.Buffer
	final   bool
	argc    int
	argv    map[string]int
	args    []byte
	symbols map[string]byte
}

func (b *IrBuilder) Push(value string, t gruelparser.TokenType, argc int) error {
	if b.final {
		return fmt.Errorf("code already finalized")
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
		return fmt.Errorf("string unsupported")
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
				argc--
			}
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
		if v != gruelparser.TypeBool && v != gruelparser.TypeInt && v != gruelparser.TypeFloat {
			return nil, fmt.Errorf("symbol %s must have a value type", k)
		}
	}

	b := IrBuilder{symbols: symbols, argv: make(map[string]int, len(symbols))}
	if err := b.Append(ast); err != nil {
		return nil, err
	}
	return &b, nil
}
