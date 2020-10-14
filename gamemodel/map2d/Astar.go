package map2d

import (
	"errors"
	"time"

	"github.com/binglai-com/gameutils/gamelog/filelog"
	"github.com/globalsign/mgo/bson"
)

//A星寻路
type Astar struct {
	maxrow     int32
	maxcol     int32
	nodes      [][]*pathnode
	directions [][]int32
}

//寻路节点
type pathnode struct {
	f      int
	g      int
	h      int
	pathid string    //当前所属路径编号
	cl     int       //是否进入关闭队列
	cur    Point     //当前节点信息
	parent *pathnode //寻路父节点
}

//创建并初始化A星寻路
func CreateAstar(maxrow, maxcol int32, dir [][]int32) (*Astar, error) {
	if maxrow < 0 || maxcol < 0 {
		return nil, errors.New("CreateAstar maxrow,maxcol invalid.")
	}
	a := new(Astar)
	a.maxrow, a.maxcol = maxrow, maxcol
	a.nodes = make([][]*pathnode, maxrow)
	for i := int32(0); i < maxrow; i++ {
		a.nodes[i] = make([]*pathnode, maxcol)
		for j := int32(0); j < maxcol; j++ {
			var pn pathnode
			pn.cur = Point{i, j}
			a.nodes[i][j] = &pn
		}
	}
	a.directions = dir
	return a, nil
}

//定位点
func _findpoint(a *Astar, row, col int32) (Point, bool) {
	if row < 0 || row >= a.maxrow {
		return nil, false
	}

	if col < 0 || col >= a.maxcol {
		return nil, false
	}

	return a.nodes[row][col].cur, true
}

//启动寻路算法
func (m *Astar) FindPath(from, to Point, passcheck func(p Point) bool) []Point {
	st := time.Now()

	maxrow, maxcol := m.maxrow, m.maxcol
	if from.GetRow() < 0 || from.GetRow() >= maxrow || from.GetCol() < 0 || from.GetCol() >= maxcol {
		return nil
	}
	if to.GetRow() < 0 || to.GetRow() >= maxrow || to.GetCol() < 0 || to.GetCol() >= maxcol {
		return nil
	}

	if passcheck != nil && !passcheck(to) { //判断目的地的通过性
		filelog.DEBUG("astar", " find path return 111")
		return nil
	}
	var tmpcloselist = make([]*pathnode, 0)
	openlist := make([]*pathnode, 0)
	curpathid := bson.NewObjectId().Hex()
	startnode := m.nodes[from.GetRow()][from.GetCol()]
	startnode.g = 0
	startnode.h = 0
	startnode.pathid = curpathid
	startnode.cl = 1
	startnode.parent = nil
	openlist = append(openlist, startnode)
	for len(openlist) > 0 {
		dei := -1
		maxf := 99999999
		for i := 0; i < len(openlist); i += 1 { ///取出最小F节点
			if maxf > openlist[i].f {
				maxf = openlist[i].f
				dei = i
			}
		}
		if dei == -1 {
			cost := time.Since(st)
			if cost > 2*time.Millisecond {
				filelog.INFO("astar", "get path nil cost:", cost, " from:", from, " to:", to, " open:", len(openlist))
			}
			filelog.DEBUG("astar", " find path return 222")
			return nil
		}
		var cnode *pathnode
		cnode = openlist[dei]
		openlist = append(openlist[:dei], openlist[dei+1:]...)
		cnode.cl = 2 ///关闭
		tmpcloselist = append(tmpcloselist, cnode)
		for _, v := range m.directions {
			npos, find := _findpoint(m, cnode.cur.GetRow()+v[0], cnode.cur.GetCol()+v[1])
			if !find {
				continue
			}

			if passcheck != nil && !passcheck(npos) {
				continue
			}

			///检查closelist

			add := m.nodes[npos.GetRow()][npos.GetCol()]

			if add.pathid == curpathid && add.cl == 2 {
				continue
			}
			if npos.Equal(to) {
				res := make([]Point, 0)
				addnode := cnode
				res = append(res, npos)
				res = append(res, addnode.cur)
				for addnode.parent != nil {
					addnode = addnode.parent
					res = append(res, addnode.cur)
				}
				cost := time.Since(st)
				if cost > time.Millisecond {
					filelog.INFO("astar", "cost:", cost, " from:", from, " to:", to, " open:", len(openlist))
				}
				var tmpres = make([]Point, 0)
				for i := len(res) - 1; i >= 0; i-- { //倒序
					tmpres = append(tmpres, res[i])
				}
				filelog.DEBUG("astar", " find path return 333")
				return tmpres
			}
			///检查openlist
			h := npos.CalcDistanceH(to)
			g := cnode.g + 1
			f := h + g
			if add.pathid == curpathid && f < add.f {
				add.f = f
				add.g = g
				add.h = h
				add.parent = cnode
			} else if add.pathid != curpathid {
				add.f = f
				add.g = g
				add.h = h
				add.parent = cnode
				add.pathid = curpathid
				add.cl = 1
				openlist = append(openlist, add)
			}
		}
	}

	cost := time.Since(st)
	if cost > 2*time.Millisecond {
		filelog.INFO("astar", "full no path cost:", cost, " from:", from, " to:", to, " open:", len(openlist))
	}
	maxh := 99999999
	var cnode *pathnode = nil
	for _, a := range tmpcloselist {
		if maxh > a.h && !a.cur.Equal(from) { //不等于起点的h最小的点
			maxh = a.h
			cnode = a
		}
	}

	if cnode == nil { ///取出最小F节点
		filelog.DEBUG("astar", " find path return 444")
		return nil
	}

	res := make([]Point, 0)
	addnode := cnode
	res = append(res, addnode.cur)
	for addnode.parent != nil {
		addnode = addnode.parent
		res = append(res, addnode.cur)
	}

	var tmpres = make([]Point, 0)
	for i := len(res) - 1; i >= 0; i-- { //倒序
		tmpres = append(tmpres, res[i])
	}
	filelog.DEBUG("astar", " find path return 555")
	return tmpres

	//	return nil
}
