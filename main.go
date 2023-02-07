package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type Cell int

const UNKNOWN = 0
const PAINTED = 1
const CLEAR = 2

type Coordinate struct {
	Row int
	Col int
}

func NilCoordinate() Coordinate {
	return Coordinate{-1, -1}
}

func (c Coordinate) IsNil() bool {
	return c.Row == -1 && c.Col == -1
}

func (c Coordinate) String() string {
	return fmt.Sprintf("(r%d, c%d)", c.Row, c.Col)
}

func (c Coordinate) Plus(dr int, dc int) Coordinate {
	return Coordinate{c.Row + dr, c.Col + dc}
}

func (c Coordinate) ManhattanDistance(target Coordinate) int {
	return int(math.Abs(float64(target.Row-c.Row)) + math.Abs(float64(target.Col-c.Col)))
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

func (s *CoordinateSet) IsEmpty() bool {
	return len(s.Map) == 0
}

func (s *CoordinateSet) Add(c Coordinate) {
	s.Map[c] = true
}

func (s *CoordinateSet) Del(c Coordinate) {
	delete(s.Map, c)
}

func (s *CoordinateSet) RemoveAll(other *CoordinateSet) {
	for v := range other.Map {
		s.Del(v)
	}
}

func (s CoordinateSet) Plus(other *CoordinateSet) *CoordinateSet {
	cs := s.Copy()
	for v := range other.Map {
		cs.Add(v)
	}
	return cs
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

func (s *CoordinateSet) Borders(c Coordinate) bool {
	if s.Contains(c) {
		return false
	}
	for k := range s.Map {
		if AreAdjacent(k, c) {
			return true
		}
	}
	return false
}

func (superset *CoordinateSet) ContainsAll(subset *CoordinateSet) bool {
	for tester := range subset.Map {
		if !superset.Contains(tester) {
			return false
		}
	}
	return true
}

func (s *CoordinateSet) CanAddWall(c Coordinate) bool {
	/**
	 * Diagram for the algorithm below:
	 *
	 * n is a slice containing flags for whether neighbors are painted.
	 * The asterisk represents c, and the digits represent indexes into n
	 * indicating whether that cell is painted.
	 *
	 *  012
	 *  3*4
	 *  567
	 *
	**/

	n := make([]bool, 8)
	idx := 0
	for dr := -1; dr < 2; dr++ {
		for dc := -1; dc < 2; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			n[idx] = s.Contains(c.Plus(dr, dc))
			idx++
		}
	}
	//Maybe the compiler is smart enough to optimize these conditionals for us;
	//just in case, I'll do it manually
	if n[1] {
		if n[0] && n[3] {
			return false
		}
		if n[2] && n[4] {
			return false
		}
	}
	if n[6] {
		if n[3] && n[5] {
			return false
		}
		if n[4] && n[7] {
			return false
		}
	}
	return true
}

func (s *CoordinateSet) String() string {
	out := ""
	for m := range s.Map {
		out += fmt.Sprintf("(r%d, c%d) ", m.Row, m.Col)
	}
	return out
}

func (s *CoordinateSet) SerializedString() string {
	slice := s.ToSlice()
	sort.Sort(CoordinateSlice(slice))
	out := fmt.Sprintf("%v", slice)
	return out
}

type CoordinateSlice []Coordinate

func (c CoordinateSlice) Len() int      { return len(c) }
func (c CoordinateSlice) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c CoordinateSlice) Less(i, j int) bool {
	return c[i].Row < c[j].Row || (c[i].Row == c[j].Row && c[i].Col < c[j].Col)
}

type CoordinateSetSet struct {
	Map map[string]bool
}

func EmptyCoordinateSetSet() *CoordinateSetSet {
	css := CoordinateSetSet{}
	css.Map = make(map[string]bool)
	return &css
}

func (css *CoordinateSetSet) Add(cs *CoordinateSet) bool {
	str := cs.SerializedString()
	if _, ok := css.Map[str]; ok {
		return false
	}
	css.Map[str] = true
	return true
}

func (css *CoordinateSetSet) Contains(cs *CoordinateSet) bool {
	str := cs.SerializedString()
	_, ok := css.Map[str]
	return ok
}

type IslandSpec struct {
	Col  int
	Row  int
	Size int
}

