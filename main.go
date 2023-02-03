package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

type Cell int

const UNKNOWN = 0
const PAINTED = 1
const CLEAR = 2

type Coordinate struct {
	Row int
	Col int
}

func (c Coordinate) String() string {
	return fmt.Sprintf("(r%d, c%d)", c.Row, c.Col)
}

func (c Coordinate) Plus(dr int, dc int) Coordinate {
	return Coordinate{c.Row + dr, c.Col + dc}
}

// TODO: make this a generic?
type CoordinateSet struct {
	Map map[Coordinate]bool
}

func EmptyCoordinateSet() *CoordinateSet {
	return &CoordinateSet{make(map[Coordinate]bool)}
}

func ToCoordinateSet(members []Coordinate) *CoordinateSet {
	cs := CoordinateSet{make(map[Coordinate]bool)}
	for _, c := range members {
		cs.Add(c)
	}
	return &cs
}

func (s *CoordinateSet) Size() int {
	return len(s.Map)
}

func (s *CoordinateSet) Add(c Coordinate) {
	s.Map[c] = true
}

func (s *CoordinateSet) Del(c Coordinate) {
	delete(s.Map, c)
}

func (s CoordinateSet) Plus(other *CoordinateSet) *CoordinateSet {
	for v := range other.Map {
		s.Add(v)
	}
	return &s
}

func (s *CoordinateSet) Contains(c Coordinate) bool {
	if val, ok := s.Map[c]; val && ok {
		return true
	}
	return false
}

func (s *CoordinateSet) ToSlice() []Coordinate {
	out := make([]Coordinate, 0, len(s.Map))
	for k := range s.Map {
		out = append(out, k)
	}
	return out
}

func (s *CoordinateSet) Copy() *CoordinateSet {
	cs := CoordinateSet{make(map[Coordinate]bool)}
	for k := range s.Map {
		cs.Add(k)
	}
	return &cs
}

func (s *CoordinateSet) String() string {
	out := ""
	for m := range s.Map {
		out += fmt.Sprintf("(r%d, c%d) ", m.Row, m.Col)
	}
	return out
}

type IslandSpec struct {
	Col  int
	Row  int
	Size int
}

type Island struct {
	Members      *CoordinateSet
	CurrentSize  int
	TargetSize   int //an island with TargetSize=0 is one not joined to a numbered cell
	BordersAdded bool
	IslandType   int
}

func (i *Island) String() string {
	if i.IslandType == CLEAR_ISLAND {
		return i.Members.String() + fmt.Sprintf(" %d/%d", i.CurrentSize, i.TargetSize)
	} else {
		return i.Members.String() + fmt.Sprintf(" %d", i.CurrentSize)
	}
}

func (i *Island) Contains(c Coordinate) bool {
	return i.Members.Contains(c)
}

const CLEAR_ISLAND = 0
const WALL_ISLAND = 1

func MakeIsland(r int, c int, sz int) Island {
	return Island{ToCoordinateSet([]Coordinate{{r, c}}), 1, sz, false, CLEAR_ISLAND}
}

func MakeWallIsland(r int, c int) Island {
	return Island{ToCoordinateSet([]Coordinate{{r, c}}), 1, 0, false, WALL_ISLAND}
}

type ProblemDef struct {
	Width           int
	Height          int
	IslandSpecs     []IslandSpec
	TargetWallCount int
}

func (p ProblemDef) String() string {
	s := make([]string, 0)
	for ri := 0; ri < p.Height; ri++ {
		s = append(s, strings.Repeat("_", p.Width))
	}
	for _, spec := range p.IslandSpecs {
		s[spec.Row] = s[spec.Row][:spec.Col] + islandSpecChar(spec.Size) + s[spec.Row][spec.Col+1:]
	}
	return strings.Join(s, "\n")
}

type Board struct {
	Problem     ProblemDef
	Grid        [][]Cell
	Islands     []Island
	WallIslands []Island
}

func NewGrid(w int, h int) [][]Cell {
	cells := make([][]Cell, h)
	for i := 0; i < h; i++ {
		cells[i] = make([]Cell, w)
	}
	return cells
}

