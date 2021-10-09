// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

// 矩形
type Rectangle struct {
	Coord             // 左下角原点
	Width, Height int // 宽度、高度
}

func NewRectangle(x, y, w, h int) Rectangle {
	return Rectangle{
		Coord: Coord{X: x,
			Y: y,
		},
		Width:  w,
		Height: h,
	}
}

// 四个顶点
func (r *Rectangle) GetVertexes() [4]Coord {
	return [4]Coord{
		{r.X, r.Y},
		{r.X + r.Width, r.Y},
		{r.X + r.Width, r.Y + r.Height},
		{r.X, r.Y + r.Height},
	}
}

// 展开或者收缩矩形
func (r *Rectangle) Inflate(width, height int) {
	r.X -= width
	r.Y -= height
	r.Width += 2 * width
	r.Height += 2 * height
}

// 点是否在矩形内
func (r *Rectangle) Contains(x, y int) bool {
	return r.X <= x && x < r.X+r.Width &&
		r.Y <= y && y < r.Y+r.Height
}

// 点是否在矩形内
func (r *Rectangle) ContainsPoint(pt Coord) bool {
	return r.Contains(pt.X, pt.Y)
}

// 是否包含
func (r *Rectangle) ContainsRegion(rec *Rectangle) bool {
	return r.X <= rec.X && (rec.X+rec.Width) <= (r.X+r.Width) &&
		r.Y <= rec.Y && (rec.Y+rec.Height) <= (r.Y+r.Height)
}

// 是否相交
func (r *Rectangle) IsIntersectsWith(rec *Rectangle) bool {
	return (rec.X < r.X+r.Width) && r.X < (rec.X+rec.Width) &&
		(rec.Y < r.Y+r.Height) && r.Y < (rec.Y+rec.Height)
}

// 矩形相交区域
func RectIntersect(a *Rectangle, b *Rectangle) Rectangle {
	var x1 = MaxInt(a.X, b.X)
	var x2 = MaxInt(a.X+a.Width, b.X+b.Width)
	var y1 = MaxInt(a.Y, b.Y)
	var y2 = MaxInt(a.Y+a.Height, b.Y+b.Height)
	if x2 >= x1 && y2 >= y1 {
		return NewRectangle(x1, y1, x2-x1, y2-y1)
	}
	return Rectangle{}
}

// 矩形结合区域
func RectUnion(a *Rectangle, b *Rectangle) Rectangle {
	var x1 = MaxInt(a.X, b.X)
	var x2 = MaxInt(a.X+a.Width, b.X+b.Width)
	var y1 = MaxInt(a.Y, b.Y)
	var y2 = MaxInt(a.Y+a.Height, b.Y+b.Height)
	return NewRectangle(x1, y1, x2-x1, y2-y1)
}
