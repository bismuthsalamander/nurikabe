package nurigobe

import (
	"fmt"
	"os"
	"strings"
)

func islandSpecChar(sz int) string {
	if sz < 10 {
		return string(rune(sz + '0'))
	}
	if sz < 36 {
		return string(rune((sz - 10) + 'a'))
	}
	if sz < 62 {
		return string(rune((sz - 36) + 'A'))
	}
	return "?"
}

func parseIslandSpecChar(c rune) int {
	if c >= '1' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'z' {
		return int((c - 'a') + 10)
	}
	if c >= 'A' && c <= 'W' {
		return int((c - 'A') + 36)
	}
	return -1
}

func DefFromString(input string) ProblemDef {
	prob := ProblemDef{}
	lines := make([]string, 0)
	for _, txt := range strings.Split(input, "\n") {
		txt = strings.Trim(txt, "\r\n")
		if len(txt) > 0 {
			lines = append(lines, txt)
		}
	}
	prob.Width = len(lines[0])
	prob.Height = len(lines)
	prob.Size = prob.Width * prob.Height
	for ri, row := range lines {
		for ci, cell := range row {
			count := parseIslandSpecChar(cell)
			if count > -1 {
				prob.IslandSpecs = append(prob.IslandSpecs, IslandSpec{ci, ri, count})
				prob.TargetWallCount += count
			}
		}
	}
	return prob
}

func BoardFromString(input string) *Board {
	def := DefFromString(input)
	b := Board{def, NewGrid(def.Width, def.Height), NewGrid(def.Width, def.Height), make([]*Island, 0), make([]*Island, 0), make([]*CoordinateSet, 0), 0}
	for _, spec := range b.Problem.IslandSpecs {
		b.Grid[spec.Row][spec.Col] = CLEAR
		b.TotalMarked++
		b.Islands = append(b.Islands, MakeRootedIsland(spec.Row, spec.Col, spec.Size))
		b.DiagonalSets = append(b.DiagonalSets, SingleCoordinateSet(Coordinate{spec.Row, spec.Col}))
	}
	lines := make([]string, 0)
	for _, txt := range strings.Split(input, "\n") {
		txt = strings.Trim(txt, "\r\n")
		if len(txt) > 0 {
			lines = append(lines, txt)
		}
	}
	for ri, row := range lines {
		for ci, cell := range row {
			if cell == 'X' {
				b.MarkPainted(ri, ci)
			} else if cell == '.' {
				b.MarkClear(ri, ci)
			}
		}
	}
	return &b
}

func GetBoardFromFile(f string) *Board {
	data, err := os.ReadFile(f)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil
	}
	return BoardFromString(string(data))
}
