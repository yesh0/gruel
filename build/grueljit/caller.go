package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strings"

	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

const no_local_marker = "FIXME: EDIT THIS TO ADD IN: NO_LOCAL_POINTERS"
const tls_marker = "FIXME: EDIT THIS TO ADD IN: MOVQ (TLS), DI"

func main() {
	alignment := 16
	TEXT("CallJit", NOSPLIT|NOFRAME, "func(f uint64, args []uint64, stack uint64) uint64")
	Doc(
		"Calls a jit_function_t without locking an OS thread.",
		"- f: a jit_function value",
		"- args: a pointer to an []uint64 argument array, probably from unsafe.Pointer(&args[0])",
		"- stack: stack size needed",
	)
	Comment(no_local_marker)
	// Reserves space for stack alignment
	AllocLocal(alignment)

	// System V
	arg1 := reg.RDI
	arg2 := reg.RSI

	Label("jit_entry")
	g := arg1
	top := arg2
	Comment("top(SI) = SP - max_stack_size")
	MOVQ(reg.RSP, top)
	Load(Param("stack"), g)
	SUBQ(g, top)
	Comment("if top <= g(DI).stackguard1 { goto stack_grow }")
	Comment(tls_marker)
	CMPQ(top, operand.Mem{Base: g, Disp: 16})
	JBE(operand.LabelRef("jit_stack_grow"))

	Comment("System V calling conventions")
	Load(Param("f"), arg1)
	Load(Param("args").Base(), arg2)

	Comment("Align the stack")
	MOVQ(reg.RSP, reg.RBX)
	ORQ(operand.I8(alignment-1), reg.RSP)
	INCQ(reg.RSP)

	Comment("Call relative")
	CALL(operand.LabelRef("call_jit_function+0x00(SB)"))

	Comment("Restore stack")
	MOVQ(reg.RBX, reg.RSP)

	Comment("System V: Return value")
	Store(reg.RAX, ReturnIndex(0))
	RET()

	Label("jit_stack_grow")
	CALL(operand.LabelRef("runtime·morestack_noctxt<>+0x00(SB)"))
	JMP(operand.LabelRef("jit_entry"))

	Generate()
	replaceMarkers()
}

const file = "caller.s"

func replaceMarkers() {
	content, err := os.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}
	s := bufio.NewScanner(bytes.NewBuffer(content))
	output := bytes.Buffer{}
	for s.Scan() {
		line := string(s.Bytes())
		switch {
		case strings.Contains(line, "#include \"textflag.h\""):
			output.WriteString(line)
			output.WriteString("\n#include \"funcdata.h\"")
		case strings.Contains(line, no_local_marker):
			output.WriteString("\tNO_LOCAL_POINTERS")
		case strings.Contains(line, tls_marker):
			output.WriteString("\tMOVQ (TLS), DI")
		default:
			output.WriteString(line)
		}
		output.WriteRune('\n')
	}
	os.WriteFile(file, output.Bytes(), 0644)
}
