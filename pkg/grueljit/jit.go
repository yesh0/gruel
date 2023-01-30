package grueljit

/*

#cgo CFLAGS:  -I../../libjit/include
#cgo LDFLAGS: -L../../libjit -ljit -lm
#include "gruel_jit.h"

*/
import "C"
import (
	"runtime"
	"unsafe"
)

// Compiles the byte code and returns a function handle.
func CompileOpcodes(code []byte) uint64 {
	handle := uint64(C.compile_opcodes((C.long)(len(code)/8), (*C.long)(unsafe.Pointer(&code[0]))))
	runtime.KeepAlive(code)
	return handle
}

// Frees the resources.
func Free(function uint64) {
	C.free_function((C.long)(function))
}

// Returns false if the code is interpreted
// (which may very likely overflow the stack).
func IsJit() bool {
	return C.is_jit_supported() != 0
}
