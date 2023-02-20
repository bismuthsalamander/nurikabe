package main

import (
	"fmt"
	"os"
)

func (b *Board) FindPossibleIslandsRec(c chan *CoordinateSet, members *CoordinateSet, targetSize int, css *CoordinateSetSet) {
	if css.Contains(members) {
		return
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
			if membersNew.Size() == targetSize /*&& !b.SetSplitsWalls(membersNew)*/ {
				css.Add(membersNew)
				c <- membersNew.Copy()
			} else if membersNew.Size() < targetSize {
				b.FindPossibleIslandsRec(c, membersNew, targetSize, css)
				css.Add(membersNew)
			}
		}
	}
	css.Add(members)
}

func (b *Board) FindPossibleIslands(c chan *CoordinateSet, island *Island) {
	if island.CurrentSize >= island.TargetSize {
		close(c)
		return
	}
	hypoMembers := island.Members.Copy()
	css := EmptyCoordinateSetSet()
	b.FindPossibleIslandsRec(c, hypoMembers, island.TargetSize, css)
	close(c)
}

func (b *Board) PopulateIslandPossibilities() {
	Watch.Start("Pop poss")
	for _, island := range b.Islands {
		if len(island.Possibilities) > 0 {
			island.Possibilities = make([]*CoordinateSet, 0)
		}
		c := make(chan *CoordinateSet)
		go b.FindPossibleIslands(c, island)
		for p := range c {
			island.Possibilities = append(island.Possibilities, p)
		}
		island.PopulateReachables()
	}
	Watch.Stop("Pop poss")
}

func (b *Board) RepopulateIslandReachables() {
	Watch.Start("Repop reach")
	for _, island := range b.Islands {
		island.PopulateReachables()
	}
	Watch.Stop("Repop reach")
}

func (i *Island) PopulateReachables() {
	Watch.Start("Reachables")
	i.Reachable = EmptyCoordinateSet()
	for _, p := range i.Possibilities {
		i.Reachable.AddAll(p)
	}
	Watch.Stop("Reachables")
}

func (i *Island) StripPossibilities() {
	Watch.Start("Strip Poss")
	if len(i.Possibilities) == 0 {
		return
	}
	newPossibilities := make([]*CoordinateSet, 0, len(i.Possibilities))
	for _, p := range i.Possibilities {
		if p.ContainsAll(i.Members) {
			newPossibilities = append(newPossibilities, p)
		}
	}
	i.Possibilities = newPossibilities
	Watch.Stop("Strip Poss")
}

func (i *Island) MustIncludeOne(cs *CoordinateSet) {
	Watch.Start("MustInclude")
	if len(i.Possibilities) == 0 {
		return
	}
	newPossibilities := make([]*CoordinateSet, 0, len(i.Possibilities))
	for _, p := range i.Possibilities {
		if p.ContainsAtLeastOne(cs) {
			newPossibilities = append(newPossibilities, p)
		}
	}
	i.Possibilities = newPossibilities
	Watch.Stop("MustInclude")
}

func (b *Board) FillIslandNecessaries() bool {
	Watch.Start("FIN")
	didChange := false
	ct := 0
	for _, i := range b.Islands {
		ct++
		if i.CurrentSize >= i.TargetSize {
			continue
		}
		var necessary *CoordinateSet = nil
		var necessaryNeighbors *CoordinateSet = nil
		for _, p := range i.Possibilities {
			if necessary == nil {
				necessary = p.Copy()
				necessaryNeighbors = b.NeighborsWith(necessary, UNKNOWN)
				necessary.DelAll(i.Members)
				continue
			}
			for cell := range necessary.Map {
				if !p.Contains(cell) {
					necessary.Del(cell)
				}
			}
			for cell := range necessaryNeighbors.Map {
				if !p.BordersCoordinate(cell) {
					necessaryNeighbors.Del(cell)
				}
			}
			if necessary.IsEmpty() && necessaryNeighbors.IsEmpty() {
				break
			}
		}
		if necessary == nil {
			fmt.Printf("ERROR! %v has no possibilities!", i)
			os.Exit(1)
		}
		for target := range necessary.Map {
			didChange = b.MarkClear(target.Row, target.Col) || didChange
		}
		for target := range necessaryNeighbors.Map {
			didChange = b.MarkPainted(target.Row, target.Col) || didChange
		}
	}
	Watch.Stop("FIN")
	return didChange
}

