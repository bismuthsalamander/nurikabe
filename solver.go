package main

import (
	"fmt"
	"os"
)

func (b *Board) AddIslandBorders() bool {
	Watch.Start("AIB")
	defer Watch.Stop("AIB")
	didChange := false
	for _, island := range b.Islands {
		if island.ReadyForBorders {
			targets := b.NeighborsWith(island.Members, UNKNOWN)
			for coord := range targets.Map {
				didChange = b.MarkPainted(coord.Row, coord.Col) || didChange
			}
			island.ReadyForBorders = false
		}
	}
	return didChange
}

// TODO: liberty data structure? running slices?
func (b *Board) ExtendIslandsOneLiberty() bool {
	Watch.Start("EI1")
	defer Watch.Stop("EI1")
	didChange := false
	changed := true
	for changed {
		changed = false
		for _, island := range b.Islands {
			if island.CurrentSize == island.TargetSize {
				continue
			}
			lib := b.Liberties(island)
			if lib.Size() == 1 {
				c := lib.OneMember()
				result := b.MarkClear(c.Row, c.Col)
				didChange = didChange || result
				changed = changed || result
				break
			}
		}
	}
	return didChange
}

// TODO: liberty data structure? running slices?
func (b *Board) ExtendWallIslandsOneLiberty() bool {
	Watch.Start("EW1")
	defer Watch.Stop("EW1")
	didChange := false
	changed := true
outerLoop:
	for changed {
		changed = false
		for _, island := range b.WallIslands {
			if len(b.WallIslands) == 1 {
				break
			}
			if island.CurrentSize == b.Problem.TargetWallCount {
				continue
			}
			lib := b.Liberties(island)
			if lib.Size() == 1 {
				for c := range lib.Map {
					didChange = b.MarkPainted(c.Row, c.Col) || didChange
					changed = true
					continue outerLoop
				}
			}
		}
	}
	return didChange
}

// would it be faster to do this the other way around?
// have each island count their liberties and identify each cell that is a
// liberty to two different islands?
func (b *Board) PaintTwoBorderedCells() bool {
	Watch.Start("P2B")
	defer Watch.Stop("P2B")
	didChange := false
	for ri, row := range b.Grid {
		for ci, col := range row {
			if col != UNKNOWN {
				continue
			}
			borderCount := 0
			for _, i := range b.Islands {
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
				didChange = b.MarkPainted(ri, ci) || didChange
			}
		}
	}
	return didChange
}

func (b *Board) WallDfs(members *CoordinateSet) *CoordinateSet {
	necessary := EmptyCoordinateSet()
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			necessary.Add(Coordinate{r, c})
		}
	}
	necessary.DelAll(members)
	b.WallDfsRec(members, necessary)
	return necessary
}

