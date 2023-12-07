// Package xoshiro256 implements the xoshiro256** pseudo-random
// number generator. The implementation is based on the public domain
// [C implementation].
//
// [C implementation]: https://xoshiro.di.unimi.it/xoshiro256starstar.c
// github.com/seedhammer/seedhammer/blob/v1.1.1/bc/xoshiro256/xoshiro.go

package trand

import (
	"encoding/binary"
	"math"
)

type xoshiro256 struct {
	state [4]uint64
}

func (s *xoshiro256) Seed(seed [32]byte) {
	s.state[0] = binary.BigEndian.Uint64(seed[0:8])
	s.state[1] = binary.BigEndian.Uint64(seed[8:16])
	s.state[2] = binary.BigEndian.Uint64(seed[16:24])
	s.state[3] = binary.BigEndian.Uint64(seed[24:32])
}

func (s *xoshiro256) Uint64() uint64 {
	result := rotl(s.state[1]*5, 7) * 9

	t := s.state[1] << 17

	s.state[2] ^= s.state[0]
	s.state[3] ^= s.state[1]
	s.state[1] ^= s.state[2]
	s.state[0] ^= s.state[3]

	s.state[2] ^= t

	s.state[3] = rotl(s.state[3], 45)

	return result
}

func (s *xoshiro256) Intn(n int) int {
	return int(s.Float64() * float64(n))
}

func (s *xoshiro256) Float64() float64 {
	return float64(s.Uint64()) / (float64(math.MaxUint64) + 1)
}

func rotl(x uint64, k int) uint64 {
	return (x << k) | (x >> (64 - k))
}
