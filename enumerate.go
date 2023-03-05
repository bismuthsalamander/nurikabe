package main

import (
	"fmt"
)

func (s *Solver) FindPossibleIslandsRec(c chan *CoordinateSet, members *CoordinateSet, targetSize int, css *CoordinateSetSet) {
	if css.Contains(members) {
		return
	}
	potentialNew := s.b.NeighborsWith(members, UNKNOWN)
	for p := range potentialNew.Map {
		membersNew := members.Copy()
		membersNew.Add(p)
		if css.Contains(membersNew) {
			continue
		}
		for {
			mergeThese := s.b.NeighborsWith(membersNew, CLEAR)
			if mergeThese.IsEmpty() {
				break
			}
			membersNew.AddAll(mergeThese)
		}
		if css.Contains(membersNew) {
			continue
		}
		if s.b.CountNumberedIslands(membersNew) == 1 {
			if membersNew.Size() == targetSize {
				css.Add(membersNew)
				c <- membersNew.Copy()
			} else if membersNew.Size() < targetSize {
				s.FindPossibleIslandsRec(c, membersNew, targetSize, css)
				css.Add(membersNew)
			}
		}
	}
	css.Add(members)
}

func (s *Solver) FindPossibleIslands(c chan *CoordinateSet, island *Island) {
	if island.CurrentSize >= island.TargetSize {
		close(c)
		return
	}
	hypoMembers := island.Members.Copy()
	css := EmptyCoordinateSetSet()
	s.FindPossibleIslandsRec(c, hypoMembers, island.TargetSize, css)
	close(c)
}

func (s *Solver) PopulateIslandPossibilities() {
	Watch.Start("Pop poss")
	defer Watch.Stop("Pop poss")
	for _, island := range s.b.Islands {
		if len(island.Possibilities) > 0 {
			island.Possibilities = make([]*CoordinateSet, 0)
		}
		c := make(chan *CoordinateSet)
		go s.FindPossibleIslands(c, island)
		for p := range c {
			island.Possibilities = append(island.Possibilities, p)
		}
		island.PopulateReachables()
	}
	s.b.PopulateUnrootedPossibilities()
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

func (s *Solver) PopulateAllReachables() {
	s.UpdateAction("Unreachables")
	for _, island := range s.b.Islands {
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
	for idx := 0; idx < len(i.Possibilities); idx++ {
		p := i.Possibilities[idx]
		if p == nil || i.Members == nil {
			fmt.Printf("Error: unexpected nil pointer (i.Possibilities member: %v; i.Members: %v)\n", p, i.Members)
		}
		if p.ContainsAll(i.Members) {
			if !b.HasNeighborWith(p, CLEAR) {
				continue
			}
		}
		RemoveFromSlice(&i.Possibilities, idx)
		idx--
	}
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

func (s *Solver) FillIslandNecessaries() bool {
	s.UpdateAction("Filling necessaries")
	Watch.Start("FIN")
	defer Watch.Stop("FIN")
	didChange := false
	ct := 0
	for _, i := range s.b.Islands {
		ct++
		if i.CurrentSize >= i.TargetSize {
			continue
		}
		var necessary *CoordinateSet = nil
		var necessaryNeighbors *CoordinateSet = nil
		for _, p := range i.Possibilities {
			if necessary == nil {
				necessary = p.Copy()
				necessaryNeighbors = s.b.NeighborsWith(necessary, UNKNOWN)
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
				didChange = s.MarkClear(target.Row, target.Col) || didChange
			}
			for target := range necessaryNeighbors.Map {
				didChange = s.MarkPainted(target.Row, target.Col) || didChange
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

func (s *Solver) PaintUnreachables() bool {
	s.UpdateAction("Painting unreachables")
	Watch.Start("PaintUnreachables")
	defer Watch.Stop("PaintUnreachables")
	const UNREACHABLE = 0
	const REACHABLE = 1
	s.b.ClearScratchGrid()
	for _, i := range s.b.Islands {
		for r := range i.Reachable.Map {
			s.b.ScratchGrid[r.Row][r.Col] = REACHABLE
		}
	}
	changed := false
	for r := 0; r < s.b.Problem.Height; r++ {
		for c := 0; c < s.b.Problem.Width; c++ {
			if s.b.Grid[r][c] == UNKNOWN && s.b.ScratchGrid[r][c] == UNREACHABLE {
				changed = s.MarkPainted(r, c) || changed
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

func (s *Solver) FindSinglePoolPreventers() bool {
	s.UpdateAction("Single pool preventers")
	Watch.Start("Find Pool Preventers")
	defer Watch.Stop("Find Pool Preventers")
	didChange := false
	for r := 0; r < s.b.Problem.Height-1; r++ {
	onePossiblePool:
		for c := 0; c < s.b.Problem.Width-1; c++ {
			var savior *Island = nil
			cs := EmptyCoordinateSet()
			for dr := 0; dr < 2; dr++ {
				for dc := 0; dc < 2; dc++ {
					c := Coordinate{r + dr, c + dc}
					if s.b.Get(c) == CLEAR {
						continue onePossiblePool
					}
					for _, i := range s.b.Islands {
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

func (s *Solver) ConnectUnrootedIslands() bool {
	s.UpdateAction("Connect unrooted islands")
	Watch.Start("Connect Unrooted Islands")
	defer Watch.Stop("Connect Unrooted Islands")
	didChange := false
oneUnrootedIsland:
	for _, i := range s.b.Islands {
		if i.IsRooted() {
			continue
		}
		mem := i.Members.OneMember()
		var savior *Island = nil
		for _, o := range s.b.Islands {
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

func (s *Solver) EliminateWallSplitters() bool {
	s.UpdateAction("Eliminate wall splitters")
	Watch.Start("EliminateWallSplitters")
	defer Watch.Stop("EliminateWallSplitters")
	changed := false
	for _, i := range s.b.Islands {
		for idx := 0; idx < len(i.Possibilities); idx++ {
			eliminate := false
			if s.b.SetSplitsWalls(i.Possibilities[idx]) {
				eliminate = true
			} else {
				if s.b.NeighborsWith(i.Possibilities[idx], CLEAR).Size() > 0 {
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

func (s *Solver) EliminateIntolerables() bool {
	s.UpdateAction("Eliminating intolerable possibilities")
	Watch.Start("Eliminate Intolerables")
	defer Watch.Stop("Eliminate Intolerables")
	didChange := false
	for _, i := range s.b.Islands {
		if i.TargetSize <= i.CurrentSize {
			continue
		}
		for idx := 0; idx < len(i.Possibilities); idx++ {
			p := i.Possibilities[idx]
			intolerable := false
			for _, o := range s.b.Islands {
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