func (b *Board) WallDfsRec(members *CoordinateSet, necessary *CoordinateSet) {
	if necessary.IsEmpty() || members.ContainsAll(necessary) {
		return
	}
	if b.HasNeighborWith(members, PAINTED) {
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
	neighbors := b.NeighborsWith(members, UNKNOWN)
	for n := range neighbors.Map {
		if b.Get(n) == UNKNOWN && members.CanAddWall(n) {
			members.Add(n)
			b.WallDfsRec(members, necessary)
			members.Del(n)
		}
	}
}

// Two possible optimizations:
// 1. Track which CoordinateSets we've already had
// 2. Do it all in one recursive function, passing necessary through and dumping the channel; skip any possibility
// with necessary as a subset of the possibility
func (b *Board) ExtendWallIslands() bool {
	Watch.Start("EWI")
	defer Watch.Stop("EWI")
	if len(b.WallIslands) < 2 {
		return false
	}
	didChange := false
	for _, wi := range b.WallIslands {
		if wi.CurrentSize == b.Problem.TargetWallCount {
			continue
		}
		necessaryMembers := b.WallDfs(wi.Members)
		for target := range necessaryMembers.Map {
			didChange = b.MarkPainted(target.Row, target.Col) || didChange
		}
	}
	return didChange
}

func (b *Board) FillElbows() bool {
	Watch.Start("FEL")
	defer Watch.Stop("FEL")
	//TODO: make more efficient with overlapping columns that we save between inner loop iterations?
	didChange := false
	for r := 0; r < b.Problem.Height-1; r++ {
		for c := 0; c < b.Problem.Width-1; c++ {
			painted := 0
			clear := 0
			unknown := 0
			target := Coordinate{}
			for dr := 0; dr < 2; dr++ {
				for dc := 0; dc < 2; dc++ {
					switch b.Grid[r+dr][c+dc] {
					case PAINTED:
						painted++
					case CLEAR:
						clear++
					case UNKNOWN:
						unknown++
						target = Coordinate{r + dr, c + dc}
					}
				}
			}
			if painted == 3 && clear != 1 {
				didChange = b.MarkClear(target.Row, target.Col) || didChange
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

		} else {
			//fmt.Printf("Island %v is okay!! %v is in there!\n", myI, si.Members)
		}
	}
	if abort {
		os.Exit(0)
	}
}

func (b *Board) FalsifyGuess(r int, c int, cell Cell) error {
	hypo := b.Clone()
	hypo.Mark(r, c, cell)
	hypo.AutoSolve(nil, true)
	return hypo.ContainsError()
}

func (b *Board) MakeAGuess(neighborsOnly bool) bool {
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			if b.Grid[r][c] != UNKNOWN {
				continue
			}
			if neighborsOnly && !b.HasNeighborWith(SingleCoordinateSet(Coordinate{r, c}), CLEAR) {
				continue
			}
			e := b.FalsifyGuess(r, c, CLEAR)
			if e != nil {
				b.MarkPainted(r, c)
				fmt.Printf("Successfully guessed! {r%d, c%d} was painted! error %v\n", r, c, e)
				fmt.Printf("Stopwatch:\n%s", Watch.Results())
				return true
			}
			e = b.FalsifyGuess(r, c, PAINTED)
			if e != nil {
				fmt.Printf("Successfully guessed! {r%d, c%d} was clear! error %v\n", r, c, e)
				b.MarkClear(r, c)
				fmt.Printf("Stopwatch:\n%s", Watch.Results())
				return true
			}
		}
	}
	fmt.Printf("Unsuccessfully guessed.\n")
	return false
}

func (b *Board) InitSolve() {
	b.PaintTwoBorderedCells()
	b.ExtendIslandsOneLiberty()
	b.AddIslandBorders()
	b.PopulateIslandPossibilities()
}

func (b *Board) AutoSolve(sol *Board, guess bool) bool {
	Watch.Start("AutoSolve")
	changed := true
	chTmp := false
	for changed {
		changed = false
		chTmp = b.PaintTwoBorderedCells()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("1 painted 2bc %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		chTmp = b.ExtendIslandsOneLiberty()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("2 extended 1lib %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		chTmp = b.AddIslandBorders()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("3 added borders %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)

		chTmp = b.EliminateWallSplitters()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("3.5 removed splitters %v\n%v\n", changed, b.StringDebug())
		}

		b.PopulateAllReachables()
		chTmp := b.PaintUnreachables()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("4 painted URslow %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		b.StripAllPossibilities()
		chTmp = b.ExtendWallIslandsOneLiberty()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("5 extended WI1 %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		chTmp = b.ConnectUnrootedIslands()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("5.5 connected unrooted %v\n%v\n", changed, b.StringDebug())
		}
		chTmp = b.FindSinglePoolPreventers()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("5.6 connected 2x1s %v\n%v\n", changed, b.StringDebug())
		}
		chTmp = b.EliminateIntolerables()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("5.7 eliminated intolerables %v\n%v\n", changed, b.StringDebug())
		}
		chTmp = b.FillIslandNecessaries()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("6 filled necessaries %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		chTmp = b.AddIslandBorders()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("7 added borders %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		chTmp = b.FillElbows()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("8 filled elbows %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		chTmp = b.AddIslandBorders()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("9 added borders %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		chTmp = b.ExtendWallIslands()
		changed = changed || chTmp
		if chTmp && false {
			fmt.Printf("10 extended WI %v\n%v\n", changed, b.StringDebug())
		}
		Check(b, sol)
		if b.TotalMarked == b.Problem.Width*b.Problem.Height {
			break
		}
		if !changed && !guess {
			fmt.Println("Making a guess")
			changed = b.MakeAGuess(true) || changed
		}
		if !changed && !guess {
			fmt.Println("Making a guess ANYWHERE")
			changed = b.MakeAGuess(false) || changed
		}
	}
	Watch.Stop("AutoSolve")
	//fmt.Printf("Stopwatch:\n%s", Watch.Results())
	return true
}
