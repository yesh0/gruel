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

	"&&": []Operator{{0x20, 2, nil, "!jit_insn_and"}},
	"||": []Operator{{0x21, 2, nil, "!jit_insn_or"}},

	"len":   []Operator{{0x80, 1, nil, ":i:gruel_strlen"}},
	"index": []Operator{{0x81, 2, nil, ":i:s:gruel_index_of"}},

	// python build/ir/gen_go.py >> internal/ir/operators.go
	"=":       []Operator{{0x40, 2, nil, "gruel_insn_eq"}},
	"==":      []Operator{{0x41, 2, nil, "gruel_insn_eq"}},
	"!=":      []Operator{{0x42, 2, nil, "gruel_insn_ne"}},
	"<":       []Operator{{0x43, 2, nil, "jit_insn_lt"}},
	"<=":      []Operator{{0x44, 2, nil, "jit_insn_le"}},
	">":       []Operator{{0x45, 2, nil, "jit_insn_gt"}},
	">=":      []Operator{{0x46, 2, nil, "jit_insn_ge"}},
	"cmpl":    []Operator{{0x47, 2, nil, "jit_insn_cmpl"}},
	"cmpg":    []Operator{{0x48, 2, nil, "jit_insn_cmpg"}},
	"->bool":  []Operator{{0x49, 1, nil, "jit_insn_to_bool"}},
	"!":       []Operator{{0x4a, 1, nil, "jit_insn_to_not_bool"}},
	"acos":    []Operator{{0x4b, 1, nil, "jit_insn_acos"}},
	"asin":    []Operator{{0x4c, 1, nil, "jit_insn_asin"}},
	"atan":    []Operator{{0x4d, 1, nil, "jit_insn_atan"}},
	"atan2":   []Operator{{0x4e, 2, nil, "jit_insn_atan2"}},
	"ceil":    []Operator{{0x4f, 1, nil, "jit_insn_ceil"}},
	"cos":     []Operator{{0x50, 1, nil, "jit_insn_cos"}},
	"cosh":    []Operator{{0x51, 1, nil, "jit_insn_cosh"}},
	"exp":     []Operator{{0x52, 1, nil, "jit_insn_exp"}},
	"floor":   []Operator{{0x53, 1, nil, "jit_insn_floor"}},
	"log":     []Operator{{0x54, 1, nil, "jit_insn_log"}},
	"log10":   []Operator{{0x55, 1, nil, "jit_insn_log10"}},
	"pow":     []Operator{{0x56, 2, nil, "jit_insn_pow"}},
	"**":      []Operator{{0x57, 2, nil, "jit_insn_pow"}},
	"rint":    []Operator{{0x58, 1, nil, "jit_insn_rint"}},
	"round":   []Operator{{0x59, 1, nil, "jit_insn_round"}},
	"sin":     []Operator{{0x5a, 1, nil, "jit_insn_sin"}},
	"sinh":    []Operator{{0x5b, 1, nil, "jit_insn_sinh"}},
	"sqrt":    []Operator{{0x5c, 1, nil, "jit_insn_sqrt"}},
	"tan":     []Operator{{0x5d, 1, nil, "jit_insn_tan"}},
	"tanh":    []Operator{{0x5e, 1, nil, "jit_insn_tanh"}},
	"trunc":   []Operator{{0x5f, 1, nil, "jit_insn_trunc"}},
	"nan?":    []Operator{{0x60, 1, nil, "jit_insn_is_nan"}},
	"finite?": []Operator{{0x61, 1, nil, "jit_insn_is_finite"}},
	"inf?":    []Operator{{0x62, 1, nil, "jit_insn_is_inf"}},
	"abs":     []Operator{{0x63, 1, nil, "jit_insn_abs"}},
	"min":     []Operator{{0x64, 2, nil, "jit_insn_min"}},
	"max":     []Operator{{0x65, 2, nil, "jit_insn_max"}},
	"sign":    []Operator{{0x66, 1, nil, "jit_insn_sign"}},
}
