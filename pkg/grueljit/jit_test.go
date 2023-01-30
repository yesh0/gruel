package grueljit_test

import (
	"math"
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/yesh0/gruel/internal/caller"
	"github.com/yesh0/gruel/pkg/grueljit"
	"github.com/yesh0/gruel/pkg/gruelparser"
	"github.com/yesh0/gruel/pkg/ir"
)

func TestCaller(t *testing.T) {
	assert.Equal(t, uint64(0), caller.CallJit(0, 0, 0, 0))
}

func assertResult(t *testing.T, expr string, result any) {
	ast, err := gruelparser.Parse(expr)
	assert.Nil(t, err)
	code, err := ir.Compile(&ast, make(map[string]gruelparser.TokenType))
	assert.Nil(t, err)
	f := grueljit.CompileOpcodes(code)
	assert.NotEqual(t, 0, f)
	switch v := result.(type) {
	case int:
		assert.Equal(t, uint64(v), caller.CallJit(f, 0, 0, 0))
	case float64:
		assert.Greater(t, 0.00001, math.Abs(v-math.Float64frombits(caller.CallJit(f, 0, 0, 0))))
	default:
		t.Fail()
	}

	grueljit.Free(f)
}

func TestJit(t *testing.T) {
	assert.True(t, grueljit.IsJit())
	assertResult(t, "1", 1)

	assertResult(t, "(+ 123000 456)", 123456)
	assertResult(t, "(- 123000 456)", 123000-456)
	assertResult(t, "(* 123 1000)", 123000)
	assertResult(t, "(/ 1230 10)", 123)
	assertResult(t, "(% 123 100)", 23)

	// Bool
	assertResult(t, "(+ true true)", 2)
	assertResult(t, "(+ true false)", 1)

	// Floating point
	assertResult(t, "(+ 1.23 0.00456)", 1.23456)
	assertResult(t, "(/ 10 0.5)", 20.0)

	// Integer zero-division
	assertResult(t, "(/ 123000 0)", 0)
	assertResult(t, "(% 123000 0)", 0)

	assertResult(t, "(+ (- (* (/ 4 (% 6 5)) 3) 2) 1)", (4/(6%5))*3-2+1)
}

func BenchmarkEvaluationSingle(b *testing.B) {
	ast, _ := gruelparser.Parse("1")
	code, _ := ir.Compile(&ast, make(map[string]gruelparser.TokenType))
	f := grueljit.CompileOpcodes(code)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		caller.CallJit(f, 0, 0, 0)
	}
}

func BenchmarkGovaluateSingle(b *testing.B) {
	e, _ := govaluate.NewEvaluableExpression("1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Evaluate(nil)
	}
}