type Island struct {
	Members         *CoordinateSet
	CurrentSize     int
	TargetSize      int //an island with TargetSize=0 is one not joined to a numbered cell
	ReadyForBorders bool
	IslandType      int
	Root            Coordinate
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

func MakeRootedIsland(r int, c int, sz int) *Island {
	return &Island{ToCoordinateSet([]Coordinate{{r, c}}), 1, sz, sz == 1, CLEAR_ISLAND, Coordinate{r, c}}
}

func MakeUnrootedIsland(r int, c int) *Island {
	return &Island{ToCoordinateSet([]Coordinate{{r, c}}), 1, 0, false, CLEAR_ISLAND, NilCoordinate()}
}

func MakeWallIsland(r int, c int) *Island {
	return &Island{ToCoordinateSet([]Coordinate{{r, c}}), 1, 0, false, WALL_ISLAND, NilCoordinate()}
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
	Islands     []*Island
	WallIslands []*Island
	TotalMarked int
}

func NewGrid(w int, h int) [][]Cell {
	cells := make([][]Cell, h)
	for i := 0; i < h; i++ {
		cells[i] = make([]Cell, w)
	}
	return cells
}

func BoardFromDef(def ProblemDef) *Board {
	b := Board{def, NewGrid(def.Width, def.Height), make([]*Island, 0), make([]*Island, 0), 0}
	for _, spec := range b.Problem.IslandSpecs {
		b.Grid[spec.Row][spec.Col] = CLEAR
		b.TotalMarked++
		b.Islands = append(b.Islands, MakeRootedIsland(spec.Row, spec.Col, spec.Size))
	}
	return &b
}

func (b *Board) IslandAt(r int, c int) *Island {
	for _, i := range b.Islands {
		if i.Members.Contains(Coordinate{r, c}) {
			return i
		}
	}
	return nil
}

func AreAdjacent(a Coordinate, b Coordinate) bool {
	dr := math.Abs(float64(a.Row) - float64(b.Row))
	dc := math.Abs(float64(a.Col) - float64(b.Col))
	return (dr == 1 && dc == 0) || (dr == 0 && dc == 1)
	/*
		if dr > 1 || dc > 1 {
			return false
		}
		if dr > 0 && dc > 0 {
			return false
		}
		return true
	*/
}

func (i *Island) BordersCell(c Coordinate) bool {
	for m := range i.Members.Map {
		if AreAdjacent(m, c) {
			return true
		}
	}
	return false
}

func (i *Island) BordersIsland(other *Island) bool {
	for m1 := range i.Members.Map {
		for m2 := range other.Members.Map {
			if AreAdjacent(m1, m2) {
				return true
			}
		}
	}
	return false
}

func (i *Island) Absorb(other *Island) {
	i.CurrentSize += other.CurrentSize
	//TODO: If both are clear islands and have nonzero target sizes, then we've reached an incorrect state....think about how to detect that later when we use reductio/guess techniques
	//Options: (1) always detect BEFOREHAND and prevent the cell from being marked incorrectly
	//(2) Bubble up errors
	//(3) Add this condition to the consistency/could-be-correct-ness/error-freeness check (i.e., number of islands with TargetSize > 0 == len(b.Problem.IslandSpecs))

	if i.TargetSize == 0 {
		i.TargetSize = other.TargetSize
	}
	if i.Root.IsNil() {
		i.Root = other.Root
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
		newWallIslands := make([]*Island, 0)
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

func (b *Board) MarkClear(r int, c int) bool {
	if b.Grid[r][c] == CLEAR {
		return false
	}
	b.Grid[r][c] = CLEAR
	b.TotalMarked++
	b.Islands = append(b.Islands, MakeUnrootedIsland(r, c))
	b.MergeIslands()
	i := b.IslandAt(r, c)
	if i.TargetSize > 0 && i.CurrentSize == i.TargetSize {
		i.ReadyForBorders = true
	}
	return true
}

func (b *Board) MarkPainted(r int, c int) bool {
	if b.Grid[r][c] == PAINTED {
		return false
	}
	b.Grid[r][c] = PAINTED
	b.TotalMarked++
	b.WallIslands = append(b.WallIslands, MakeWallIsland(r, c))
	b.MergeWallIslands()
	return true
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
	s += fmt.Sprintf("\nTotal marked: %d\n", b.TotalMarked)
	return s
}

func (b *Board) IsInBounds(c Coordinate) bool {
	return c.Row >= 0 && c.Col >= 0 && c.Row < b.Problem.Height && c.Col < b.Problem.Width
}

func (b *Board) AreInBounds(r int, c int) bool {
	return r >= 0 && c >= 0 && r < b.Problem.Height && c < b.Problem.Width
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
	if c >= 'A' && c <= 'W' {
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

func BoardFromString(input string) *Board {
	def := DefFromString(input)
	b := Board{def, NewGrid(def.Width, def.Height), make([]*Island, 0), make([]*Island, 0), 0}
	for _, spec := range b.Problem.IslandSpecs {
		b.Grid[spec.Row][spec.Col] = CLEAR
		b.TotalMarked++
		b.Islands = append(b.Islands, MakeRootedIsland(spec.Row, spec.Col, spec.Size))
	}
	lines := make([]string, 0)
	for _, txt := range strings.Split(input, "\n") {
		txt = strings.TrimSpace(txt)
		if len(txt) > 0 {
			lines = append(lines, txt)
		}
	}
	fmt.Printf("%v\n", b)
	for ri, row := range lines {
		for ci, cell := range row {
			if cell == 'X' {
				fmt.Printf("%d %d %c %v %v PAINT\n", ri, ci, cell, cell, 'X')
				b.MarkPainted(ri, ci)
			} else if cell == '.' {
				fmt.Printf("%d %d %c %v %v CLEAR\n", ri, ci, cell, cell, '.')
				b.MarkClear(ri, ci)
			}
		}
	}
	return &b
}

func (b *Board) CellNeighbors(c Coordinate) *CoordinateSet {
	cs := EmptyCoordinateSet()
	cs.Add(c)
	return b.Neighbors(cs)
}

func (b *Board) Neighbors(c *CoordinateSet) *CoordinateSet {
	rset := EmptyCoordinateSet()
	for m := range c.Map {
		for dx := -1; dx < 2; dx += 2 {
			newCoord := m.Plus(dx, 0)
			if b.IsInBounds(newCoord) {
				rset.Add(newCoord)
			}
			newCoord = m.Plus(0, dx)
			if b.IsInBounds(newCoord) {
				rset.Add(newCoord)
			}
		}

	}
	for m := range c.Map {
		rset.Del(m)
	}
	return rset
}

func (b *Board) HasNeighborWith(c *CoordinateSet, val Cell) bool {
	for m := range c.Map {
		for dx := -1; dx < 2; dx += 2 {
			newCoord := m.Plus(dx, 0)
			if !c.Contains(newCoord) {
				if b.IsInBounds(newCoord) && b.Get(newCoord) == val {
					return true
				}
			}
			newCoord = m.Plus(0, dx)
			if !c.Contains(newCoord) {
				if b.IsInBounds(newCoord) && b.Get(newCoord) == val {
					return true
				}
			}
		}
	}
	return false
}

func (b *Board) NeighborsWith(c *CoordinateSet, val Cell) *CoordinateSet {
	rset := EmptyCoordinateSet()
	for m := range c.Map {
		for dx := -1; dx < 2; dx += 2 {
			newCoord := m.Plus(dx, 0)
			if b.IsInBounds(newCoord) && b.Get(newCoord) == val {
				rset.Add(newCoord)
			}
			newCoord = m.Plus(0, dx)
			if b.IsInBounds(newCoord) && b.Get(newCoord) == val {
				rset.Add(newCoord)
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
				var island *Island = nil
				for _, spec := range b.Problem.IslandSpecs {
					if spec.Row == r && spec.Col == c {
						island = MakeRootedIsland(r, c, spec.Size)
					}
				}
				if island == nil {
					island = MakeUnrootedIsland(r, c)
				}
				b.Islands = append(b.Islands, island)
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
			if b.AreInBounds(r+1, c+1) {
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

func (b *Board) Liberties(i *Island) *CoordinateSet {
	return b.NeighborsWith(i.Members, UNKNOWN)
}

func (b *Board) CountNumberedIslands(c *CoordinateSet) int {
	ct := 0
	for _, ispec := range b.Problem.IslandSpecs {
		if c.Contains(Coordinate{ispec.Row, ispec.Col}) {
			ct++
		}
	}
	return ct
}

func (b *Board) BordersMultipleRootedIslands(c Coordinate) bool {
	ct := 0
	for _, i := range b.Islands {
		if i.TargetSize > 0 && i.BordersCell(c) {
			ct++
			if ct > 1 {
				return true
			}
		}
	}
	return false
}

func (b *Board) BorderingIslands(c Coordinate) []*Island {
	res := make([]*Island, 0)
	for _, i := range b.Islands {
		if i.BordersCell(c) {
			res = append(res, i)
		}
	}
	return res
}

func TryParseFile(f string) {
	data, err := os.ReadFile(f)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	/*
		prob := DefFromString(string(data))
		fmt.Printf("problem:\n%s\n", prob)
		board := BoardFromDef(prob)
	*/
	board := BoardFromString(string(data))
	start := time.Now()
	fmt.Printf("Started solving: %s\n", time.Now())
	fmt.Printf("Initial:\n%s\n", board.StringDebug())
	board.AutoSolve()
	fmt.Printf("Board:\n%s\n", board.StringDebug())

	fmt.Printf("Finished solving: %s (duration %.4f)\n", time.Now(), float64(time.Now().UnixNano()-start.UnixNano())/1000000000.0)
}

func main() {
	//TryParseFile("problem1.txt")
	//TryParseFile("board2.txt")
	TryParseFile("problem2.txt")
	TryParseFile("problem3.txt")
}

//NEXT: reachability by islands
//NEXT: reachability for numberless islands?
//TEST THIS: all liberties of island one short of completion border the same unknown cell? e.g., the corner away from a cornered 2?
//NEXT: what about islands separating the grid (i.e., diagonally adjacent line on clear cells from edge to edge)?
//DEBUG: at the end of problem 2, why can't the lower-left 3 see -- because it would need to extend out from the unrooted island!!
//work on making tha thappen - switch CoordinateSet with an actual island?
//why doesn't the 5 island fill in at least the cell north of it?
//TODO: get the unrooted islands to expand outwards
