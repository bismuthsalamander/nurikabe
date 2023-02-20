package main

// Would the proposed CoordinateSet, if entered in the problem as an island,
// necessarily force the walls to be split into two (or more) wall islands?
func (b *Board) SetSplitsWalls(cs *CoordinateSet) bool {
	Watch.Start("SetSplitsWalls")
	defer Watch.Stop("SetSplitsWalls")

	//Start by merging the coordinate set with the CSes of all islands it
	//borders diagnoally
	islands := make([]*Island, len(b.Islands))
	mergedIslands := cs.Copy()
	copy(islands, b.Islands)
	for idx := 0; idx < len(islands); idx++ {
		if islands[idx].BordersSetDiagonally(mergedIslands) {
			mergedIslands.AddAll(islands[idx].Members)
			oldLen := len(islands)
			islands[idx] = islands[oldLen-1]
			islands = islands[:oldLen-1]
			idx--
		}
	}

	//To detect a wall-splitting island, we can visit each cell on the puzzle's
	//border in turn, starting at {0,0} and continuing clockwise. As we visit
	//each cell, we check whether the cell is part of the proposed coordinate
	//set, counting how often a cell's membership value (true/false) differs
	//from the previous cell.  If we have at least three changes, then the set
	//splits the walls into multiple islands. (Note that there must be an even
	//total number of changes!)
	wasLastMember := mergedIslands.Contains(Coordinate{0, 0})
	coord := Coordinate{0, 1}

	//function that walks clockwise around the puzzle's border; easier to write
	//with the ability to use return
	walkClockwise := func(c Coordinate) Coordinate {
		if coord.Row == 0 {
			if coord.Col < b.Problem.Width-1 {
				return coord.Translate(0, 1)
			}
		}
		if coord.Col == b.Problem.Width-1 {
			if coord.Row < b.Problem.Height-1 {
				return coord.Translate(1, 0)
			}
		}
		if coord.Row == b.Problem.Height-1 {
			if coord.Col > 0 {
				return coord.Translate(0, -1)
			}
		}
		if coord.Col == 0 {
			return coord.Translate(-1, 0)
		}
		return coord
	}
	changes := 0
	for {
		isMember := mergedIslands.Contains(coord)
		if isMember != wasLastMember {
			changes++
			if changes > 2 {
				return true
			}
			wasLastMember = isMember
		}
		coord = walkClockwise(coord)
		if coord.Row == 0 && coord.Col == 0 {
			break
		}
	}

	//To detect an island that would isolate an interior wall island - e.g., this:
	//
	// ___...___
	// __.._..._
	// _.._X__..
	// _.__X___.
	// __......_
	//
	// ...we can do the following:
	// - Create a new grid with mergedIsland members BLOCKED and border cells COVERED
	// - Iterate through the whole grid and paint each UNKNOWN cell COVERED if it isn't BLOCKED
	//   and it borders a COVERED cell
	// - If we have any cells that are neither BLOCKED nor COVERED, we have a splitter
	//Note that this algorithm will catch EVERY wall splitting case and is therefore
	//duplciative with the border walking algorithm implemented above.  However, leaving the one above
	//intact actually speeds up the code by detecting some wall splitters more quickly.

	//mergedBorder := b.Neighbors(mergedIslands)
	reachability := NewGrid(b.Problem.Width, b.Problem.Height)
	const UNK = 0
	const BLOCKED = 1
	const COVERED = 2
	unkCt := b.Problem.Width * b.Problem.Height
	for r := 0; r < b.Problem.Height; r++ {
		for c := 0; c < b.Problem.Width; c++ {
			coord = Coordinate{r, c}
			if mergedIslands.Contains(coord) {
				reachability[r][c] = BLOCKED
				unkCt--
			} else if b.IsOnEdge(coord) {
				reachability[r][c] = COVERED
				unkCt--
			} else {
				reachability[r][c] = UNK
			}
		}
	}

	keepGoing := true
	for keepGoing {
		keepGoing = false
		for r := 1; r < b.Problem.Height-1; r++ {
			for c := 1; c < b.Problem.Width-1; c++ {
				if reachability[r][c] != UNK {
					continue
				}
				if reachability[r-1][c] == COVERED || reachability[r+1][c] == COVERED || reachability[r][c-1] == COVERED || reachability[r][c+1] == COVERED {
					reachability[r][c] = COVERED
					unkCt--
					keepGoing = true
				}
			}
		}
		if unkCt == 0 {
			break
		}
	}

	return unkCt > 0
}
