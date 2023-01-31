// Code generated by command: go run caller.go -out caller.s -stubs caller.go. DO NOT EDIT.

package caller

// Calls a jit_function_t without locking an OS thread.
// - f: a jit_function value
// - args: a pointer to an []uint64 argument array, probably from unsafe.Pointer(&args[0])
// - buf: temporary buffer to save the C code from malloc, must be larger than args
// - l: the length of the args
func CallJit(f uint64, args uint64, buf uint64, l uint64) uint64
