package main

import (
	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

func main() {
	frameSize := 512
	TEXT("CallJit", NOSPLIT|NOFRAME, "func(f uint64, args uint64, l uint64) uint64")
	Doc(
		"Calls a jit_function_t without locking an OS thread.",
		"- f: a jit_function value",
		"- args: a pointer to an []uint64 argument array, probably from unsafe.Pointer(&args[0])",
		"- l: the length of the args",
	)
	// System V calling conventions
	// Arg 1
	Load(Param("f"), reg.RDI)
	// Arg 2
	Load(Param("args"), reg.RSI)
	// Arg 3
	Load(Param("l"), reg.RDX)
	// Stack preparation
	AllocLocal(frameSize)
	MOVQ(reg.RSP, reg.RBX)
	ADDQ(operand.I32(frameSize), reg.RSP)
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
