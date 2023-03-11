package main

import (
    "os"
	"fmt"
	"sync"
	"time"
    "github.com/bismuthsalamander/nurikabe/nurigobe"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Printf("usage: %s [problem.txt]\n", os.Args[0])
        return
    }

    fn := os.Args[1]
	b, err := nurigobe.GetBoardFromFile(fn)
    if err != nil {
        fmt.Printf("error reading problem file %s: %v\n", fn, err)
        return
    }

	startNano := time.Now().UnixNano()
	s := nurigobe.NewSolver(b)
	var wg sync.WaitGroup
	wg.Add(1)
	go s.PrintUpdates(&wg)
	s.InitSolve()
	s.AutoSolve(true, false)
	close(s.Progress)
	wg.Wait()
	fmt.Printf("%v\n", b.String())
	stopNano := time.Now().UnixNano()
	if sol, reason := b.IsSolved(); sol == false {
		fmt.Printf("Not solved (%v)\n", reason)
	}
	fmt.Printf("Total duration: %.4f\n", float64(stopNano-startNano)/1000000000.0)
}

// TODO: have group versions of RemoveFromPossibility and MarkPainted - only one trip through the possibility sets
// TODO: check only four neighbors for MergeWallIslands and MergeIslands (could save ~10% of problem 3 time)
