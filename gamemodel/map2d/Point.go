package map2d

import (
	"fmt"
	"math"
)

type Point []int32 //地图位置描述 [row,col]
//取行号
func (p Point) GetRow() int32 {
	if len(p) > 0 {
		return p[0]
	}
	return -1
}

//取列号
func (p Point) GetCol() int32 {
	if len(p) > 1 {
		return p[1]
	}
	return -1
}

//取列号
func (p Point) Add(row, col int32) Point {
	return Point{p.GetRow() + row, p.GetCol() + col}
}

//转成字符串
func (p Point) ToString() string {
	return fmt.Sprintf("(%d,%d)", p.GetRow(), p.GetCol())
}

//值相等
func (p Point) Equal(t Point) bool {
	lenp := len(p)
	lent := len(t)
	if lenp != lent {
		return false
	}

	maxlen := 2
	if lenp < 2 {
		maxlen = lenp
	}

	for i := 0; i < maxlen; i++ {
		if p[i] != t[i] {
			return false
		}
	}
	return true
}

//计算曼哈顿距离
func (a1 Point) CalcDistanceH(a2 Point) int {
	return int(math.Abs(float64(a1.GetRow()-a2.GetRow())) + math.Abs(float64(a1.GetCol()-a2.GetCol())))
}

type PointSet []Point

func (this PointSet) IndexOf(p Point) int {
	for idx, v := range this {
		if v.Equal(p) {
			return idx
		}
	}
	return -1
}
