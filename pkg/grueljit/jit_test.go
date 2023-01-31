package grueljit_test

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

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
		assert.Greater(t, 0.00001, math.Abs(v-actual.(float64)))
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
	v, err := f.Call(map[string]any{"x": 9., "y": 4})
	assert.Nil(t, err)
	assert.Greater(t, 0.00001, math.Abs(22-v.(float64)))

	for i := 0; i < 0; i++ {
		x := math.Remainder(rand.Float64(), 10)
		y := int64(rand.Uint64())
		v, err := f.Call(map[string]any{"x": x, "y": uint64(y)})
		assert.Nil(t, err)
		assert.Greater(t, 0.0001, math.Abs((2*x+float64(y%9))-v.(float64)))
	}
}

func TestMoreArgs(t *testing.T) {
	expr := "(+ a b)"
	values := map[string]any{
		"a": 1.,
		"b": 2.,
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
		assert.Equal(t, float64((i+2)*(i+3)/2), sum)

		v := fmt.Sprintf("%c", rune('c'+i))
		expr = fmt.Sprintf("(+ %s %s)", v, expr)
		values[v] = float64(i + 3)
		var_map[v] = grueljit.TypeFloat
	}
}

var string_only = []string{
	"len", "index",
}

func TestOps(t *testing.T) {
	for name, ops := range ir.Operators {
		no_arithmetic := false
		for _, v := range string_only {
			if v == name {
				no_arithmetic = true
			}
		}
		if no_arithmetic {
			continue
		}
		for _, op := range ops {
			expr := "(" + name + " x y x y)"
			if op.Argc == 1 {
				expr = "(" + name + " x)"
			}
			f, err := grueljit.Compile(expr, map[string]byte{"x": grueljit.TypeFloat, "y": grueljit.TypeInt})
			assert.Nil(t, err, "compilation error: %s", expr)
			_, err = f.Call(map[string]any{"x": 1., "y": 1})
			assert.Nil(t, err)
			f.Free()

			f, err = grueljit.Compile(expr, map[string]byte{"y": grueljit.TypeFloat, "x": grueljit.TypeInt})
			assert.Nil(t, err)
			_, err = f.Call(map[string]any{"y": 1., "x": 1})
			assert.Nil(t, err)
			f.Free()
		}
	}
}

func TestString(t *testing.T) {
	assertResult(t, "(len \"Hello\")", 5)
	assertResult(t, "(== \"Hello\" \"Hello\")", 1)
	assertResult(t, "(== \"Hello\" \"hello\")", 0)
	assertResult(t, "(== \"1\" 1)", 0)
	assertResult(t, "(index \"The quick brown fox jumps over the lazy dog\" \"quick\")", 4)
}

func TestStringArgs(t *testing.T) {
	sentence := "The quick brown fox jumps over the lazy dog"
	f, err := grueljit.Compile("(index \""+sentence+"\" s)", map[string]byte{"s": grueljit.TypeString})
	assert.Nil(t, err)
	for i := 0; i < 300; i++ {
		a := rand.Int() % len(sentence)
		b := rand.Int() % len(sentence)
		if a > b {
			b, a = a, b
		}

		result, err := f.Call(map[string]any{"s": sentence[a:b]})
		assert.Nil(t, err)
		assert.Equal(t, uint64(strings.Index(sentence, sentence[a:b])), result)
	}
}