func BoardFromDef(def ProblemDef) Board {
	b := Board{def, NewGrid(def.Width, def.Height), make([]Island, 0), make([]Island, 0)}
	for _, spec := range b.Problem.IslandSpecs {
		b.Grid[spec.Row][spec.Col] = CLEAR
		b.Islands = append(b.Islands, MakeIsland(spec.Row, spec.Col, spec.Size))
	}
	return b
}

func AreAdjacent(r1 int, c1 int, r2 int, c2 int) bool {
	dr := math.Abs(float64(r1) - float64(r2))
	dc := math.Abs(float64(c1) - float64(c2))
	if dr > 1 || dc > 1 {
		return false
	}
	if dr > 0 && dc > 0 {
		return false
	}
	return true
}

func (i Island) BordersCell(r int, c int) bool {
	for m := range i.Members.Map {
		if AreAdjacent(m.Row, m.Col, r, c) {
			return true
		}
	}
	return false
}

func (i Island) BordersIsland(other Island) bool {
	for m1 := range i.Members.Map {
		for m2 := range other.Members.Map {
			if AreAdjacent(m1.Row, m1.Col, m2.Row, m2.Col) {
				return true
			}
		}
	}
	return false
}

func (i *Island) Absorb(other Island) {
	i.CurrentSize += other.CurrentSize
	//TODO: If both are clear islands and have nonzero target sizes, then we've reached an incorrect state....think about how to detect that later when we use reductio/guess techniques
	//Options: (1) always detect BEFOREHAND and prevent the cell from being marked incorrectly
	//(2) Bubble up errors
	//(3) Add this condition to the consistency/could-be-correct-ness/error-freeness check (i.e., number of islands with TargetSize > 0 == len(b.Problem.IslandSpecs))

	if i.TargetSize == 0 {
		i.TargetSize = other.TargetSize
	}
	cs := i.Members.Plus(other.Members)
	i.Members = cs
}

func (b *Board) MergeAll() {
	b.MergeIslands()
	b.MergeWallIslands()
}

func (b *Board) MergeIslands() {
	changed := true
	for changed {
		changed = false
		for i := 0; i < len(b.Islands); i++ {
			for j := i + 1; j < len(b.Islands); j++ {
				if b.Islands[i].BordersIsland(b.Islands[j]) {
					changed = true
					b.Islands[i].Absorb(b.Islands[j])
					b.Islands = append(b.Islands[:j], b.Islands[j+1:]...)
				}
			}
		}
	}
}

func (b *Board) MergeWallIslands() {
	changed := true
	for changed {
		changed = false
		newWallIslands := make([]Island, 0)
		for i := 0; i < len(b.WallIslands); i++ {
			for j := i + 1; j < len(b.WallIslands); j++ {
				if b.WallIslands[i].BordersIsland(b.WallIslands[j]) {
					changed = true
					b.WallIslands[i].Absorb(b.WallIslands[j])
					b.WallIslands = append(b.WallIslands[:j], b.WallIslands[j+1:]...)
				}
			}
			newWallIslands = append(newWallIslands, b.WallIslands[i])
		}
		b.WallIslands = newWallIslands
	}
}

func (b *Board) MarkClear(r int, c int) {
	if b.Grid[r][c] == CLEAR {
		return
	}
	b.Grid[r][c] = CLEAR
	b.Islands = append(b.Islands, MakeIsland(r, c, 0))
	b.MergeIslands()
}

func (b *Board) MarkPainted(r int, c int) {
	if b.Grid[r][c] == PAINTED {
		return
	}
	b.Grid[r][c] = PAINTED
	b.WallIslands = append(b.WallIslands, MakeWallIsland(r, c))
	b.MergeWallIslands()
}

func (b *Board) CharAt(r int, c int) string {
	switch b.Grid[r][c] {
	case UNKNOWN:
		return "_"
	case PAINTED:
		return "X"
	case CLEAR:
		for _, spec := range b.Problem.IslandSpecs {
			if spec.Col == c && spec.Row == r {
				return islandSpecChar(spec.Size)
			}
		}
		return "."
	}
	return "?"
}

