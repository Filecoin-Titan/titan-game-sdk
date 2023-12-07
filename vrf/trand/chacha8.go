package trand

import (
	"unsafe"

	cc "github.com/nixberg/chacha-rng-go"
)

// chacha8 represents a ChaCha8 random number generator
type chacha8 struct {
	s *cc.ChaCha
}

// newChacha8 creates a new ChaCha8 random number generator with the given seed
func newChacha8(seed [32]byte) *chacha8 {
	p := unsafe.Pointer(&seed)
	seed32s := unsafe.Slice((*uint32)(p), 8)
	var seed32 [8]uint32
	for i := 0; i < 8; i++ {
		seed32[i] = seed32s[i]
	}

	s := cc.Seeded8(seed32, 0)
	return &chacha8{
		s: s,
	}
}

// Uint64 generates a random uint64 using ChaCha8
func (c *chacha8) Uint64() uint64 {
	return c.s.Uint64()
}

// Float64 generates a random float64 using ChaCha8
func (c *chacha8) Float64() float64 {
	return c.s.Float64()
}

// Intn generates a random integer in the range [0, n) using ChaCha8
func (c *chacha8) Intn(n int) int {
	return int(c.Float64() * float64(n))
}
