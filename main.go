package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	b := GetBoardFromFile("problem3.txt")

	startNano := time.Now().UnixNano()

	//bc := GetBoardFromFile("problem3-solved.txt")
	//fmt.Printf("Got board BC: %v\n", bc)
	//bc.PopulateIslandPossibilities()
	//fmt.Printf("BC: %v\n", bc)
	s := Solver{b, nil, "", make(chan ProgressUpdate, b.Problem.Size*2)}
	var wg sync.WaitGroup
	wg.Add(1)
	go PrintUpdates(&s, &wg)
	s.InitSolve()
	s.AutoSolve(true, false)
	close(s.Progress)
	wg.Wait()
	fmt.Printf("%v\n", b.String())
	stopNano := time.Now().UnixNano()
	if sol, reason := s.b.IsSolved(); sol == false {
		fmt.Printf("Not solved (%v)\n", reason)
	}
	fmt.Printf("Total duration: %.4f\n", float64(stopNano-startNano)/1000000000.0)
}

// TODO: have group versions of RemoveFromPossibility and MarkPainted - only one trip through the possibility sets
// TODO: check only four neighbors for MergeWallIslands and MergeIslands (could save ~10% of problem 3 time)