func (b *Board) String() string {
	s := ""
	for ri, row := range b.Grid {
		for ci := range row {
			s += b.CharAt(ri, ci)
		}
		if ri != b.Problem.Height-1 {
			s += "\n"
		}
	}
	return s
}

func (b *Board) Get(c Coordinate) Cell {
	return b.Grid[c.Row][c.Col]
}

func (b *Board) StringDebug() string {
	s := b.String() + "\n"
	if len(b.Islands) > 0 {
		s += "Islands:\n"
		for _, island := range b.Islands {
			s += fmt.Sprintf("%v\n", island)
		}
	}
	if len(b.WallIslands) > 0 {
		s += "Wall islands:\n"
		for _, island := range b.WallIslands {
			s += fmt.Sprintf("%v\n", island)
		}
	}
	solved, err := b.IsSolved()
	s += fmt.Sprintf("Solved: %v", solved)
	if err != nil {
		s += fmt.Sprintf(" (reason: %v)", err)
	}
	return s
}

func islandSpecChar(sz int) string {
	if sz < 10 {
		return string(rune(sz + '0'))
	}
	if sz < 36 {
		return string(rune((sz - 10) + 'a'))
	}
	if sz < 62 {
		return string(rune((sz - 36) + 'A'))
	}
	return "?"
}

func parseIslandSpecChar(c rune) int {
	if c >= '1' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'z' {
		return int((c - 'a') + 10)
	}
	if c >= 'A' && c <= 'Z' {
		return int((c - 'A') + 36)
	}
	return -1
}

func DefFromString(input string) ProblemDef {
	prob := ProblemDef{}
	lines := make([]string, 0)
	for _, txt := range strings.Split(input, "\n") {
		txt = strings.TrimSpace(txt)
		if len(txt) > 0 {
			lines = append(lines, txt)
		}
	}
	prob.Width = len(lines[0])
	prob.Height = len(lines)
	for ri, row := range lines {
		for ci, cell := range row {
			count := parseIslandSpecChar(cell)
			if count > -1 {
				prob.IslandSpecs = append(prob.IslandSpecs, IslandSpec{ci, ri, count})
				prob.TargetWallCount += count
			}
		}
	}
	return prob
}

func (b *Board) CellNeighbors(c Coordinate) *CoordinateSet {
	rset := EmptyCoordinateSet()
	if c.Row-1 >= 0 {
		rset.Add(c.Plus(-1, 0))
	}
	if c.Row+1 < b.Problem.Height {
		rset.Add(c.Plus(1, 0))
	}
	if c.Col-1 >= 0 {
		rset.Add(c.Plus(0, -1))
	}
	if c.Col+1 < b.Problem.Width {
		rset.Add(c.Plus(0, 1))
	}
	return rset
}

func (b *Board) Neighbors(c *CoordinateSet) *CoordinateSet {
	rset := EmptyCoordinateSet()
	for m := range c.Map {
		if m.Row-1 >= 0 {
			rset.Add(m.Plus(-1, 0))
		}
		if m.Row+1 < b.Problem.Height {
			rset.Add(m.Plus(1, 0))
		}
		if m.Col-1 >= 0 {
			rset.Add(m.Plus(0, -1))
		}
		if m.Col+1 < b.Problem.Width {
			rset.Add(m.Plus(0, 1))
		}
	}
	for m := range c.Map {
		rset.Del(m)
	}
	return rset
}

func (b *Board) NeighborsWith(c *CoordinateSet, val Cell) *CoordinateSet {
	rset := EmptyCoordinateSet()
	for m := range c.Map {
		for dx := -1; dx < 2; dx += 2 {
			if m.Row+dx >= 0 && b.Grid[m.Row+dx][m.Col] == val {
				rset.Add(m.Plus(dx, 0))
			}
			if m.Col+dx >= 0 && b.Grid[m.Row][m.Col+dx] == val {
				rset.Add(m.Plus(0, dx))
			}
		}
	}
	for m := range c.Map {
		rset.Del(m)
	}
	return rset
}

