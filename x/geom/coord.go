// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

import "math"

// 坐标
type Coord struct {
	X int
	Y int
}

func NewCoord(x, y int) Coord {
	return Coord{x, y}
}

// 两点之间的距离
func DistanceOf(a, b Coord) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Hypot(float64(dx), float64(dy))
}
