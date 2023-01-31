package grueljit_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/yesh0/gruel/internal/caller"
	"github.com/yesh0/gruel/internal/ir"
	"github.com/yesh0/gruel/pkg/grueljit"
)

func TestCaller(t *testing.T) {
	assert.Equal(t, uint64(0), caller.CallJit(0, 0, 0))
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

	assertResult(t, "(/ 31536000. 365 24 60 60)", 1.)

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
	f, err := grueljit.Compile("(+ (* 2 x) (% y 9))", map[string]byte{
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

func TestMoreArgs(t *testing.T) {
	expr := "(+ a b)"
	values := map[string]uint64{
		"a": math.Float64bits(1),
		"b": math.Float64bits(2),
	}
	var_map := map[string]byte{
		"a": grueljit.TypeFloat,
		"b": grueljit.TypeFloat,
	}
	for i := 0; i < 32; i++ {
		f, err := grueljit.Compile(expr, var_map)
		assert.Nil(t, err)
		sum, err := f.Call(values)
		assert.Nil(t, err)
		assert.Equal(t, float64((i+2)*(i+3)/2), math.Float64frombits(sum))

		v := fmt.Sprintf("%c", rune('c'+i))
		expr = fmt.Sprintf("(+ %s %s)", v, expr)
		values[v] = math.Float64bits(float64(i + 3))
		var_map[v] = grueljit.TypeFloat
	}
}

func TestOps(t *testing.T) {
	for name, ops := range ir.Operators {
		for _, op := range ops {
			expr := "(" + name + " x y x y)"
			if op.Argc == 1 {
				expr = "(" + name + " x)"
			}
			f, err := grueljit.Compile(expr, map[string]byte{"x": grueljit.TypeFloat, "y": grueljit.TypeInt})
			assert.Nil(t, err)
			_, err = f.Call(map[string]uint64{"x": math.Float64bits(1.), "y": 1})
			assert.Nil(t, err)
			f.Free()

			f, err = grueljit.Compile(expr, map[string]byte{"y": grueljit.TypeFloat, "x": grueljit.TypeInt})
			assert.Nil(t, err)
			_, err = f.Call(map[string]uint64{"y": math.Float64bits(1.), "x": 1})
			assert.Nil(t, err)
			f.Free()
		}
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
