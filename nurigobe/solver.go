package nurigobe

import (
	"fmt"
	"os"
)

type ProgressUpdate struct {
	CurrentAction string
	TotalMarked   int
	GridSize      int
}

type Solver struct {
	b        *Board
	solution *Board
	Action   string
	Progress chan ProgressUpdate
}

func NewSolver(b *Board) *Solver {
    s := Solver{b, nil, "", make(chan ProgressUpdate, b.Problem.Size*2)}
    return &s
}


func (s *Solver) UpdateAction(a string) {
	s.Action = a
	s.SendProgress()
}

func (s *Solver) SendProgress() {
	if s.Progress == nil {
		return
	}
	s.Progress <- ProgressUpdate{
		s.Action,
		s.b.TotalMarked,
		s.b.Problem.Size,
	}
}

func (s *Solver) MarkPainted(r int, c int) bool {
	result := s.b.MarkPainted(r, c)
	if result {
		s.SendProgress()
	}
	return result
}

func (s *Solver) MarkClear(r int, c int) bool {
	result := s.b.MarkClear(r, c)
	if result {
		s.SendProgress()
	}
	return result
}

func (s *Solver) Mark(r int, c int, cell Cell) bool {
	result := s.b.Mark(r, c, cell)
	s.SendProgress()
	return result
}

func (s *Solver) AddIslandBorders() bool {
	s.UpdateAction("Adding island borders")
	Watch.Start("AIB")
	defer Watch.Stop("AIB")
	didChange := false
	for _, island := range s.b.Islands {
		if island.ReadyForBorders {
			targets := s.b.NeighborsWith(island.Members, UNKNOWN)
			for coord := range targets.Map {
				didChange = s.MarkPainted(coord.Row, coord.Col) || didChange
			}
			island.ReadyForBorders = false
		}
	}
	return didChange
}

// TODO: liberty data structure? running slices?
func (s *Solver) ExtendIslandsOneLiberty() bool {
	s.UpdateAction("Extending islands (1 liberty)")
	Watch.Start("EI1")
	defer Watch.Stop("EI1")
	didChange := false
	changed := true
	for changed {
		changed = false
		for _, island := range s.b.Islands {
			if island.CurrentSize == island.TargetSize {
				continue
			}
			lib := s.b.Liberties(island)
			if lib.Size() == 1 {
				c := lib.OneMember()
				result := s.MarkClear(c.Row, c.Col)
				didChange = didChange || result
				changed = changed || result
				break
			}
		}
	}
	return didChange
}

// TODO: liberty data structure? running slices?
func (s *Solver) ExtendWallIslandsOneLiberty() bool {
	s.UpdateAction("Extend wall islands (1 liberty)")
	Watch.Start("EW1")
	defer Watch.Stop("EW1")
	didChange := false
	changed := true
	for changed {
		changed = false
		for _, island := range s.b.WallIslands {
			if len(s.b.WallIslands) == 1 {
				break
			}
			if island.CurrentSize == s.b.Problem.TargetWallCount {
				return didChange
			}
			lib := s.b.Liberties(island)
			if lib.Size() == 1 {
				c := lib.OneMember()
				result := s.MarkPainted(c.Row, c.Col)
				didChange = didChange || result
				changed = changed || result
				break
			}
		}
	}
	return didChange
}

// would it be faster to do this the other way around?
// have each island count their liberties and identify each cell that is a
// liberty to two different islands?
func (s *Solver) PaintTwoBorderedCells() bool {
	s.UpdateAction("Two-bordered cells")
	Watch.Start("P2B")
	defer Watch.Stop("P2B")
	didChange := false
	for ri, row := range s.b.Grid {
		for ci, col := range row {
			if col != UNKNOWN {
				continue
			}
			borderCount := 0
			for _, i := range s.b.Islands {
				if i.TargetSize == 0 {
					continue
				}
				if i.BordersCell(Coordinate{ri, ci}) {
					borderCount++
					if borderCount > 1 {
						break
					}
				}
			}
			if borderCount > 1 {
				didChange = s.MarkPainted(ri, ci) || didChange
			}
		}
	}
	return didChange
}

func (s *Solver) WallDfs(members *CoordinateSet) *CoordinateSet {
	necessary := EmptyCoordinateSet()
	for r := 0; r < s.b.Problem.Height; r++ {
		for c := 0; c < s.b.Problem.Width; c++ {
			necessary.Add(Coordinate{r, c})
		}
	}
	necessary.DelAll(members)
	s.WallDfsRec(members, necessary)
	return necessary
}

func (s *Solver) WallDfsRec(members *CoordinateSet, necessary *CoordinateSet) {
	if necessary.IsEmpty() || members.ContainsAll(necessary) {
		return
	}
	if s.b.HasNeighborWith(members, PAINTED) {
		for nCheck := range necessary.Map {
			if !members.Contains(nCheck) {
				necessary.Del(nCheck)
			}
		}
		return
	}
	if necessary.IsEmpty() {
		return
	}
	neighbors := s.b.NeighborsWith(members, UNKNOWN)
	for n := range neighbors.Map {
		if s.b.Get(n) == UNKNOWN && members.CanAddWall(n) {
			members.Add(n)
			s.WallDfsRec(members, necessary)
			members.Del(n)
		}
	}
}

