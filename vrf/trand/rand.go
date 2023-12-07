package trand

// RNGType represents the type of random number generator
type RNGType = string

// Constants defining different types of random number generators
const (
	RNGType_Normal RNGType = "normal" // Fast random number generator
	RNGType_Cipher RNGType = "cipher" // Slow, but cryptographically secure random number generator
)

// Rng is the interface for a random number generator
type Rng interface {
	Intn(n int) int
	Uint64() uint64
	Float64() float64
}

// NewRng creates a new random number generator based on the specified type and seed
func NewRng(seed [32]byte, typ RNGType) Rng {
	switch typ {
	case RNGType_Cipher:
		return newChacha8(seed)
	case RNGType_Normal:
		fallthrough
	default:
		r := &xoshiro256{}
		r.Seed(seed)
		return r
	}
}
