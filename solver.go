package main

import (
	"context"
	"fmt"
	"os"
)

func (b *Board) ExtendIslands() bool {
	Watch.Start("EI-")
	didChange := false
	for _, i := range b.Islands {
		if i.CurrentSize >= i.TargetSize {
			continue
		}
		ctx, cancel := context.WithCancel(context.Background())
		c := make(chan *CoordinateSet)
		go b.IslandDfs(ctx, c, i)
		necessary := <-c
		necessaryNeighbors := b.NeighborsWith(necessary, UNKNOWN)
		necessary.RemoveAll(i.Members)
		ct := 0
		for members := range c {
			ct++
			Watch.Start("BDR")
			for cell := range necessary.Map {
				if !members.Contains(cell) {
					necessary.Del(cell)
				}
			}
			for cell := range necessaryNeighbors.Map {
				if !members.BordersCoordinate(cell) {
					necessaryNeighbors.Del(cell)
				}
			}
			Watch.Stop("BDR")
			if necessary.IsEmpty() && necessaryNeighbors.IsEmpty() {
				cancel()
				break
			}
		}
		for target := range necessary.Map {
			didChange = b.MarkClear(target.Row, target.Col) || didChange
		}
		for target := range necessaryNeighbors.Map {
			didChange = b.MarkPainted(target.Row, target.Col) || didChange
		}
		cancel()
	}
	Watch.Stop("EI-")
	return didChange
}

