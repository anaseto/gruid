package gorltk

import "fmt"

// Position represents an (X,Y) position in a grid.
type Position struct {
	X int
	Y int
}

func (pos Position) E() Position {
	return Position{pos.X + 1, pos.Y}
}

func (pos Position) SE() Position {
	return Position{pos.X + 1, pos.Y + 1}
}

func (pos Position) NE() Position {
	return Position{pos.X + 1, pos.Y - 1}
}

func (pos Position) N() Position {
	return Position{pos.X, pos.Y - 1}
}

func (pos Position) S() Position {
	return Position{pos.X, pos.Y + 1}
}

func (pos Position) W() Position {
	return Position{pos.X - 1, pos.Y}
}

func (pos Position) SW() Position {
	return Position{pos.X - 1, pos.Y + 1}
}

func (pos Position) NW() Position {
	return Position{pos.X - 1, pos.Y - 1}
}

func (pos Position) Distance(to Position) int {
	deltaX := abs(to.X - pos.X)
	deltaY := abs(to.Y - pos.Y)
	return deltaX + deltaY
}

func (pos Position) MaxCardinalDist(to Position) int {
	deltaX := abs(to.X - pos.X)
	deltaY := abs(to.Y - pos.Y)
	if deltaX > deltaY {
		return deltaX
	}
	return deltaY
}

func (pos Position) DistanceX(to Position) int {
	deltaX := abs(to.X - pos.X)
	return deltaX
}

func (pos Position) DistanceY(to Position) int {
	deltaY := abs(to.Y - pos.Y)
	return deltaY
}

func (pos Position) Neighbors(nb []Position, keep func(Position) bool) []Position {
	neighbors := [8]Position{pos.E(), pos.W(), pos.N(), pos.S(), pos.NE(), pos.NW(), pos.SE(), pos.SW()}
	nb = nb[:0]
	for _, npos := range neighbors {
		if keep(npos) {
			nb = append(nb, npos)
		}
	}
	return nb
}

func (pos Position) CardinalNeighbors(nb []Position, keep func(Position) bool) []Position {
	neighbors := [4]Position{pos.E(), pos.W(), pos.N(), pos.S()}
	nb = nb[:0]
	for _, npos := range neighbors {
		if keep(npos) {
			nb = append(nb, npos)
		}
	}
	return nb
}

type Direction int

const (
	NoDir Direction = iota
	E
	ENE
	NE
	NNE
	N
	NNW
	NW
	WNW
	W
	WSW
	SW
	SSW
	S
	SSE
	SE
	ESE
)

func (dir Direction) String() (s string) {
	switch dir {
	case NoDir:
		s = ""
	case E:
		s = "E"
	case ENE:
		s = "ENE"
	case NE:
		s = "NE"
	case NNE:
		s = "NNE"
	case N:
		s = "N"
	case NNW:
		s = "NNW"
	case NW:
		s = "NW"
	case WNW:
		s = "WNW"
	case W:
		s = "W"
	case WSW:
		s = "WSW"
	case SW:
		s = "SW"
	case SSW:
		s = "SSW"
	case S:
		s = "S"
	case SSE:
		s = "SSE"
	case SE:
		s = "SE"
	case ESE:
		s = "ESE"
	}
	return s
}

func (pos Position) To(dir Direction) Position {
	to := pos
	switch dir {
	case E, ENE, ESE:
		to = pos.E()
	case NE:
		to = pos.NE()
	case NNE, N, NNW:
		to = pos.N()
	case NW:
		to = pos.NW()
	case WNW, W, WSW:
		to = pos.W()
	case SW:
		to = pos.SW()
	case SSW, S, SSE:
		to = pos.S()
	case SE:
		to = pos.SE()
	}
	return to
}

func (pos Position) Dir(from Position) Direction {
	deltaX := abs(pos.X - from.X)
	deltaY := abs(pos.Y - from.Y)
	switch {
	case pos.X > from.X && pos.Y == from.Y:
		return E
	case pos.X > from.X && pos.Y < from.Y:
		switch {
		case deltaX > deltaY:
			return ENE
		case deltaX == deltaY:
			return NE
		default:
			return NNE
		}
	case pos.X == from.X && pos.Y < from.Y:
		return N
	case pos.X < from.X && pos.Y < from.Y:
		switch {
		case deltaY > deltaX:
			return NNW
		case deltaX == deltaY:
			return NW
		default:
			return WNW
		}
	case pos.X < from.X && pos.Y == from.Y:
		return W
	case pos.X < from.X && pos.Y > from.Y:
		switch {
		case deltaX > deltaY:
			return WSW
		case deltaX == deltaY:
			return SW
		default:
			return SSW
		}
	case pos.X == from.X && pos.Y > from.Y:
		return S
	case pos.X > from.X && pos.Y > from.Y:
		switch {
		case deltaY > deltaX:
			return SSE
		case deltaX == deltaY:
			return SE
		default:
			return ESE
		}
	default:
		panic(fmt.Sprintf("internal error: invalid position:%+v-%+v", pos, from))
	}
}

func (dir Direction) InViewCone(from, to Position) bool {
	if to == from {
		return true
	}
	d := to.Dir(from)
	if d == dir || from.Distance(to) <= 1 {
		return true
	}
	switch dir {
	case E:
		switch d {
		case ESE, ENE, NE, SE:
			return true
		}
	case NE:
		switch d {
		case ENE, NNE, N, E:
			return true
		}
	case N:
		switch d {
		case NNE, NNW, NE, NW:
			return true
		}
	case NW:
		switch d {
		case NNW, WNW, N, W:
			return true
		}
	case W:
		switch d {
		case WNW, WSW, NW, SW:
			return true
		}
	case SW:
		switch d {
		case WSW, SSW, W, S:
			return true
		}
	case S:
		switch d {
		case SSW, SSE, SW, SE:
			return true
		}
	case SE:
		switch d {
		case SSE, ESE, S, E:
			return true
		}
	}
	return false
}

func (dir Direction) CounterClockwise() (d Direction) {
	switch dir {
	case E:
		d = NE
	case NE:
		d = N
	case N:
		d = NW
	case NW:
		d = W
	case W:
		d = SW
	case SW:
		d = S
	case S:
		d = SE
	case SE:
		d = E
	}
	return d
}

func (dir Direction) Clockwise() (d Direction) {
	switch dir {
	case E:
		d = SE
	case NE:
		d = E
	case N:
		d = NE
	case NW:
		d = N
	case W:
		d = NW
	case SW:
		d = W
	case S:
		d = SW
	case SE:
		d = S
	}
	return d
}