func (b *Board) RebuildIslands() {
	b.Islands = b.Islands[:0]
	b.WallIslands = b.WallIslands[:0]
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			switch b.Grid[r][c] {
			case PAINTED:
				b.WallIslands = append(b.WallIslands, MakeWallIsland(r, c))
			case CLEAR:
				tSize := 0
				for _, spec := range b.Problem.IslandSpecs {
					if spec.Row == r && spec.Col == c {
						tSize = spec.Size
					}
				}
				b.Islands = append(b.Islands, MakeIsland(r, c, tSize))
			}
		}
	}
	b.MergeAll()
}

func (b *Board) IsSolved() (bool, error) {
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			if b.Grid[r][c] == UNKNOWN {
				return false, fmt.Errorf("cell %v is unknown", Coordinate{r, c})
			}
			if r < b.Problem.Height-1 && c < b.Problem.Width-1 {
				painted := 0
				for dr := 0; dr < 2; dr++ {
					for dc := 0; dc < 2; dc++ {
						if b.Grid[r+dr][c+dc] == PAINTED {
							painted++
						}
					}
				}
				if painted == 4 {
					return false, fmt.Errorf("two-by-two black square at %v", Coordinate{r, c})
				}
			}
		}
	}
	b.RebuildIslands()
	if len(b.WallIslands) > 1 {
		return false, fmt.Errorf("walls are not all joined")
	}
	for _, i := range b.Islands {
		if i.CurrentSize != i.TargetSize {
			coord := Coordinate{}
			for k := range i.Members.Map {
				coord = k
				break
			}
			return false, fmt.Errorf("island at %v has size %d (should be %d)", coord, i.CurrentSize, i.TargetSize)
		}
	}
	return true, nil
}

func (b *Board) Liberties(i Island) *CoordinateSet {
	return b.NeighborsWith(i.Members, UNKNOWN)
}

func (b *Board) CountIslands(c *CoordinateSet) int {
	ct := 0
	for _, ispec := range b.Problem.IslandSpecs {
		if c.Contains(Coordinate{ispec.Row, ispec.Col}) {
			ct++
		}
	}
	return ct
}

func TryParseFile(f string) {
	data, err := os.ReadFile(f)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	prob := DefFromString(string(data))
	fmt.Printf("problem:\n%s\n", prob)
	board := BoardFromDef(prob)
	fmt.Printf("Initial:\n%s\n", board.StringDebug())
	board.PaintTwoBorderedCells()
	fmt.Printf("After painting two-bordered:\n%s\n", board.StringDebug())
	board.AddIslandBorders()
	fmt.Printf("After adding borders:\n%s\n", board.StringDebug())
	board.AddIslandBorders()
	fmt.Printf("After adding borders again:\n%s\n", board.StringDebug())
	board.ExpandIslandsOneLiberty()
	fmt.Printf("After expanding islands one liberty:\n%s\n", board.StringDebug())
	board.ExpandWallIslands()
	fmt.Printf("After expanding wall islands:\n%s\n", board.StringDebug())
	board.AddIslandBorders()
	fmt.Printf("After adding more borders:\n%s\n", board.StringDebug())
	board.ExpandIslandsOneLiberty()
	fmt.Printf("After expanding wall isalnds on eliberty:\n%s\n", board.StringDebug())
	board.ExpandWallIslands()
	fmt.Printf("After expanding wall islands:\n%s\n", board.StringDebug())
	board.AddIslandBorders()
	fmt.Printf("after borders board:\n%s\n", board.StringDebug())
	board.ExtendWallIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExtendWallIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExpandIslandsOneLiberty()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExtendWallIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExpandIslandsOneLiberty()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.AddIslandBorders()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExtendIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.MarkPainted(7, 2)
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.FillElbows()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.AddIslandBorders()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.FillElbows()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExtendIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.AddIslandBorders()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.RebuildIslands()
	fmt.Printf("board after rebuild:\n%s\n", board.StringDebug())
}

func main() {
	fmt.Printf("Hello, world!")
	TryParseFile("problem1.txt")
}

//NEXT: reachability by islands
//NEXT: reachability for numberless islands?
//TEST THIS: all liberties of island one short of completion border the same unknown cell? e.g., the corner away from a cornered 2?
//NEXT: what about islands separating the grid (i.e., diagonally adjacent line on clear cells from edge to edge)?
