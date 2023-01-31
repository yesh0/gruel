package grueljit_test

import (
	"strings"
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/yesh0/gruel/pkg/grueljit"
)

func BenchmarkNop(b *testing.B) {
	f, _ := grueljit.Compile("1", nil)
	f.Free()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Evaluating freed expressions returns immediately
		f.Call(nil)
	}
}

func BenchmarkSingle(b *testing.B) {
	f, _ := grueljit.Compile("1", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// LibJIT optimizes constant expressions away.
		// So it is unfair actually.
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

func BenchmarkAddition(b *testing.B) {
	f, _ := grueljit.Compile("(+ 1 x)", map[string]byte{"x": grueljit.TypeInt})
	args := map[string]any{"x": 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Call(args)
	}
}

func BenchmarkGovaluateAddition(b *testing.B) {
	e, _ := govaluate.NewEvaluableExpression("1 + x")
	args := map[string]any{"x": 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Evaluate(args)
	}
}

func BenchmarkSome(b *testing.B) {
	f, _ := grueljit.Compile(
		"(+ (/ (* requests_made requests_succeeded) 100) 90)",
		map[string]byte{
			"requests_made":      grueljit.TypeInt,
			"requests_succeeded": grueljit.TypeInt,
		},
	)
	args := map[string]any{
		"requests_made":      100,
		"requests_succeeded": 80,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Call(args)
	}
}

func BenchmarkGovaluateSome(b *testing.B) {
	e, _ := govaluate.NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
	args := map[string]any{
		"requests_made":      100,
		"requests_succeeded": 80,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Evaluate(args)
	}
}

func BenchmarkString(b *testing.B) {
	f, _ := grueljit.Compile(
		"(== \"Hello World\" s)",
		map[string]byte{
			"s": grueljit.TypeString,
		},
	)
	args := map[string]any{
		"s": "Hello World",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Call(args)
	}
}

func BenchmarkGovaluateString(b *testing.B) {
	e, _ := govaluate.NewEvaluableExpression("'Hello World' == s")
	args := map[string]any{
		"s": "Hello World",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Evaluate(args)
	}
}

func benchmarkLengthier(b *testing.B, n int) {
	sb := strings.Builder{}
	sb.WriteString("(*")
	for i := 0; i < n; i++ {
		sb.WriteString(" x")
	}
	sb.WriteString(")")
	f, _ := grueljit.Compile(sb.String(), map[string]byte{"x": grueljit.TypeFloat})
	args := map[string]any{
		"x": 1.23,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Call(args)
	}
}

func benchmarkGovaluateLengthier(b *testing.B, n int) {
	sb := strings.Builder{}
	sb.WriteString("x")
	for i := 0; i < n-1; i++ {
		sb.WriteString(" + x")
	}
	f, _ := govaluate.NewEvaluableExpression(sb.String())
	args := map[string]any{
		"x": 1.23,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Evaluate(args)
	}
}

func BenchmarkJit8(b *testing.B) {
	benchmarkLengthier(b, 8)
}

func BenchmarkGovaluate8(b *testing.B) {
	benchmarkGovaluateLengthier(b, 8)
}

func BenchmarkJit16(b *testing.B) {
	benchmarkLengthier(b, 16)
}

func BenchmarkGovaluate16(b *testing.B) {
	benchmarkGovaluateLengthier(b, 16)
}

func BenchmarkJit32(b *testing.B) {
	benchmarkLengthier(b, 32)
}

func BenchmarkGovaluate32(b *testing.B) {
	benchmarkGovaluateLengthier(b, 32)
}
