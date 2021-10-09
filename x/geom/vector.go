// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package geom

import (
	"math"
)

// 二维向量
type Vector struct {
	X, Y int
}

func NewVectorFrom(a, b Coord) Vector {
	return Vector{
		X: b.X - a.X,
		Y: b.Y - a.Y,
	}
}

// 向量长度
func (a *Vector) Length() float64 {
	return math.Hypot(float64(a.X), float64(a.Y))
}

func (a *Vector) SquaredLength() int {
	return a.X*a.X + a.Y*a.Y
}

// 单位向量
func (a *Vector) Normalize() Vector {
	if a.X == 0 && a.Y == 0 {
		return *a
	}
	return a.Mul(1 / a.Length())
}

// 加法
func (a *Vector) Add(b *Vector) Vector {
	return Vector{
		X: a.X + b.X,
		Y: a.Y + b.Y,
	}
}

// 减法
func (a *Vector) Sub(b *Vector) Vector {
	return Vector{
		a.X - b.X,
		a.Y - b.Y,
	}
}

// 乘法
func (a *Vector) Mul(m float64) Vector {
	return Vector{
		X: int(float64(a.X) * m),
		Y: int(float64(a.Y) * m),
	}
}

// 点积（dot product)
func (a *Vector) Dot(b *Vector) int {
	return a.X*b.X + a.Y*b.Y
}

// 叉积(cross product)
func (a *Vector) Cross(b *Vector) int {
	return a.X*b.Y - a.Y*b.X
}

// 按比例截断
func (a *Vector) Trunc(ratio float64) Vector {
	return Vector{
		X: int(ratio * float64(a.X)),
		Y: int(ratio * float64(a.Y)),
	}
}

// 向量按照角度逆时针旋转获得新的向量，如果需要顺时针旋转，angle传入负值
// 对于任意两个不同点A和B，A绕B旋转θ角度后的坐标为： (Δx*cosθ- Δy * sinθ+ xB, Δy*cosθ + Δx * sinθ+ yB )
func (a *Vector) Rotate(angle float64) Vector {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	x := float64(a.X)*cos - float64(a.Y)*sin
	z := float64(a.X)*sin + float64(a.Y)*cos
	return Vector{
		X: int(x),
		Y: int(z),
	}
}

// 获取2个相交向量的角度, cosθ = a x b / |a| |b|
func (a *Vector) GetAngle(b *Vector) float64 {
	t := float64(a.Dot(b)) / a.Length() * b.Length()
	angle := math.Acos(t)
	if math.IsNaN(angle) {
		if t > 0 {
			angle = 0
		} else {
			angle = math.Pi
		}
	}
	return angle
}

// 转换成坐标
func (a Vector) ToCoord(start Coord) Coord {
	return Coord{
		X: start.X + a.X,
		Y: start.Y + a.Y,
	}
}
