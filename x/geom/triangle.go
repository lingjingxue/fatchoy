// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

// 三角形
type Triangle struct {
	Vertexes [3]Coord
}

func NewTriangle(a, b, c Coord) Triangle {
	return Triangle{
		Vertexes: [3]Coord{a, b, c},
	}
}

// 判断三角形是否合法，两边之和大于第三边
func (t *Triangle) IsValid() bool {
	var a = DistanceOf(t.Vertexes[0], t.Vertexes[1])
	var b = DistanceOf(t.Vertexes[1], t.Vertexes[2])
	var c = DistanceOf(t.Vertexes[0], t.Vertexes[2])
	return a+b > c && a+c > b && b+c > a
}

// 计算面积
func (t *Triangle) Area() int {
	return CalcPolygonArea(t.Vertexes[:])
}

// 如果p点在三角形（ABC)内，则PAB, PAC, PBC的面积应该与ABC相等
func (t *Triangle) Contains(p Coord) bool {
	var area1 = CalcPolygonArea([]Coord{p, t.Vertexes[0], t.Vertexes[1]})
	var area2 = CalcPolygonArea([]Coord{p, t.Vertexes[1], t.Vertexes[2]})
	var area3 = CalcPolygonArea([]Coord{p, t.Vertexes[0], t.Vertexes[2]})
	var area = CalcPolygonArea(t.Vertexes[:])
	return area1+area2+area3 == area
}
