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
