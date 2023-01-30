// This package compiles AST into a mysterious stack-based IR
// so that we don't end up calling tons of CGO libjit functions.
package ir

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/yesh0/gruel/pkg/gruelparser"
)

// Builds byte code
type IrBuilder struct {
	b bytes.Buffer
}

type Operator struct {
	Opcode uint64
	Argc   int
	Argt   gruelparser.TokenType
}

// Maps operators to opcodes
var Operators = map[string]Operator{
	"+": {0x0001, 2, gruelparser.TypeInt},
	"-": {0x0002, 2, gruelparser.TypeInt},
	"*": {0x0003, 2, gruelparser.TypeInt},
	"/": {0x0004, 2, gruelparser.TypeInt},
	"%": {0x0005, 2, gruelparser.TypeInt},
}

func (b *IrBuilder) Push(value string, t gruelparser.TokenType) error {
	var output uint64
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
		return fmt.Errorf("symbol unsupported")
	case gruelparser.TypeParenthesis:
		opcode, ok := Operators[value]
		if ok {
			output = opcode.Opcode
		} else {
			return fmt.Errorf("operator %s not found", value)
		}
	}
	tType := uint64(t)
	binary.Write(&b.b, binary.LittleEndian, &tType)
	binary.Write(&b.b, binary.LittleEndian, &output)
	return nil
}

func (b *IrBuilder) Code() []byte {
	return b.b.Bytes()
}

func (b *IrBuilder) Append(ast *gruelparser.GruelAstNode) error {
	if ast.Parameters != nil {
		for _, node := range ast.Parameters {
			if err := b.Append(&node); err != nil {
				return err
			}
		}
	}
	return b.Push(ast.Value, ast.Type)
}

// Compiles the AST into byte codes
func Compile(ast *gruelparser.GruelAstNode) ([]byte, error) {
	var b IrBuilder
	if err := b.Append(ast); err != nil {
		return nil, err
	}
	return b.Code(), nil
}
