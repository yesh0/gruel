package grueljit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yesh0/gruel/internal/caller"
)

func TestCaller(t *testing.T) {
	assert.Equal(t, uint64(0xdeadbeef00000000), caller.CallJit(0, 0, 0, 0))
}
