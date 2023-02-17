package main

import "fmt"

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
			if membersNew.Size() == targetSize && !b.SetSplitsWalls(membersNew) {
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
		fmt.Printf("CS >= target")
		close(c)
		return
	}
	hypoMembers := island.Members.Copy()
	css := EmptyCoordinateSetSet()
	b.FindPossibleIslandsRec(c, hypoMembers, island.TargetSize, css)
	close(c)
}
