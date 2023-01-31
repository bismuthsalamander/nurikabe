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

type IslandSpec struct {
	Col  int
	Row  int
	Size int
}

type Island struct {
	Members      []Coordinate
	CurrentSize  int
	TargetSize   int //an island with TargetSize=0 is one not joined to a numbered cell
	BordersAdded bool
	IslandType   int
}

const CLEAR_ISLAND = 0
const WALL_ISLAND = 1

func MakeIsland(r int, c int, sz int) Island {
	return Island{[]Coordinate{{r, c}}, 1, sz, false, CLEAR_ISLAND}
}

func MakeWallIsland(r int, c int) Island {
	return Island{[]Coordinate{{r, c}}, 1, 0, false, WALL_ISLAND}
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
	for _, m := range i.Members {
		if AreAdjacent(m.Row, m.Col, r, c) {
			return true
		}
	}
	return false
}

func (i Island) BordersIsland(other Island) bool {
	for _, m1 := range i.Members {
		for _, m2 := range other.Members {
			if AreAdjacent(m1.Row, m1.Col, m2.Row, m2.Col) {
				return true
			}
		}
	}
	return false
}

func (i Island) Contains(c Coordinate) bool {
	return contains(i.Members, c)
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
	i.Members = append(i.Members, other.Members...)
}

func (b *Board) MergeAll() {
	b.MergeIslands()
	b.MergeWallIslands()
}

