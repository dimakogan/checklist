package boosted

import (
	"testing"

	"gotest.tools/assert"
)

func TestPRP(t *testing.T) {
	key := []byte{0x01, 0x10, 0x02, 0x12, 0x03, 0x13, 0x04, 0x14}
	blockLenBits := 6
	prp, err := NewPRP(key, blockLenBits)
	assert.NilError(t, err)

	domainSize := 1 << blockLenBits
	outs := make(map[uint32]bool)

	for i := 0; i < domainSize; i++ {
		val := prp.Eval(uint32(i))
		assert.Equal(t, prp.Invert(val), uint32(i))
		outs[val] = true
	}
	assert.Equal(t, len(outs), domainSize)
	for i := 0; i < domainSize; i++ {
		assert.Check(t, outs[uint32(i)])
	}
}
