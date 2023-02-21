package main

import (
	"fmt"
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
			if membersNew.Size() == targetSize {
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
	defer Watch.Stop("Pop poss")
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
	b.PopulateUnrootedPossibilities()
}

func (b *Board) PopulateUnrootedPossibilities() {
	for _, island := range b.Islands {
		if island.IsRooted() {
			continue
		}
		if len(island.Possibilities) > 0 {
			continue
		}
		for _, o := range b.Islands {
			for _, p := range o.Possibilities {
				if p.ContainsAll(island.Members) {
					island.Possibilities = append(island.Possibilities, p)
				}
			}
		}
	}
}

func (b *Board) PopulateAllReachables() {
	for _, island := range b.Islands {
		island.PopulateReachables()
	}
}

func (i *Island) PopulateReachables() {
	Watch.Start("Reachables")
	defer Watch.Stop("Reachables")
	i.Reachable = EmptyCoordinateSet()
	for _, p := range i.Possibilities {
		i.Reachable.AddAll(p)
	}
}

func (b *Board) StripAllPossibilities() {
	for _, i := range b.Islands {
		b.StripPossibilities(i)
	}
}

func (b *Board) StripPossibilities(i *Island) {
	Watch.Start("Strip Poss")
	defer Watch.Stop("Strip Poss")
	if len(i.Possibilities) == 0 || i.IsComplete() {
		return
	}
	newPossibilities := make([]*CoordinateSet, 0, len(i.Possibilities))
	for _, p := range i.Possibilities {
		if p == nil || i.Members == nil {
			fmt.Printf("Error: unexpected nil pointer (i.Possibilities member: %v; i.Members: %v)\n", p, i.Members)
		}
		if !p.ContainsAll(i.Members) {
			continue
		}
		extras := b.NeighborsWith(p, CLEAR)
		if extras.Size() > 0 {
			continue
		}
		newPossibilities = append(newPossibilities, p)
	}
	i.Possibilities = newPossibilities
}

func (i *Island) MustIncludeOne(cs *CoordinateSet) bool {
	Watch.Start("MustInclude")
	defer Watch.Stop("MustInclude")
	if len(i.Possibilities) == 0 {
		return false
	}
	oldLen := len(i.Possibilities)
	newPossibilities := make([]*CoordinateSet, 0, len(i.Possibilities))
	for _, p := range i.Possibilities {
		if p.ContainsAtLeastOne(cs) {
			newPossibilities = append(newPossibilities, p)
		}
	}
	i.Possibilities = newPossibilities
	return len(i.Possibilities) != oldLen
}

func (b *Board) FillIslandNecessaries() bool {
	Watch.Start("FIN")
	defer Watch.Stop("FIN")
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
		if necessary != nil {
			for target := range necessary.Map {
				didChange = b.MarkClear(target.Row, target.Col) || didChange
			}
			for target := range necessaryNeighbors.Map {
				didChange = b.MarkPainted(target.Row, target.Col) || didChange
			}
		}
	}
	return didChange
}

func (b *Board) RemoveFromPossibilities(newlyPainted Coordinate) {
	Watch.Start("RemoveFromPossibilities")
	defer Watch.Stop("RemoveFromPossibilities")
	for _, i := range b.Islands {
		for idx := 0; idx < len(i.Possibilities); idx++ {
			if i.Possibilities[idx].Contains(newlyPainted) {
				RemoveFromSlice(&i.Possibilities, idx)
				idx--
			}
		}
	}
}

func (b *Board) PaintUnreachables() bool {
	Watch.Start("PaintUnreachables")
	defer Watch.Stop("PaintUnreachables")
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
	return changed
}

func (i *Island) CanReach(target Coordinate) bool {
	//Unrooted islands will never be the only islands that can reach a cell, so we can skip them
	if !i.IsRooted() {
		return true
	}
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

func (b *Board) FindSinglePoolPreventers() bool {
	Watch.Start("Find Pool Preventers")
	defer Watch.Stop("Find Pool Preventers")
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
			if savior != nil {
				didChange = savior.MustIncludeOne(cs) || didChange
			}
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
			didChange = savior.MustIncludeOne(SingleCoordinateSet(mem)) || didChange
		}
	}
	return didChange
}

func (b *Board) EliminateWallSplitters() bool {
	Watch.Start("EliminateWallSplitters")
	defer Watch.Stop("EliminateWallSplitters")
	changed := false
	for _, i := range b.Islands {
		for idx := 0; idx < len(i.Possibilities); idx++ {
			eliminate := false
			if b.SetSplitsWalls(i.Possibilities[idx]) {
				eliminate = true
			} else {
				if b.NeighborsWith(i.Possibilities[idx], CLEAR).Size() > 0 {
					eliminate = true
				}
			}
			if eliminate {
				RemoveFromSlice(&i.Possibilities, idx)
				idx--
				changed = true
			}
		}
	}
	return changed
}

func (i *Island) CanToleratePossibility(cs *CoordinateSet) bool {
	for _, p := range i.Possibilities {
		if !i.IsRooted() && p.Equals(cs) {
			return true
		}
		if !p.ContainsAtLeastOne(cs) {
			return true
		}
	}
	return false
}

func (b *Board) EliminateIntolerables() bool {
	Watch.Start("Eliminate Intolerables")
	defer Watch.Stop("Eliminate Intolerables")
	didChange := false
	for _, i := range b.Islands {
		if i.TargetSize <= i.CurrentSize {
			continue
		}
		for idx := 0; idx < len(i.Possibilities); idx++ {
			p := i.Possibilities[idx]
			intolerable := false
			for _, o := range b.Islands {
				if o.IsRooted() && o.Root == i.Root {
					continue
				}
				if o.IsComplete() || len(o.Possibilities) == 0 {
					continue
				}
				if !o.CanToleratePossibility(p) {
					intolerable = true
					break
				}
			}
			if intolerable {
				RemoveFromSlice(&i.Possibilities, idx)
				idx--
				didChange = true
			}
		}
	}
	return didChange
}