func (b *Board) AddIslandBorders() bool {
	Watch.Start("AIB")
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
			// One of the only ugly parts of using CoordinateSet for everything
			if lib.Size() == 1 {
				for c := range lib.Map {
					result := b.MarkClear(c.Row, c.Col)
					didChange = didChange || result
					changed = changed || result
				}
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
	for k := range members.Map {
		necessary.Del(k)
	}
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

func (b *Board) IslandDfsRec(ctx context.Context, c chan *CoordinateSet, members *CoordinateSet, targetSize int, css *CoordinateSetSet) {
	if css.Contains(members) {
		return
	}
	select {
	case <-ctx.Done():
		return
	default:
	}
	potentialNew := b.NeighborsWith(members, UNKNOWN)
	for p := range potentialNew.Map {
		membersNew := members.Copy()
		membersNew.Add(p)
		if css.Contains(membersNew) {
			continue
		}
		for {
			mergeThese := b.NeighborsWith(membersNew, CLEAR)
			if mergeThese.IsEmpty() {
				break
			}
			membersNew.AddAll(mergeThese)
		}
		if css.Contains(membersNew) {
			continue
		}
		if b.CountNumberedIslands(membersNew) == 1 {
			if membersNew.Size() == targetSize && !b.SetSplitsWalls(membersNew) {
				css.Add(membersNew)
				c <- membersNew.Copy()
			} else if membersNew.Size() < targetSize {
				b.IslandDfsRec(ctx, c, membersNew, targetSize, css)
				css.Add(membersNew)
			}
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
	css.Add(members)
}

func (b *Board) IslandDfs(ctx context.Context, c chan *CoordinateSet, island *Island) {
	if island.CurrentSize >= island.TargetSize {
		close(c)
		return
	}
	hypoMembers := island.Members.Copy()
	css := EmptyCoordinateSetSet()
	b.IslandDfsRec(ctx, c, hypoMembers, island.TargetSize, css)
	close(c)
}

// "Can any of these source nodes reach the destination node within n steps?"
// No need for priority queues; we just expand the neighbor set on each iteration,
// making sure to exclude nodes already in the set, returning true if we ever see the
// destination node and returning false after n iterations.
//
// TODO: should we exclude certain impossible island expansions? e.g., diagonally connected
// line of clear cells from wall to wall? including cells that neighbors another island?
func (b *Board) CanIslandReach(i *Island, target Coordinate) bool {
	n := i.TargetSize - i.CurrentSize
	//Shortcut: if the island has no member m with manhattan distance(m, target) <= n, we return false
	found := false
	for source := range i.Members.Map {
		if source.ManhattanDistance(target) <= n {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	return b.CanIslandReachRec(i, i.Members, target, n)
}

func (b *Board) CanIslandReachRec(originIsland *Island, reachable *CoordinateSet, target Coordinate, n int) bool {
	if reachable.Contains(target) && n >= 0 {
		return true
	} else if n <= 0 {
		return false
	}
	newReachable := b.Neighbors(reachable)
	for k := range newReachable.Map {
		//For the newly reachables that do NOT border an existing island, add them in
		//For the newly reachables that DO border an existing island, remove them from newReachable, but
		//spawn off a recursive call that adds in EACH member of that island and reduces n accordingly
		//Could we include no-number islands by simply absorbing the entire island, then checking for
		//n>=0 and reachable.Contains(target)?
		if b.Get(k) == PAINTED {
			newReachable.Del(k)
		} else if b.Get(k) == UNKNOWN {
			if b.BordersMultipleRootedIslands(k) {
				newReachable.Del(k)
			} else {
				/*ni := b.BorderingIsland(k)
				if ni != nil && ni.Root != originIsland.Root {
					newReachable.Del(k)
					if b.CanIslandReachRec(originIsland, reachable.Plus(ni.Members), target, n-ni.Members.Size()) {
						return true
					}
				}*/
			}
		} else if b.Get(k) == CLEAR { //it's UNKNOWN
			fmt.Printf("ERROR!!!!")
			os.Exit(0)
		}
	}
	return b.CanIslandReachRec(originIsland, reachable.Plus(newReachable), target, n-1)
}

const REACHABLE = 1
const UNREACHABLE = 0

func (b *Board) TrackReachableIslands(results [][]Cell, originIsland *Island, reachable *CoordinateSet, n int) {
	if n >= 0 {
		for k := range reachable.Map {
			results[k.Row][k.Col] = REACHABLE
		}
	}
	if n <= 0 {
		return
	}
	newReachable := b.Neighbors(reachable)
	for k := range newReachable.Map {
		//For the newly reachables that do NOT border an existing island, add them in
		//For the newly reachables that DO border an existing island, remove them from newReachable, but
		//spawn off a recursive call that adds in EACH member of that island and reduces n accordingly
		//Could we include no-number islands by simply absorbing the entire island, then checking for
		//n>=0 and reachable.Contains(target)?
		if b.Get(k) == PAINTED {
			newReachable.Del(k)
		} else if b.Get(k) == UNKNOWN {
			if b.BordersMultipleRootedIslands(k) {
				newReachable.Del(k)
			} else {
				neighboringIslands := b.BorderingIslands(k)
				additionalNeighbors := EmptyCoordinateSet()
				for _, ni := range neighboringIslands {
					//this isn't right because we have to add ALL islands at once that border this cell
					if ni != nil && ni.Root != originIsland.Root {
						additionalNeighbors.AddAll(ni.Members)
					}
				}
				if additionalNeighbors.Size() > 0 {
					additionalNeighbors.Add(k)
					newReachable.Del(k)
					b.TrackReachableIslands(results, originIsland, reachable.Plus(additionalNeighbors), n-additionalNeighbors.Size())
				}
			}
		} else if b.Get(k) == CLEAR { //it's UNKNOWN
			fmt.Printf("ERROR!!!!")
			fmt.Printf("origin was %v\nreachable is %v\nsteps left is %v\nresults: %v\n", originIsland, reachable, n, results)
			fmt.Printf("Offending cell is %v\n", k)
			for r := 0; r < b.Problem.Height; r++ {
				for c := 0; c < b.Problem.Width; c++ {
					if results[r][c] == REACHABLE {
						fmt.Printf("X")
					} else {
						fmt.Printf(" ")
					}
				}
				fmt.Printf("\n")
			}
			fmt.Printf("%v\n", b.StringDebug())
			os.Exit(0)
		}
	}
	b.TrackReachableIslands(results, originIsland, reachable.Plus(newReachable), n-1)
}

func (b *Board) PaintUnreachables() bool {
	Watch.Start("PU-")
	reachability := NewGrid(b.Problem.Width, b.Problem.Height)
	for _, i := range b.Islands {
		if i.TargetSize > 0 {
			b.TrackReachableIslands(reachability, i, i.Members.Copy(), i.TargetSize-i.CurrentSize)
		}
	}
	changed := false
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			if b.Grid[r][c] == UNKNOWN && reachability[r][c] == UNREACHABLE {
				changed = b.MarkPainted(r, c) || changed
			}
		}
	}
	Watch.Stop("PU-")
	return changed
}

func (b *Board) PaintUnreachablesSlow() bool {
	didChange := false
	for r := 0; r < b.Problem.Height; r++ {
	oneCell:
		for c := 0; c < b.Problem.Width; c++ {
			coord := Coordinate{r, c}
			for _, i := range b.Islands {
				if b.CanIslandReach(i, coord) {
					continue oneCell
				}
			}
			didChange = b.MarkPainted(coord.Row, coord.Col) || didChange
		}
	}
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

func (b *Board) AutoSolve() bool {
	changed := true
	for changed {
		changed = false
		changed = b.PaintTwoBorderedCells() || changed
		changed = b.ExtendIslandsOneLiberty() || changed
		changed = b.AddIslandBorders() || changed
		changed = b.PaintUnreachables() || changed

		changed = b.ExtendWallIslandsOneLiberty() || changed
		//changed = b.ExtendIslands() || changed
		//changed = b.PaintUnreachables() || changed
		//changed = b.ExtendWallIslandsOneLiberty() || changed
		changed = b.ExtendIslands() || changed
		changed = b.AddIslandBorders() || changed

		changed = b.FillElbows() || changed
		//changed = b.PreventWallSplits() || changed
		changed = b.AddIslandBorders() || changed
		changed = b.ExtendWallIslands() || changed
		if b.TotalMarked == b.Problem.Width*b.Problem.Height {
			break
		}
	}
	fmt.Printf("Stopwatch:\n%s", Watch.Results())
	return true
}
