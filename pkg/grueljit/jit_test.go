package grueljit_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/yesh0/gruel/internal/caller"
	"github.com/yesh0/gruel/pkg/grueljit"
	"github.com/yesh0/gruel/pkg/gruelparser"
)

func TestCaller(t *testing.T) {
	assert.Equal(t, uint64(0), caller.CallJit(0, 0, 0, 0))
}

func assertResult(t *testing.T, expr string, result any) {
	f, err := grueljit.Compile(expr, nil)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, f)
	switch v := result.(type) {
	case int:
		actual, err := f.Call(nil)
		assert.Nil(t, err)
		assert.Equal(t, uint64(v), actual)
	case float64:
		actual, err := f.Call(nil)
		assert.Nil(t, err)
		assert.Greater(t, 0.00001, math.Abs(v-math.Float64frombits(actual)))
	default:
		t.Fail()
	}

	f.Free()
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

func TestArgs(t *testing.T) {
	f, err := grueljit.Compile("(+ (* 2 x) (% y 9))", map[string]gruelparser.TokenType{
		"x": grueljit.TypeFloat,
		"y": grueljit.TypeInt,
	})
	assert.Nil(t, err)
	assert.True(t, f.Float())
	v, err := f.Call(map[string]uint64{"x": math.Float64bits(9.), "y": 4})
	assert.Nil(t, err)
	assert.Greater(t, 0.00001, math.Abs(22-math.Float64frombits(v)))

	for i := 0; i < 10000; i++ {
		x := math.Remainder(rand.Float64(), 10)
		y := int64(rand.Uint64())
		v, err := f.Call(map[string]uint64{"x": math.Float64bits(x), "y": uint64(y)})
		assert.Nil(t, err)
		assert.Greater(t, 0.0001, math.Abs((2*x+float64(y%9))-math.Float64frombits(v)))
	}
}

func BenchmarkEvaluationSingle(b *testing.B) {
	f, _ := grueljit.Compile("1", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Call(nil)
	}
}

func BenchmarkGovaluateSingle(b *testing.B) {
	e, _ := govaluate.NewEvaluableExpression("1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Evaluate(nil)
	}
}
