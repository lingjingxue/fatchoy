// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

// 找出一个点p把线段(a,b)按比例m:n切分成2段
func LineSection(a, b Coord, m, n int) Coord {
	return Coord{
		X: (n*a.X + m*b.X) / (m + n),
		Y: (n*a.Y + m*b.Y) / (m + n),
	}
}

// 获取线段(a,b)的中间点
func GetMiddlePointOf(a, b Coord) Coord {
	return LineSection(a, b, 1, 1)
}

// 点q是否在线段(p1,p2)上
func IsOnSegment(q, p1, p2 Coord) bool {
	var v1 = NewVectorFrom(q, p1)
	var v2 = NewVectorFrom(q, p2)
	if v1.Cross(&v2) != 0 {
		return false
	}
	var dotProduct = v1.Dot(&v2)
	if dotProduct < 0 {
		return false
	}
	var v3 = NewVectorFrom(p1, p2)
	if dotProduct > v3.SquaredLength() {
		return false
	}
	return true
}

// 3个坐标的朝向
// 0 --> p, q, r 共线
// 1 --> 顺时针
// 2 --> 逆时针
// Slope of line segment (p1, p2): σ = (y2 - y1)/(x2 - x1)
// Slope of line segment (p2, p3): τ = (y3 - y2)/(x3 - x2)
// If  σ > τ, the orientation is clockwise (right turn)
func OrientationOfCoord(p1, p2, p3 Coord) int {
	var v = (p2.Y-p1.Y)*(p3.X-p2.X) - (p2.X-p1.X)*(p3.Y-p2.Y)
	if v > 0 {
		return 1
	} else if v < 0 {
		return -1
	}
	return 0
}

// 线段（p1,q1)与线段(p2,q2)是否相交
// 1. General Case:
// 	– (p1, q1, p2) 和 (p1, q1, q2) have different orientations and
// 	– (p2, q2, p1) and (p2, q2, q1) have different orientations.
// 2. Special Case
//	– (p1, q1, p2), (p1, q1, q2), (p2, q2, p1), and (p2, q2, q1) are all collinear and
//	– the x-projections of (p1, q1) and (p2, q2) intersect
//	– the y-projections of (p1, q1) and (p2, q2) intersect
func IsSegmentIntersect(p1, q1, p2, q2 Coord) bool {
	var o1 = OrientationOfCoord(p1, q1, p2)
	var o2 = OrientationOfCoord(p1, q1, q2)
	var o3 = OrientationOfCoord(p2, q2, p1)
	var o4 = OrientationOfCoord(p2, q2, q1)

	// General case
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Given three colinear points p, q, r, the function checks if
	// point q lies on line segment 'pr'
	onSegment := func(p, q, r Coord) bool {
		if (q.X <= MaxInt(p.X, r.X) && q.X >= MinInt(p.X, r.X)) &&
			(q.Y <= MaxInt(p.Y, r.Y) && q.Y >= MinInt(p.Y, r.Y)) {
			return true
		}
		return false
	}

	// Special Cases

	// p1, q1 and p2 are colinear and p2 lies on segment p1q1
	if o1 == 0 && onSegment(p1, p2, q1) {
		return true
	}
	// p1, q1 and q2 are colinear and q2 lies on segment p1q1
	if o2 == 0 && onSegment(p1, q2, q1) {
		return true
	}
	// p2, q2 and p1 are colinear and p1 lies on segment p2q2
	if o3 == 0 && onSegment(p2, p1, q2) {
		return true
	}
	// p2, q2 and q1 are colinear and q1 lies on segment p2q2
	if o4 == 0 && onSegment(p2, p1, q2) {
		return true
	}
	return false
}

// 线段（p1,q1)与线段(p2,q2)的交点，假设两条线段已经相交
// 当已经确定两个线段AB和CD是相交的时候，我们可以把AB和CD看成一个四边形的两条对角线，它们相交与点O，
// 我们可以通过三角形面积公式求出ABC和ABD的面积，它们的比值就是OC和OD的比值，然后再用定比分点公式求出O的坐标。
// see https://leetcode-cn.com/problems/intersection-lcci/solution/
func SegmentIntersectPoint(p1, q1, p2, q2 Coord) Coord {
	var area1 = CalcPolygonArea([]Coord{p1, q1, p2})
	var area2 = CalcPolygonArea([]Coord{p1, q1, p2})
	return LineSection(p2, q2, area1, area2)
}