// Two possible optimizations:
// 1. Track which CoordinateSets we've already had
// 2. Do it all in one recursive function, passing necessary through and dumping the channel; skip any possibility
// with necessary as a subset of the possibility
func (s *Solver) ExtendWallIslands() bool {
	s.UpdateAction("Extend wall islands")
	Watch.Start("EWI")
	defer Watch.Stop("EWI")
	if len(s.b.WallIslands) < 2 {
		return false
	}
	didChange := false
	for _, wi := range s.b.WallIslands {
		if wi.CurrentSize == s.b.Problem.TargetWallCount {
			return didChange
		}
		necessaryMembers := s.WallDfs(wi.Members)
		for target := range necessaryMembers.Map {
			didChange = s.MarkPainted(target.Row, target.Col) || didChange
		}
	}
	return didChange
}

func (s *Solver) FillElbows() bool {
	s.UpdateAction("Fill elbows")
	Watch.Start("FEL")
	defer Watch.Stop("FEL")
	//TODO: make more efficient with overlapping columns that we save between inner loop iterations?
	didChange := false
	for r := 0; r < s.b.Problem.Height-1; r++ {
		for c := 0; c < s.b.Problem.Width-1; c++ {
			painted := 0
			clear := 0
			target := Coordinate{}
			for dr := 0; dr < 2; dr++ {
				for dc := 0; dc < 2; dc++ {
					switch s.b.Grid[r+dr][c+dc] {
					case PAINTED:
						painted++
					case CLEAR:
						clear++
					case UNKNOWN:
						target = Coordinate{r + dr, c + dc}
					}
				}
			}
			if painted == 3 && clear != 1 {
				didChange = s.MarkClear(target.Row, target.Col) || didChange
			}
		}
	}
	return didChange
}

func Check(b *Board, soln *Board) {
	abort := false
	if soln == nil {
		return
	}
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			if b.Grid[r][c] == UNKNOWN {
				continue
			}
			if b.Grid[r][c] != soln.Grid[r][c] {
				fmt.Printf("ERROR AT %v, %v\n%v\n", r, c, b)
				abort = true
			}
		}
	}
	for _, si := range soln.Islands {
		myI := b.IslandAt(si.Root.Row, si.Root.Col)
		found := false
		if myI.Members.Equals(si.Members) {
			found = true
		}
		for _, p := range myI.Possibilities {
			if p.Equals(si.Members) {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("ERROR on island %v (correct is %v)\n", myI, si)
			for _, p := range myI.Possibilities {
				fmt.Printf("Poss %v\n", p)
			}
			abort = true

		}
	}
	if abort {
		os.Exit(0)
	}
}

func (s *Solver) FalsifyGuess(r int, c int, cell Cell, skipExpensive bool) error {
	hypo := Solver{s.b.Clone(), nil, s.Action, nil}
	hypo.b.Mark(r, c, cell)
	hypo.AutoSolve(false, skipExpensive)
	return hypo.b.ContainsError()
}

func (s *Solver) MakeAGuess(neighborsOnly bool, skipExpensive bool) bool {
	if neighborsOnly {
		s.UpdateAction("Make a guess (island neighbors)")
	} else {
		s.UpdateAction("Make a guess (non island neighbors)")
	}
	for r := 0; r < s.b.Problem.Height; r++ {
		for c := 0; c < s.b.Problem.Width; c++ {
			if s.b.Grid[r][c] != UNKNOWN {
				continue
			}
			if neighborsOnly && !s.b.HasNeighborWith(SingleCoordinateSet(Coordinate{r, c}), CLEAR) {
				continue
			}
			e := s.FalsifyGuess(r, c, CLEAR, skipExpensive)
			if e != nil {
				s.MarkPainted(r, c)
				return true
			}
			e = s.FalsifyGuess(r, c, PAINTED, skipExpensive)
			if e != nil {
				s.MarkClear(r, c)
				return true
			}
		}
	}
	return false
}

func (s *Solver) InitSolve() {
	s.UpdateAction("Initialize solve")
	s.PaintTwoBorderedCells()
	s.ExtendIslandsOneLiberty()
	s.AddIslandBorders()
	s.PopulateIslandPossibilities()
}

func (s *Solver) AutoSolve(makeGuesses bool, skipExpensive bool) bool {
	Watch.Start("AutoSolve")
	changed := true
	for changed {
		changed = false
		changed = changed || s.PaintTwoBorderedCells()
		changed = changed || s.ExtendIslandsOneLiberty()
		changed = changed || s.AddIslandBorders()
		changed = changed || s.PaintUnreachables()
		s.UpdateAction("Stripping possibilities")
		changed = changed || s.b.StripAllPossibilities()
		changed = changed || s.ExtendWallIslandsOneLiberty()
		changed = changed || s.ConnectUnrootedIslands()
		changed = changed || s.FindSinglePoolPreventers()
		changed = changed || s.FillIslandNecessaries()
		changed = changed || s.AddIslandBorders()
		changed = changed || s.FillElbows()
		changed = changed || s.AddIslandBorders()
		changed = changed || s.ExtendWallIslands()

		if s.b.TotalMarked == s.b.Problem.Size {
			break
		}
		if err := s.b.ContainsError(); err != nil {
			break
		}
		if !skipExpensive {
			changed = changed || s.EliminateIntolerables()
			changed = changed || s.EliminateWallSplitters()
		}
		if makeGuesses {
			changed = changed || s.MakeAGuess(true, true)
			changed = changed || s.MakeAGuess(false, true)
			changed = changed || s.MakeAGuess(true, skipExpensive || false)
			changed = changed || s.MakeAGuess(false, skipExpensive || false)
		}
	}
	Watch.Stop("AutoSolve")
	return true
}
