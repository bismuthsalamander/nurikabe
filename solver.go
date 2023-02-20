package main

import (
	"fmt"
	"os"
)

func (b *Board) AddIslandBorders() bool {
	Watch.Start("AIB")
	didChange := false
	for _, island := range b.Islands {
		if island.ReadyForBorders {
			fmt.Printf("It's ready")
			targets := b.NeighborsWith(island.Members, UNKNOWN)
			fmt.Printf("Have targets sz %d\n", targets.Size())
			for coord := range targets.Map {
				didChange = b.MarkPainted(coord.Row, coord.Col) || didChange
			}
			island.ReadyForBorders = false
		}
	}
	Watch.Stop("AIB")
	return didChange
}

// TODO: liberty data structure? running slices?
func (b *Board) ExtendIslandsOneLiberty() bool {
	Watch.Start("EI1")
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
				fmt.Printf("Extending island %v to %v\n", island, c)
				result := b.MarkClear(c.Row, c.Col)
				didChange = didChange || result
				changed = changed || result
				fmt.Printf("Island at %v is now %v\n", c, b.IslandAt(c.Row, c.Col))
				break
			}
		}
	}
	Watch.Stop("EI1")
	return didChange
}

// TODO: liberty data structure? running slices?
func (b *Board) ExtendWallIslandsOneLiberty() bool {
	Watch.Start("EW1")
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
	Watch.Stop("EW1")
	return didChange
}

// would it be faster to do this the other way around?
// have each island count their liberties and identify each cell that is a
// liberty to two different islands?
func (b *Board) PaintTwoBorderedCells() bool {
	Watch.Start("P2B")
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
	Watch.Stop("P2B")
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
	Watch.Stop("EWI")
	return didChange
}

func (b *Board) FillElbows() bool {
	Watch.Start("FEL")
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
	Watch.Stop("FEL")
	return didChange
}

func Check(b *Board, soln *Board) {
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
				os.Exit(0)
			}
		}
	}
}

func (b *Board) AutoSolve(sol *Board) bool {
	b.PaintTwoBorderedCells()
	b.ExtendIslandsOneLiberty()
	b.AddIslandBorders()
	b.PopulateIslandPossibilities()
	changed := true
	for changed {
		changed = false
		changed = b.PaintTwoBorderedCells() || changed
		fmt.Printf("1 painted 2bc\n%v\n", b.String())
		Check(b, sol)
		changed = b.ExtendIslandsOneLiberty() || changed
		fmt.Printf("2 extended 1lib\n%v\n", b.String())
		Check(b, sol)
		changed = b.AddIslandBorders() || changed
		fmt.Printf("3 added borders\n%v\n", b.String())
		Check(b, sol)
		//changed = b.PaintUnreachablesSlow() || changed
		changed = b.EliminateWallSplitters() || changed
		fmt.Printf("3.5 removed splitters\n%v\n", b.String())
		//changed = b.PaintUnreachablesSlow() || changed
		//changed = b.PaintUnreachablesFast() || changed
		//changed2 := b.PaintUnreachablesSlow()
		b.RepopulateIslandReachables()
		changed2 := b.PaintUnreachablesFast()
		changed = changed2 || changed
		fmt.Printf("4 painted URslow\n%v\n", b.String())
		if changed2 {
			fmt.Println("Changed2!!!!!!")
		}
		Check(b, sol)
		changed = b.ExtendWallIslandsOneLiberty() || changed
		fmt.Printf("5 extended WI1\n%v\n", b.String())
		Check(b, sol)
		//changed = b.ExtendIslands() || changed
		//changed = b.ExtendWallIslandsOneLiberty() || changed
		//changed = b.ExtendIslands() || changed

		changed = b.ConnectUnrootedIslands() || changed
		fmt.Printf("5.5 connected unrooted\n%v\n", b.String())
		changed = b.ConnectTwoByOnes() || changed
		fmt.Printf("5.6 connected 2x1s\n%v\n", b.String())
		changed = b.FillIslandNecessaries() || changed
		fmt.Printf("6 filled necessaries\n%v\n", b.String())
		Check(b, sol)
		changed = b.AddIslandBorders() || changed
		fmt.Printf("7 added borders\n%v\n", b.String())
		Check(b, sol)
		changed = b.FillElbows() || changed
		fmt.Printf("8 filled elbows\n%v\n", b.String())
		Check(b, sol)
		//changed = b.PreventWallSplits() || changed
		changed = b.AddIslandBorders() || changed
		fmt.Printf("9 added borders\n%v\n", b.String())
		Check(b, sol)
		changed = b.ExtendWallIslands() || changed
		fmt.Printf("10 extended WI\n%v\n", b.String())
		Check(b, sol)
		if b.TotalMarked == b.Problem.Width*b.Problem.Height {
			break
		}
		if !changed {
			//fmt.Printf("*************RESCUE")
			//changed = b.PaintUnreachablesSlow()
		}
	}
	fmt.Printf("Stopwatch:\n%s", Watch.Results())
	return true
}
