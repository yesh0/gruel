package gruelparser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yesh0/gruel/pkg/gruelparser"
)

func assertToken(t assert.TestingT, expected string, expectedType gruelparser.TokenType, expr string, n int) {
	r := gruelparser.NewTokenReader(expr)
	for i := 0; i < n; i++ {
		_, _, err := r.NextToken()
		assert.Nil(t, err)
	}
	token, actualType, err := r.NextToken()
	assert.Nil(t, err)
	assert.Equal(t, expectedType, actualType)
	assert.Equal(t, expected, token)
}

func assertTokens(t assert.TestingT, expected string, expectedType gruelparser.TokenType, expr string) {
	assertToken(t, expected, expectedType, expr, 0)
	assertToken(t, expected, expectedType, "("+expr, 1)
	assertToken(t, expected, expectedType, "("+expr+")", 1)
	assertToken(t, expected, expectedType, "(+ "+expr+" 123)", 2)
	assertToken(t, expected, expectedType, "(+ "+expr+" \"123\")", 2)
	assertToken(t, expected, expectedType, "(+ (- "+expr+" 321) 123)", 4)
	assertToken(t, expected, expectedType, "(- (- 3 1) (+ \"str\" sym "+expr+" 321) 123)", 11)
}

func TestSplit(t *testing.T) {
	// strings
	assertTokens(t, "", gruelparser.TypeString, "\"\"")
	assertTokens(t, "esc \t\n\u1234\"'\\", gruelparser.TypeString, "\"esc \\t\\n\\u1234\\\"'\\\\\"")

	// floats
	assertTokens(t, "0.123", gruelparser.TypeFloat, "0.123")
	assertTokens(t, ".456f", gruelparser.TypeFloat, ".456f")
	assertTokens(t, "-.1456f", gruelparser.TypeFloat, "-.1456f")
	// ints
	assertTokens(t, "0x123ABC", gruelparser.TypeInt, "0x123ABC")
	assertTokens(t, "0o556677", gruelparser.TypeInt, "0o556677")
	assertTokens(t, "-0556677", gruelparser.TypeInt, "-0556677")
	assertTokens(t, "-0556677", gruelparser.TypeInt, "-0556677")

	// bools
	assertTokens(t, "true", gruelparser.TypeBool, "true")
	assertTokens(t, "false", gruelparser.TypeBool, "false")

	// symbols
	assertTokens(t, "+", gruelparser.TypeSymbol, "+")
	assertTokens(t, "-", gruelparser.TypeSymbol, "-")
	assertTokens(t, "contains?", gruelparser.TypeSymbol, "contains?")
	assertTokens(t, "'\\n", gruelparser.TypeSymbol, "'\\n")

	// Parenthesis
	assertTokens(t, "(", gruelparser.TypeParenthesis, "(")
	assertTokens(t, ")", gruelparser.TypeParenthesis, ")")
}
