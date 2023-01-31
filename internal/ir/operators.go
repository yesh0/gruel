package ir

type Operator struct {
	Opcode int
	// Argument count
	Argc int
	// - 0 or nil: libjit converts things automatically
	// - Others: see grueljit.Type*
	Argf []byte
	// The libjit function to call
	//
	// - Prefix with ':' to indicate that it is an intrinsic function
	JitFunction string
}

var Operators = map[string]([]Operator){
	"+":   []Operator{{0x01, 2, nil, "jit_insn_add"}},
	"-":   []Operator{{0x02, 2, nil, "jit_insn_sub"}, {0x03, 1, nil, "jit_insn_neg"}},
	"*":   []Operator{{0x04, 2, nil, "jit_insn_mul"}},
	"/":   []Operator{{0x05, 2, nil, "jit_insn_div"}},
	"%":   []Operator{{0x06, 2, nil, "jit_insn_rem"}},
	"&":   []Operator{{0x07, 2, nil, "jit_insn_and"}},
	"|":   []Operator{{0x08, 2, nil, "jit_insn_or"}},
	"^":   []Operator{{0x09, 2, nil, "jit_insn_xor"}, {0x0a, 1, nil, "jit_insn_not"}},
	"<<":  []Operator{{0x0b, 2, nil, "jit_insn_shl"}},
	">>":  []Operator{{0x0c, 2, nil, "jit_insn_shr"}},
	">>>": []Operator{{0x0d, 2, nil, "jit_insn_ushr"}},

	"**":    []Operator{{0x10, 2, nil, "jit_insn_pow"}},
	"acos":  []Operator{{0x11, 1, nil, "jit_insn_acos"}},
	"asin":  []Operator{{0x12, 1, nil, "jit_insn_asin"}},
	"atan":  []Operator{{0x13, 1, nil, "jit_insn_atan"}},
	"atan2": []Operator{{0x14, 2, nil, "jit_insn_atan2"}},
	"cos":   []Operator{{0x15, 1, nil, "jit_insn_cos"}},
	"cosh":  []Operator{{0x16, 1, nil, "jit_insn_cosh"}},
	"exp":   []Operator{{0x17, 1, nil, "jit_insn_exp"}},
	"log":   []Operator{{0x18, 1, nil, "jit_insn_log"}},
	"log10": []Operator{{0x19, 1, nil, "jit_insn_log10"}},
	"pow":   []Operator{{0x1a, 2, nil, "jit_insn_pow"}},
	"sin":   []Operator{{0x1b, 1, nil, "jit_insn_sin"}},
	"sinh":  []Operator{{0x1c, 1, nil, "jit_insn_sinh"}},
	"sqrt":  []Operator{{0x1d, 1, nil, "jit_insn_sqrt"}},
	"tan":   []Operator{{0x1e, 1, nil, "jit_insn_tan"}},
	"tanh":  []Operator{{0x1f, 1, nil, "jit_insn_tanh"}},
}
