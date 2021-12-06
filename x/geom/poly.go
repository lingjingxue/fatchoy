// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

// 凸多边形
type ConvexPolygon struct {
	Vertexes []Coord
}

func (c *ConvexPolygon) Contains(x, y int) bool {
	// TODO:
	return false
}

// https://en.wikipedia.org/wiki/Shoelace_formula
func CalcPolygonArea(vertexes []Coord) int {
	var area int
	var j = len(vertexes) - 1 // j is previous vertex to i
	for i := 0; i < len(vertexes); i++ {
		area += (vertexes[j].X + vertexes[i].X) * (vertexes[j].Y - vertexes[i].Y)
		j = i
	}
	return AbsInt(area / 2)
}
