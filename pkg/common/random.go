package common

import (
	"math/rand"
	"time"
)

type intRange struct {
	min, max int
}

func (ir intRange) randomInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(ir.max-ir.min+1) + ir.min
}

func RandomInt(min, max int) int {
	ir := intRange{min, max}
	return ir.randomInt()
}

type floatRange struct {
	min, max float64
}

func (fr floatRange) randomFloat() float64 {
	rand.Seed(time.Now().UnixNano())
	return fr.min + (rand.Float64() * (fr.max + fr.min))
}

func RandomFloat(min, max float64) float64 {
	fr := floatRange{min, max}
	return fr.randomFloat()
}
