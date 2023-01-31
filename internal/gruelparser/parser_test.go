package gruelparser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yesh0/gruel/internal/gruelparser"
)

func assertError(t assert.TestingT, expr string, msg string) {
	_, err := gruelparser.Parse(expr)
	assert.Equal(t, msg, err.Error())
}

func assertAst(t assert.TestingT, expr string, ast string) {
	node, err := gruelparser.Parse(expr)
	assert.Nil(t, err)
	assert.Equal(t, ast, node.String())
}

func TestParser(t *testing.T) {
	assertError(t, "()", "expecting symbolic operator")
	assertError(t, "(", "EOF")
	assertError(t, "(\"str\")", "expecting symbolic operator")
	assertError(t, "(+", "EOF")

	assertAst(t, "\"\"", "\"\"")
	assertAst(t,
		"(with-eval-after-load 'evil-maps\n"+
			"  (define-key evil-motion-state-map (kbd \"SPC\") nil)\n"+
			"  (define-key evil-motion-state-map (kbd \"RET\") nil)\n"+
			"  (define-key evil-motion-state-map (kbd \"TAB\") nil))",
		"(with-eval-after-load 'evil-maps "+
			"(define-key evil-motion-state-map (kbd \"SPC\") nil) "+
			"(define-key evil-motion-state-map (kbd \"RET\") nil) "+
			"(define-key evil-motion-state-map (kbd \"TAB\") nil))")
}
