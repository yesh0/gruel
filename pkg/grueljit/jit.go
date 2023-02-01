package grueljit

/*

#cgo CFLAGS:  -I../../libjit/include
#cgo LDFLAGS: -L../../libjit -ljit -lm
#include "gruel_jit.h"

*/
import "C"
import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/yesh0/gruel/internal/caller"
	"github.com/yesh0/gruel/internal/gruelparser"
	"github.com/yesh0/gruel/internal/ir"
)

//go:generate go run ../../build/ir/operators.go gruel_jit.c

const (
	TypeBool   byte = byte(gruelparser.TypeBool)
	TypeInt    byte = byte(gruelparser.TypeInt)
	TypeFloat  byte = byte(gruelparser.TypeFloat)
	TypeString byte = byte(gruelparser.TypeString)
)

type Function struct {
	function   uint64
	arg_types  []byte
	arg_map    map[string]int
	max_stack  int
	stringc    int
	float      bool
	references any
}

func Compile(code string, symbols map[string]byte) (*Function, error) {
	ast, err := gruelparser.Parse(code)
	if err != nil {
		return nil, err
	}
	builder, err := ir.Compile(&ast, symbols)
	if err != nil {
		return nil, err
	}
	f, err := compileOpcodes(builder)
	if f != nil {
		runtime.SetFinalizer(f, free)
	}
	return f, err
}

func free(f *Function) {
	f.Free()
}

// Compiles the byte code and returns a function handle.
func compileOpcodes(ir *ir.IrBuilder) (*Function, error) {
	code := ir.Code()
	args := ir.Args()
	var args_ptr *C.char
	if len(args) != 0 {
		args_ptr = (*C.char)(unsafe.Pointer(&args[0]))
	}

	handle := uint64(C.compile_opcodes(
		(C.long)(len(code)/8),
		(*C.long)(unsafe.Pointer(&code[0])),
		(C.long)(len(args)),
		args_ptr,
	))

	if handle == 0 {
		return nil, fmt.Errorf("unexpected error when passing to libjit")
	}

	runtime.KeepAlive(args)
	runtime.KeepAlive(code)
	var float bool
	if code[0] == 0xff {
		float = true
	}

	return &Function{
		function: handle, arg_map: ir.ArgMap(),
		float: float, references: ir.References(),
		arg_types: ir.Args(),
		stringc:   ir.StringArgc(),
		max_stack: ir.MaxStack() + 256,
	}, nil
}

// Frees the resources.
func (f *Function) Free() {
	C.free_function((C.long)(f.function))
	f.function = 0
}

func (f *Function) convertResult(v uint64, err error) (any, error) {
	if f.Float() {
		return math.Float64frombits(v), err
	} else {
		return v, err
	}
}

func (f *Function) Call(args map[string]any) (any, error) {
	argc := len(f.arg_map)
	if argc == 0 {
		return f.convertResult(f.CallRaw(nil))
	}

	if args == nil {
		return nil, fmt.Errorf("requires parameters")
	}

	params := make([]uint64, argc+2*f.stringc)
	strings := params[argc:]
	for name, index := range f.arg_map {
		value, ok := args[name]
		if !ok {
			return nil, fmt.Errorf("parameter %s not found", name)
		}
		target := f.arg_types[index]
		if v, ok := value.(string); ok {
			if target == TypeString {
				hdr := (*reflect.StringHeader)(unsafe.Pointer(&v))
				strings[0] = uint64(hdr.Data)
				strings[1] = uint64(hdr.Len)
				params[index] = uint64(uintptr(unsafe.Pointer(&strings[0])))
				strings = strings[2:]
			} else {
				return nil, fmt.Errorf("unsupported conversion from string")
			}
		} else {
			if target == TypeString {
				return nil, fmt.Errorf("unsupported conversion into string")
			} else {
				converted, err := convertType(value, target)
				if err == nil {
					params[index] = converted
				} else {
					return nil, err
				}
			}
		}
	}
	return f.convertResult(f.CallRaw(params))
}

func convertType(param any, target byte) (uint64, error) {
	var out uint64
	var realType = TypeInt
	switch v := param.(type) {
	case bool:
		if v {
			out = 1
		}
	case uint:
		out = uint64(v)
	case uint8:
		out = uint64(v)
	case uint16:
		out = uint64(v)
	case uint32:
		out = uint64(v)
	case uint64:
		out = uint64(v)
	case int:
		out = uint64(v)
	case int8:
		out = uint64(v)
	case int16:
		out = uint64(v)
	case int32:
		out = uint64(v)
	case int64:
		out = uint64(v)
	case float32:
		out = math.Float64bits(float64(v))
		realType = TypeFloat
	case float64:
		out = math.Float64bits(v)
		realType = TypeFloat
	default:
		return 0, fmt.Errorf("unsupported type")
	}
	switch {
	case target == TypeFloat:
		if realType != TypeFloat {
			out = math.Float64bits(float64(out))
		}
	case target == TypeBool:
		if out != 0 {
			out = 1
		}
	case target == TypeInt:
		if realType == TypeFloat {
			out = uint64(math.Float64frombits(out))
		}
	}
	return out, nil
}

// Calls the function
func (f *Function) CallRaw(params []uint64) (uint64, error) {
	argc := len(f.arg_map)
	if params != nil {
		if len(params) < argc {
			return 0, fmt.Errorf("arguments not enough")
		}
	} else if argc != 0 {
		return 0, fmt.Errorf("no arguments provided")
	}

	ret := caller.CallJit(
		f.function,
		params,
		uint64(f.max_stack),
	)
	runtime.KeepAlive(params)
	return ret, nil
}

func (f *Function) Float() bool {
	return f.float
}

// Returns false if the code is interpreted
// (which may very likely overflow the stack).
func IsJit() bool {
	return C.is_jit_supported() != 0
}
