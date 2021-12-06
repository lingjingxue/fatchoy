// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package mathext

import (
	"math/rand"
	"sync"
)

// 线性同余法的随机数生成器
// see https://en.wikipedia.org/wiki/Linear_congruential_generator
type LCG struct {
	seed  uint32
	guard sync.Mutex
}

func (g *LCG) Seed(seed uint32) {
	g.guard.Lock()
	g.seed = seed*214013 + 2531011
	g.guard.Unlock()
}

func (g *LCG) Rand() uint32 {
	g.guard.Lock()
	g.seed = g.seed*214013 + 2531011
	var r = uint32(g.seed>>16) & 0x7fff
	g.guard.Unlock()
	return r
}

// Random integer in [min, max]
func RandInt(min, max int) int {
	if min > max {
		panic("RandInt,min greater than max")
	}
	if min == max {
		return min
	}
	return rand.Intn(max-min+1) + min
}

// Random number in [min, max]
func RandFloat(min, max float64) float64 {
	if min > max {
		panic("RandFloat: min greater than max")
	}
	if min == max {
		return min
	}
	return rand.Float64()*(max-min) + min
}

// 集合内随机取数, [min,max]
func RangePerm(min, max int) []int {
	if min > max {
		panic("RangePerm: min greater than max")
	}
	if min == max {
		return []int{min}
	}
	list := rand.Perm(max - min + 1)
	for i := 0; i < len(list); i++ {
		list[i] += min
	}
	return list
}
