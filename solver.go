package main

func (b *Board) ExtendIslands() {
	for _, i := range b.Islands {
		if i.CurrentSize == i.TargetSize {
			continue
		}
		c := make(chan *CoordinateSet)
		go b.IslandDfs(c, &i)
		necessary := <-c
		necessaryNeighbors := b.NeighborsWith(necessary, UNKNOWN)
		for members := range c {
			for cell := range necessary.Map {
				if !members.Contains(cell) {
					necessary.Del(cell)
				}
			}
			for cell := range necessaryNeighbors.Map {
				if !members.Contains(cell) {
					necessaryNeighbors.Del(cell)
				}
			}
		}
		for target := range necessary.Map {
			b.MarkClear(target.Row, target.Col)
		}
		for target := range necessaryNeighbors.Map {
			b.MarkPainted(target.Row, target.Col)
		}
	}
}

func (b *Board) AddIslandBorders() {
	for _, island := range b.Islands {
		if !island.BordersAdded && island.CurrentSize == island.TargetSize {
			targets := b.NeighborsWith(island.Members, UNKNOWN)
			for coord := range targets.Map {
				b.MarkPainted(coord.Row, coord.Col)
			}
			island.BordersAdded = true
		}
	}
}

// TODO: liberty data structure? running slices?
func (b *Board) ExpandIslandsOneLiberty() {
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
					b.MarkClear(c.Row, c.Col)
					changed = true
				}
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
			if lib.Size() == 1 {
				for c := range lib.Map {
					b.MarkPainted(c.Row, c.Col)
					changed = true
				}
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

func (b *Board) WallDfsRec(c chan *CoordinateSet, members *CoordinateSet, depth int) {
	neighbors := b.NeighborsWith(members, PAINTED)
	if neighbors.Size() > 0 {
		c <- members.Copy()
		return
	}
	neighbors = b.NeighborsWith(members, UNKNOWN)
	for n := range neighbors.Map {
		if b.Get(n) == PAINTED {
			c <- members.Copy()
		} else if b.Get(n) == UNKNOWN {
			members.Add(n)
			b.WallDfsRec(c, members, depth+1)
			members.Del(n)
		}
	}
}

func (b *Board) WallDfs(c chan *CoordinateSet, island *Island) {
	members := island.Members.Copy()
	b.WallDfsRec(c, members, 0)
	close(c)
}

func (b *Board) ExtendWallIslands() {
	for _, wi := range b.WallIslands {
		if wi.CurrentSize == b.Problem.TargetWallCount {
			continue
		}
		c := make(chan *CoordinateSet)
		go b.WallDfs(c, &wi)
		necessary := <-c
		for p := range c {
			for cell := range necessary.Map {
				if !p.Contains(cell) {
					necessary.Del(cell)
				}
			}
		}
		for target := range necessary.Map {
			b.MarkPainted(target.Row, target.Col)
		}
	}
}

func (b *Board) IslandDfsRec(c chan *CoordinateSet, members *CoordinateSet, targetSize int) {
	potentialNew := b.NeighborsWith(members, UNKNOWN)
	for p := range potentialNew.Map {
		membersNew := members.Copy()
		membersNew.Add(p)
		membersNew = membersNew.Plus(b.NeighborsWith(members, CLEAR))
		if b.CountIslands(membersNew) == 1 {
			if membersNew.Size() == targetSize {
				c <- membersNew.Copy()
			} else {
				b.IslandDfsRec(c, membersNew, targetSize)
			}
		}
	}
}

func (b *Board) IslandDfs(c chan *CoordinateSet, island *Island) {
	if island.CurrentSize == island.TargetSize {
		close(c)
		return
	}
	hypoMembers := EmptyCoordinateSet().Plus(island.Members)
	b.IslandDfsRec(c, hypoMembers, island.TargetSize)
	close(c)
}

func (b *Board) FillElbows() {
	//TODO: make more efficient with overlapping columns that we save between inner loop iterations?
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
				b.MarkClear(target.Row, target.Col)
			}
		}
	}
}
