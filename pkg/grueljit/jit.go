package grueljit

/*

#cgo CFLAGS:  -I../../libjit/include
#cgo LDFLAGS: -L../../libjit -ljit -lm
#include "gruel_jit.h"

*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/yesh0/gruel/internal/caller"
	"github.com/yesh0/gruel/internal/gruelparser"
	"github.com/yesh0/gruel/internal/ir"
)

//go:generate go run ../../build/ir/operators.go gruel_jit.c

const (
	TypeBool  byte = byte(gruelparser.TypeBool)
	TypeInt   byte = byte(gruelparser.TypeInt)
	TypeFloat byte = byte(gruelparser.TypeFloat)
)

type Function struct {
	function uint64
	arg_map  map[string]int
	float    bool
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

	return &Function{function: handle, arg_map: ir.ArgMap(), float: float}, nil
}

// Frees the resources.
func (f *Function) Free() {
	C.free_function((C.long)(f.function))
	f.function = 0
}

// Calls the function
func (f *Function) Call(args map[string]uint64) (uint64, error) {
	argc := len(f.arg_map)
	if argc == 0 {
		return caller.CallJit(f.function, 0, 0), nil
	}

	if args == nil {
		return 0, fmt.Errorf("requires parameters")
	}

	params := make([]uint64, argc)
	for name, index := range f.arg_map {
		value, ok := args[name]
		if !ok {
			return 0, fmt.Errorf("parameter %s not found", name)
		}
		params[index] = value
	}
	address := uint64(uintptr(unsafe.Pointer(&params[0])))
	return caller.CallJit(
		f.function,
		address,
		uint64(argc),
	), nil
}

func (f *Function) Float() bool {
	return f.float
}

// Returns false if the code is interpreted
// (which may very likely overflow the stack).
func IsJit() bool {
	return C.is_jit_supported() != 0
}
