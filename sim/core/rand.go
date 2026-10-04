package core

import (
	"math"
	"math/rand"
)

// implementing Source or Source64 is possible, but adds too much overhead
type Rand interface {
	Next() uint64
	NextFloat64() float64
	Seed(int64)

	GetSeed() int64

	rand.Source64
}

func NewSplitMix(seed uint64) *SplitMix64 {
	return &SplitMix64{state: seed, start: seed}
}

// adapted from https://prng.di.unimi.it/splitmix64.c
type SplitMix64 struct {
	state uint64
	start uint64 // track starting seed
}

func (sm *SplitMix64) Next() uint64 {
	sm.state += 0x9e3779b97f4a7c15
	result := sm.state
	result = (result ^ (result >> 30)) * 0xbf58476d1ce4e5b9
	result = (result ^ (result >> 27)) * 0x94d049bb133111eb
	return result ^ (result >> 31)
}

func (sm *SplitMix64) NextFloat64() float64 {
	return float64(sm.Next()>>11) * 0x1p-53
}

func (sm *SplitMix64) Seed(s int64) {
	sm.start = uint64(s)
	sm.state = uint64(s)
}

func (sm *SplitMix64) GetSeed() int64 {
	return int64(sm.start)
}

func (sm *SplitMix64) Int63() int64 {
	return int64(sm.Next() & math.MaxInt64)
}

func (sm *SplitMix64) Uint64() uint64 {
	return sm.Next()
}
