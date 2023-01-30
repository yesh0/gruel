package main

import (
	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("CallJit", NOSPLIT, "func(f uint64, args uint64, buf uint64, l uint64) uint64")
	Doc(
		"Calls a jit_function_t without locking an OS thread.",
		"- f: a jit_function value",
		"- args: a pointer to an []uint64 argument array",
		"- buf: temporary buffer to save the C code from malloc, must be larger than args",
		"- l: the length of the args",
	)
	// System V calling conventions
	// Arg 1
	Load(Param("f"), reg.RDI)
	// Arg 2
	Load(Param("args"), reg.RSI)
	// Arg 3
	Load(Param("buf"), reg.RDX)
	// Arg 4
	Load(Param("l"), reg.RCX)
	// Stack preparation
	AllocLocal(512)
	MOVQ(reg.RSP, reg.RBX)
	ADDQ(operand.I32(512), reg.RSP)
	MOVQ(operand.I64(-1<<4), reg.RAX)
	ANDQ(reg.RAX, reg.RSP)
	// Call relative
	CALL(operand.LabelRef("call_jit_function+0x00(SB)"))
	// Restore stack
	MOVQ(reg.RBX, reg.RSP)
	// Return value
	Store(reg.RAX, ReturnIndex(0))
	RET()
	Generate()
}