// TODO: have group versions of this and MarkPainted? Run through a whole slice for efficiency?
func (b *Board) RemoveFromPossibilities(newlyPainted Coordinate) {
	Watch.Start("RemoveFromPossibilities")
	for _, i := range b.Islands {
		for idx := 0; idx < len(i.Possibilities); idx++ {
			if i.Possibilities[idx].Contains(newlyPainted) {
				oldLen := len(i.Possibilities)
				i.Possibilities[idx] = i.Possibilities[oldLen-1]
				i.Possibilities = i.Possibilities[:oldLen-1]
				idx--
			}
		}
	}
	for _, i := range b.Islands {
		for idx := 0; idx < len(i.Possibilities); idx++ {
			if i.Possibilities[idx].Contains(newlyPainted) {
				fmt.Printf("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ error %v contains %v\n", i.Possibilities[idx], newlyPainted)
			}
		}
	}
	Watch.Stop("RemoveFromPossibilities")
}

func (b *Board) PaintUnreachablesFast() bool {
	Watch.Start("PaintUnreachablesFast")
	const UNREACHABLE = 0
	const REACHABLE = 1
	reachability := NewGrid(b.Problem.Width, b.Problem.Height)
	for _, i := range b.Islands {
		for r := range i.Reachable.Map {
			reachability[r.Row][r.Col] = REACHABLE
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
	Watch.Stop("PaintUnreachablesFast")
	return changed
}

func (i *Island) CanReach(target Coordinate) bool {
	//XXX - will this give us errors when we implement unrooted reachout?
	if !i.IsRooted() {
		return true
	}
	//TODO - make sure this optimization actually helps?
	if i.Root.ManhattanDistance(target) >= i.TargetSize {
		return false
	}
	for _, p := range i.Possibilities {
		if p.Contains(target) {
			return true
		}
	}
	return false
}

func (b *Board) ConnectTwoByOnes() bool {
	Watch.Start("Connect Two By Ones")
	defer Watch.Stop("Connect Two By Ones")
	didChange := false
	for r := 0; r < b.Problem.Height-1; r++ {
	onePossiblePool:
		for c := 0; c < b.Problem.Width-1; c++ {
			var savior *Island = nil
			cs := EmptyCoordinateSet()
			for dr := 0; dr < 2; dr++ {
				for dc := 0; dc < 2; dc++ {
					c := Coordinate{r + dr, c + dc}
					if b.Get(c) == CLEAR {
						continue onePossiblePool
					}
					for _, i := range b.Islands {
						if i.CanReach(c) {
							if savior != nil && savior != i {
								continue onePossiblePool
							}
							savior = i
						}
					}
					cs.Add(c)
				}
			}
			didChange = true
			savior.MustIncludeOne(cs)
		}
	}
	return didChange
}

func (b *Board) ConnectUnrootedIslands() bool {
	Watch.Start("Connect Unrooted Islands")
	defer Watch.Stop("Connect Unrooted Islands")
	didChange := false
oneUnrootedIsland:
	for _, i := range b.Islands {
		if i.IsRooted() {
			continue
		}
		mem := i.Members.OneMember()
		var savior *Island = nil
		for _, o := range b.Islands {
			if !o.IsRooted() {
				continue
			}
			if o.CanReach(mem) {
				if savior != nil {
					continue oneUnrootedIsland
				}
				savior = o
			}
		}
		if savior != nil {
			fmt.Printf("Island %v saved by %v\n", i, savior)
			savior.MustIncludeOne(SingleCoordinateSet(mem))
			didChange = true
		}
	}
	return didChange
}

func (b *Board) EliminateWallSplitters() bool {
	Watch.Start("EliminateWallSplitters")
	changed := false
	for _, i := range b.Islands {
		for idx := 0; idx < len(i.Possibilities); idx++ {
			eliminate := false
			if b.SetSplitsWalls(i.Possibilities[idx]) {
				eliminate = true
			} else {
				Watch.Start("EliminateTouchAnotherNumberedIsland")
				//Eliminate possibilities that would touch another numbered island
				tmpMembers := i.Possibilities[idx].Copy()
				for {
					mergeThese := b.NeighborsWith(tmpMembers, CLEAR)
					if mergeThese.IsEmpty() {
						break
					}
					tmpMembers.AddAll(mergeThese)
				}
				if b.CountNumberedIslands(tmpMembers) > 1 || tmpMembers.Size() > i.TargetSize {
					eliminate = true
				}
				Watch.Stop("EliminateTouchAnotherNumberedIsland")
			}
			if eliminate {
				oldLen := len(i.Possibilities)
				i.Possibilities[idx] = i.Possibilities[oldLen-1]
				i.Possibilities = i.Possibilities[:oldLen-1]
				idx--
				changed = true
			}
		}
	}
	Watch.Stop("EliminateWallSplitters")
	return changed
}