func (b *Board) MergeIslands() {
	changed := true
	for changed {
		changed = false
		newIslands := make([]Island, 0)
		for i := 0; i < len(b.Islands); i++ {
			for j := i + 1; j < len(b.Islands); j++ {
				if b.Islands[i].BordersIsland(b.Islands[j]) {
					changed = true
					b.Islands[i].Absorb(b.Islands[j])
					b.Islands = append(b.Islands[:j], b.Islands[j+1:]...)
				}
			}
			newIslands = append(newIslands, b.Islands[i])
		}
		b.Islands = newIslands
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

// eventually this func will include tracking "wall islands" or WallGroups
func (b *Board) MarkPainted(r int, c int) {
	if b.Grid[r][c] == PAINTED {
		return
	}
	b.Grid[r][c] = PAINTED
	b.WallIslands = append(b.WallIslands, MakeWallIsland(r, c))
	b.MergeWallIslands()
}

func (b Board) CharAt(r int, c int) string {
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
		//return string([]rune{0x81})
		return "."
	}
	return "?"
}

func (b Board) String() string {
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

func (b Board) Get(c Coordinate) Cell {
	return b.Grid[c.Row][c.Col]
}

func (b Board) StringDebug() string {
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
	if s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
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

func (b Board) CellNeighbors(c Coordinate) []Coordinate {
	rset := make([]Coordinate, 0, 4)
	if c.Row-1 >= 0 {
		rset = append(rset, Coordinate{c.Row - 1, c.Col})
	}
	if c.Row+1 < b.Problem.Height {
		rset = append(rset, Coordinate{c.Row + 1, c.Col})
	}
	if c.Col-1 >= 0 {
		rset = append(rset, Coordinate{c.Row, c.Col - 1})
	}
	if c.Col+1 < b.Problem.Width {
		rset = append(rset, Coordinate{c.Row, c.Col + 1})
	}
	return rset
}

func (b Board) Neighbors(i Island) []Coordinate {
	rset := make(map[Coordinate]bool)
	for _, m := range i.Members {
		if m.Row-1 >= 0 {
			rset[Coordinate{m.Row - 1, m.Col}] = true
		}
		if m.Row+1 < b.Problem.Height {
			rset[Coordinate{m.Row + 1, m.Col}] = true
		}
		if m.Col-1 >= 0 {
			rset[Coordinate{m.Row, m.Col - 1}] = true
		}
		if m.Col+1 < b.Problem.Width {
			rset[Coordinate{m.Row, m.Col + 1}] = true
		}
	}
	for _, m := range i.Members {
		delete(rset, Coordinate{m.Row, m.Col})
	}
	res := make([]Coordinate, len(rset))
	idx := 0
	for k := range rset {
		res[idx] = k
		idx++
	}
	return res
}

// TODO: it's probably more efficient to copy and paste Neighbors() and add the
// emptiness checks inline.
func (b Board) Liberties(i Island) []Coordinate {
	n := b.Neighbors(i)
	for idx := 0; idx < len(n); idx++ {
		if b.Grid[n[idx].Row][n[idx].Col] != UNKNOWN {
			n = append(n[:idx], n[idx+1:]...)
			idx--
		}
	}
	return n
}

func (b *Board) AddIslandBorders() {
	for _, island := range b.Islands {
		if !island.BordersAdded && island.CurrentSize == island.TargetSize {
			targets := b.Neighbors(island)
			for _, coord := range targets {
				b.MarkPainted(coord.Row, coord.Col)
			}
			island.BordersAdded = true
		}
	}
}

// TODO: liberty data structure? running slices?
func (b *Board) ExpandIslands() {
	changed := true
	for changed {
		changed = false
		for _, island := range b.Islands {
			if island.CurrentSize == island.TargetSize {
				continue
			}
			lib := b.Liberties(island)
			if len(lib) == 1 {
				b.MarkClear(lib[0].Row, lib[0].Col)
				changed = true
			}
		}
	}
}

// TODO: liberty data structure? running slices?
func (b *Board) ExpandWallIslands() {
	changed := true
	for changed {
		changed = false
		for _, island := range b.WallIslands {
			if island.CurrentSize == b.Problem.TargetWallCount {
				continue
			}
			lib := b.Liberties(island)
			if len(lib) == 1 {
				b.MarkPainted(lib[0].Row, lib[0].Col)
				changed = true
			}
		}
	}
}

// would it be faster to do this the other way around?
// have each island count their liberties and identify each cell that is a
// liberty to two different islands?
func (b *Board) PaintTwoBorderedCells() {
	for ri, row := range b.Grid {
		for ci, col := range row {
			if col != UNKNOWN {
				continue
			}
			borderCount := 0
			for _, i := range b.Islands {
				if i.BordersCell(ri, ci) {
					borderCount++
					if borderCount > 1 {
						break
					}
				}
			}
			if borderCount > 1 {
				b.MarkPainted(ri, ci)
			}
		}
	}
}

func contains(haystack []Coordinate, needle Coordinate) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}

func (b *Board) WallDfsRec(c chan []Coordinate, pathToHere []Coordinate, island *Island) {
	neighbors := b.CellNeighbors(pathToHere[len(pathToHere)-1])
	for _, n := range neighbors {
		if contains(pathToHere, n) || island.Contains(n) {
			continue
		}
		pathToHere = append(pathToHere, n)
		if b.Get(n) == PAINTED {
			out := make([]Coordinate, len(pathToHere))
			copy(out, pathToHere)
			c <- out
		} else if b.Get(n) == UNKNOWN {
			b.WallDfsRec(c, pathToHere, island)
		}
		pathToHere = pathToHere[:len(pathToHere)-1]
	}
}

func (b *Board) WallDfs(c chan []Coordinate, island *Island) {
	p := make([]Coordinate, 0, 1)
	for _, m := range island.Members {
		p = append(p, m)
		b.WallDfsRec(c, p, island)
		p = p[:0]
	}
	close(c)
}

func (b *Board) ExtendWallIslands() {
	for _, wi := range b.WallIslands {
		if wi.CurrentSize == b.Problem.TargetWallCount {
			continue
		}
		c := make(chan []Coordinate)
		go b.WallDfs(c, &wi)
		necessary := <-c
		//trim off first and last, because origin and destination are already painted
		necessary = necessary[1 : len(necessary)-1]
		for path := range c {
			for i := 0; i < len(necessary); i++ {
				if !contains(path, necessary[i]) {
					necessary = append(necessary[0:i], necessary[i+1:]...)
					i--
				}
			}
		}
		for _, target := range necessary {
			b.MarkPainted(target.Row, target.Col)
		}
	}
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
	board.ExpandIslands()
	board.ExpandWallIslands()
	board.AddIslandBorders()
	board.ExpandIslands()
	board.ExpandWallIslands()
	board.AddIslandBorders()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExtendWallIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExtendWallIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExpandIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExtendWallIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.ExpandIslands()
	fmt.Printf("board:\n%s\n", board.StringDebug())
	board.AddIslandBorders()
	fmt.Printf("board:\n%s\n", board.StringDebug())
}

func main() {
	fmt.Printf("Hello, world!")
	TryParseFile("problem1.txt")
}

//NEXT: reachability by islands
//NEXT: same thing for numberless islands?
//NEXT: all liberties of island one short of completion border the same unknown cell? e.g., the corner away from a cornered 2?
//NEXT: fill elbows
