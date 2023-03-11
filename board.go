package main

import (
	"fmt"
	"math"
	"os"
	"sort"
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

func NilCoordinate() Coordinate {
	return Coordinate{-1, -1}
}

func (c Coordinate) IsNil() bool {
	return c.Row == -1 && c.Col == -1
}

func (c Coordinate) String() string {
	return fmt.Sprintf("(r%d, c%d)", c.Row, c.Col)
}

func (c Coordinate) Translate(dr int, dc int) Coordinate {
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

func EmptyCoordinateSetSz(sz int) *CoordinateSet {
	return &CoordinateSet{make(map[Coordinate]bool, sz)}
}

func SingleCoordinateSet(c Coordinate) *CoordinateSet {
	cs := EmptyCoordinateSet()
	cs.Add(c)
	return cs
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

func (s *CoordinateSet) DelAll(other *CoordinateSet) {
	for v := range other.Map {
		s.Del(v)
	}
}

func (cs *CoordinateSet) AddAll(other *CoordinateSet) {
	for v := range other.Map {
		cs.Add(v)
	}
}

func (s *CoordinateSet) Plus(other *CoordinateSet) *CoordinateSet {
	cs := s.Copy()
	for v := range other.Map {
		cs.Add(v)
	}
	return cs
}

func (s *CoordinateSet) Minus(other *CoordinateSet) *CoordinateSet {
	cs := s.Copy()
	for v := range other.Map {
		cs.Del(v)
	}
	return cs
}

func (s *CoordinateSet) Contains(c Coordinate) bool {
	if val, ok := s.Map[c]; val && ok {
		return true
	}
	return false
}

func (s *CoordinateSet) ContainsAtLeastOne(other *CoordinateSet) bool {
	for k := range other.Map {
		if s.Contains(k) {
			return true
		}
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
	cs := CoordinateSet{make(map[Coordinate]bool, s.Size())}
	for k := range s.Map {
		cs.Add(k)
	}
	return &cs
}

func (s *CoordinateSet) OneMember() Coordinate {
	for k := range s.Map {
		return k
	}
	return NilCoordinate()
}

func (s *CoordinateSet) BordersCoordinate(c Coordinate) bool {
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

func (cs *CoordinateSet) Equals(other *CoordinateSet) bool {
	if cs.Size() != other.Size() {
		return false
	}
	for k := range cs.Map {
		if !other.Contains(k) {
			return false
		}
	}
	return true
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
			n[idx] = s.Contains(c.Translate(dr, dc))
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
	Possibilities   []*CoordinateSet
	Reachable       *CoordinateSet
}

func (i *Island) Clone() *Island {
	new := Island{
		i.Members.Copy(),
		i.CurrentSize,
		i.TargetSize,
		i.ReadyForBorders,
		i.IslandType,
		i.Root,
		nil,
		nil,
	}
	if new.IslandType == CLEAR_ISLAND {
		//we can just copy the pointers because a possibility is never modified once it's in place.
		new.Possibilities = make([]*CoordinateSet, len(i.Possibilities))
		copy(new.Possibilities, i.Possibilities)
		i.Reachable = i.Reachable.Copy()
	}
	return &new
}

func (i *Island) String() string {
	if i.IslandType == WALL_ISLAND {
		return i.Members.String() + fmt.Sprintf(" %d", i.CurrentSize)
	}
	if i.IsComplete() {
		return i.Members.String() + fmt.Sprintf(" %d/%d", i.CurrentSize, i.TargetSize)
	}
	out := i.Members.String() + fmt.Sprintf(" %d/%d poss %d", i.CurrentSize, i.TargetSize, len(i.Possibilities))
	if len(i.Possibilities) >= 1 {
		return i.Members.String() + fmt.Sprintf(" %d/%d poss %d", i.CurrentSize, i.TargetSize, len(i.Possibilities))
	}
	out = out + " "
	for idx, p := range i.Possibilities {
		out += fmt.Sprintf("%d: %v\n", idx, p.Minus(i.Members))
	}
	return out
}

func (i *Island) Contains(c Coordinate) bool {
	return i.Members.Contains(c)
}

func (i *Island) IsRooted() bool {
	return !i.Root.IsNil()
}

func (i *Island) IsComplete() bool {
	return i.TargetSize == i.CurrentSize
}

const CLEAR_ISLAND = 0
const WALL_ISLAND = 1

func MakeRootedIsland(r int, c int, sz int) *Island {
	return &Island{SingleCoordinateSet(Coordinate{r, c}), 1, sz, sz == 1, CLEAR_ISLAND, Coordinate{r, c}, make([]*CoordinateSet, 0, 10), EmptyCoordinateSet()}
}

func MakeUnrootedIsland(r int, c int) *Island {
	return &Island{SingleCoordinateSet(Coordinate{r, c}), 1, 0, false, CLEAR_ISLAND, NilCoordinate(), make([]*CoordinateSet, 0, 10), EmptyCoordinateSet()}
}

func MakeWallIsland(r int, c int) *Island {
	return &Island{SingleCoordinateSet(Coordinate{r, c}), 1, 0, false, WALL_ISLAND, NilCoordinate(), nil, nil}
}

type ProblemDef struct {
	Width           int
	Height          int
	Size            int
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
	Problem      ProblemDef
	Grid         [][]Cell
	ScratchGrid  [][]Cell
	Islands      []*Island
	WallIslands  []*Island
	DiagonalSets []*CoordinateSet
	TotalMarked  int
}

func NewGrid(w int, h int) [][]Cell {
	cells := make([][]Cell, h)
	for i := 0; i < h; i++ {
		cells[i] = make([]Cell, w)
	}
	return cells
}

func BoardFromDef(def ProblemDef) *Board {
	b := Board{def, NewGrid(def.Width, def.Height), NewGrid(def.Width, def.Height), make([]*Island, 0), make([]*Island, 0), make([]*CoordinateSet, 0), 0}
	for _, spec := range b.Problem.IslandSpecs {
		b.Grid[spec.Row][spec.Col] = CLEAR
		b.TotalMarked++
		b.Islands = append(b.Islands, MakeRootedIsland(spec.Row, spec.Col, spec.Size))
		b.DiagonalSets = append(b.DiagonalSets, SingleCoordinateSet(Coordinate{spec.Row, spec.Col}))
	}
	return &b
}

func (b *Board) ClearScratchGrid() {
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			b.ScratchGrid[r][c] = UNKNOWN
		}
	}
}

func (b *Board) Clone() *Board {
	Watch.Start("Clone board")
	defer Watch.Stop("Clone board")
	//merge the wall islands
	//new := BoardFromDef(b.Problem)
	new := Board{b.Problem, NewGrid(b.Problem.Width, b.Problem.Height), b.ScratchGrid, make([]*Island, 0, len(b.Islands)), make([]*Island, 0, len(b.WallIslands)), make([]*CoordinateSet, 0, len(b.DiagonalSets)), b.TotalMarked}
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			new.Grid[r][c] = b.Grid[r][c]
		}
	}
	for _, i := range b.Islands {
		new.Islands = append(new.Islands, i.Clone())
	}
	for _, i := range b.WallIslands {
		new.WallIslands = append(new.WallIslands, i.Clone())
	}
	for _, cs := range b.DiagonalSets {
		new.DiagonalSets = append(new.DiagonalSets, cs.Copy())
	}
	if new.TotalMarked != b.TotalMarked {
		fmt.Printf("me %v new %v\n", b.TotalMarked, new.TotalMarked)
		os.Exit(0)
	}
	return &new
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
}

func AreDiagonallyAdjacent(a Coordinate, b Coordinate) bool {
	dr := a.Row - b.Row
	dc := a.Col - b.Col
	return dr <= 1 && dr >= -1 && dc <= 1 && dc >= -1
}

func (i *Island) BordersCell(c Coordinate) bool {
	for m := range i.Members.Map {
		if AreAdjacent(m, c) {
			return true
		}
	}
	return false
}

func (cs *CoordinateSet) BordersSet(other *CoordinateSet) bool {
	for m1 := range cs.Map {
		for m2 := range other.Map {
			if AreAdjacent(m1, m2) {
				return true
			}
		}
	}
	return false
}

func (cs *CoordinateSet) BordersSetDiagonally(other *CoordinateSet) bool {
	for m1 := range cs.Map {
		for m2 := range other.Map {
			if AreDiagonallyAdjacent(m1, m2) {
				return true
			}
		}
	}
	return false
}

func (i *Island) BordersIsland(other *Island) bool {
	return i.Members.BordersSet(other.Members)
}

func (i *Island) BordersSet(other *CoordinateSet) bool {
	return i.Members.BordersSet(other)
}

func (i *Island) BordersSetDiagonally(other *CoordinateSet) bool {
	return i.Members.BordersSetDiagonally(other)
}

func (i *Island) Absorb(other *Island) {
	i.CurrentSize += other.CurrentSize
	//TODO: If both are clear islands and have nonzero target sizes, then we've reached an incorrect state....think about how to detect that later when we use reductio/guess techniques
	//Options: (1) always detect BEFOREHAND and prevent the cell from being marked incorrectly
	//(2) Bubble up errors
	//(3) Add this condition to the consistency/could-be-correct-ness/error-freeness check (i.e., number of islands with TargetSize > 0 == len(b.Problem.IslandSpecs))

	if i.TargetSize == 0 {
		i.TargetSize = other.TargetSize
		i.Root = other.Root
		newPossibilities := make([]*CoordinateSet, 0, len(i.Possibilities)+len(other.Possibilities))
		if len(i.Possibilities) > 0 {
			newPossibilities = append(newPossibilities, i.Possibilities...)
		}
		if len(other.Possibilities) > 0 {
			newPossibilities = append(newPossibilities, other.Possibilities...)
		}
		i.Possibilities = newPossibilities
	}
	cs := i.Members.Plus(other.Members)
	i.Members = cs
}

func (b *Board) MergeAll() {
	b.MergeIslands()
	b.MergeDiagonalSets()
	b.MergeWallIslands()
}

func (b *Board) MergeIslands() {
	Watch.Start("MergeIslands")
	defer Watch.Stop("MergeIslands")
	changed := true
	for changed {
		changed = false
		for i := 0; i < len(b.Islands); i++ {
			for j := i + 1; j < len(b.Islands); j++ {
				if b.Islands[i].BordersIsland(b.Islands[j]) {
					changed = true
					//fmt.Printf("%v is absorbing %v\n", b.Islands[i], b.Islands[j])
					b.Islands[i].Absorb(b.Islands[j])
					//fmt.Printf("%v just absorbed!\n", b.Islands[i])
					b.Islands[j] = b.Islands[len(b.Islands)-1]
					b.Islands = b.Islands[:len(b.Islands)-1]
					j--
				}
			}
		}
	}
	b.PopulateUnrootedPossibilities()
	b.StripAllPossibilities()
}

func (b *Board) MergeDiagonalSets() {
	Watch.Start("MergeDiagonalSets")
	defer Watch.Stop("MergeDiagonalSets")
	changed := true
	for changed {
		changed = false
		for i := 0; i < len(b.DiagonalSets); i++ {
			for j := i + 1; j < len(b.DiagonalSets); j++ {
				if b.DiagonalSets[i].BordersSetDiagonally(b.DiagonalSets[j]) {
					changed = true
					b.DiagonalSets[i].AddAll(b.DiagonalSets[j])
					RemoveFromSlice(&b.DiagonalSets, j)
					j--
				}
			}
		}
	}
}

func (b *Board) MergeWallIslands() {
	Watch.Start("MergeWallIslands")
	defer Watch.Stop("MergeWallIslands")
	changed := true
	for changed {
		changed = false
		newWallIslands := make([]*Island, 0)
		for i := 0; i < len(b.WallIslands); i++ {
			for j := i + 1; j < len(b.WallIslands); j++ {
				if b.WallIslands[i].BordersIsland(b.WallIslands[j]) {
					changed = true
					b.WallIslands[i].Absorb(b.WallIslands[j])
					b.WallIslands[j] = b.WallIslands[len(b.WallIslands)-1]
					b.WallIslands = b.WallIslands[:len(b.WallIslands)-1]
					j--
				}
			}
			newWallIslands = append(newWallIslands, b.WallIslands[i])
		}
		b.WallIslands = newWallIslands
	}
}

func (b *Board) MarkClear(r int, c int) bool {
	if !b.AreInBounds(r, c) {
		//TODO: error?
		return false
	}
	if b.Grid[r][c] == CLEAR {
		return false
	}
	b.Grid[r][c] = CLEAR
	b.TotalMarked++
	b.Islands = append(b.Islands, MakeUnrootedIsland(r, c))
	b.DiagonalSets = append(b.DiagonalSets, SingleCoordinateSet(Coordinate{r, c}))
	b.MergeIslands()
	b.MergeDiagonalSets()
	i := b.IslandAt(r, c)
	if i.TargetSize > 0 && i.CurrentSize == i.TargetSize {
		i.ReadyForBorders = true
	} else {
		b.StripPossibilities(i)
	}
	//Remove this possibility from all OTHER islands
	if i.IsRooted() {
		for _, o := range b.Islands {
			if !o.IsRooted() || o.Root == i.Root {
				continue
			}
			for idx := 0; idx < len(o.Possibilities); idx++ {
				if o.Possibilities[idx].Contains(Coordinate{r, c}) {
					RemoveFromSlice(&o.Possibilities, idx)
					idx--
				}
			}
		}
	}
	return true
}

func (b *Board) Mark(r int, c int, cell Cell) bool {
	if cell == UNKNOWN {
		return false
	} else if cell == PAINTED {
		return b.MarkPainted(r, c)
	} else if cell == CLEAR {
		return b.MarkClear(r, c)
	}
	return false
}

func (b *Board) MarkPainted(r int, c int) bool {
	if !b.AreInBounds(r, c) {
		//TODO: error?
		return false
	}
	if b.Grid[r][c] == PAINTED {
		return false
	}
	b.Grid[r][c] = PAINTED
	b.TotalMarked++
	b.WallIslands = append(b.WallIslands, MakeWallIsland(r, c))
	b.MergeWallIslands()
	b.RemoveFromPossibilities(Coordinate{r, c})
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
		s += "\n"
	}
	if b.TotalMarked < b.Problem.Size {
		s += fmt.Sprintf("Total marked: %d\n", b.TotalMarked)
	}
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

func (b *Board) IsOnEdge(c Coordinate) bool {
	return c.Row == 0 || c.Col == 0 || c.Row == b.Problem.Height-1 || c.Col == b.Problem.Width-1
}

func (b *Board) StringDebug() string {
	s := b.String() + "\n"
	space := 1
	if len(b.Islands) > 0 {
		s += "Islands:\n"
		for _, island := range b.Islands {
			if !island.IsComplete() {
				s += fmt.Sprintf("%v\n", island)
				if island.CurrentSize < island.TargetSize && len(island.Possibilities) > 0 {
					space *= len(island.Possibilities)
				}
			}
		}
		s += "Diagonal sets:\n"
		for _, set := range b.DiagonalSets {
			s += fmt.Sprintf("%v\n", set)
		}
	}
	solved, err := b.IsSolved()
	s += fmt.Sprintf("Solved: %v", solved)
	if err != nil {
		s += fmt.Sprintf(" (reason: %v)\n", err)
		s += fmt.Sprintf("Possible solutions: %d\n", space)
	}
	return s
}

func (b *Board) PlusMyNeighbors(c *CoordinateSet) *CoordinateSet {
	rset := c.Copy()
	for m := range c.Map {
		for dx := -1; dx < 2; dx += 2 {
			newCoord := m.Translate(dx, 0)
			if b.IsInBounds(newCoord) {
				rset.Add(newCoord)
			}
			newCoord = m.Translate(0, dx)
			if b.IsInBounds(newCoord) {
				rset.Add(newCoord)
			}
		}

	}
	return rset
}

func (b *Board) HasNeighborWith(c *CoordinateSet, val Cell) bool {
	for m := range c.Map {
		for dx := -1; dx < 2; dx += 2 {
			newCoord := m.Translate(dx, 0)
			if b.IsInBounds(newCoord) && b.Get(newCoord) == val {
				if !c.Contains(newCoord) {
					return true
				}
			}
			newCoord = m.Translate(0, dx)
			if b.IsInBounds(newCoord) && b.Get(newCoord) == val {
				if !c.Contains(newCoord) {
					return true
				}
			}
		}
	}
	return false
}

func (b *Board) NeighborsWith(c *CoordinateSet, val Cell) *CoordinateSet {
	rset := EmptyCoordinateSetSz(c.Size() * 3)
	for m := range c.Map {
		for dx := -1; dx < 2; dx += 2 {
			newCoord := m.Translate(dx, 0)
			if b.IsInBounds(newCoord) && b.Get(newCoord) == val {
				rset.Add(newCoord)
			}
			newCoord = m.Translate(0, dx)
			if b.IsInBounds(newCoord) && b.Get(newCoord) == val {
				rset.Add(newCoord)
			}
		}
	}
	rset.DelAll(c)
	return rset
}

func (b *Board) TouchesABorder(cs *CoordinateSet) bool {
	for k := range cs.Map {
		if k.Row == 0 || k.Col == 0 || k.Row == b.Problem.Height-1 || k.Col == b.Problem.Width-1 {
			return true
		}
	}
	return false
}

func (b *Board) IsPool(rTL int, cTL int) bool {
	if !b.AreInBounds(rTL+1, cTL+1) {
		return false
	}
	painted := 0
	for dr := 0; dr < 2; dr++ {
		for dc := 0; dc < 2; dc++ {
			if b.Grid[rTL+dr][cTL+dc] == PAINTED {
				painted++
			}
		}
	}
	return painted == 4
}

func (b *Board) IsSolved() (bool, error) {
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			if b.Grid[r][c] == UNKNOWN {
				return false, fmt.Errorf("cell %v is unknown", Coordinate{r, c})
			}
			if b.IsPool(r, c) {
				return false, fmt.Errorf("two-by-two pool at %v", Coordinate{r, c})
			}
		}
	}
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

func RemoveFromSlice[T Island | CoordinateSet](s *[]*T, i int) {
	oldLen := len(*s)
	(*s)[i] = (*s)[oldLen-1]
	*s = (*s)[:oldLen-1]
}
