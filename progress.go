package main

import (
	"fmt"
	"sync"
	"time"
)

func PrintUpdates(s *Solver, wg *sync.WaitGroup) {
	defer wg.Done()
	if s.Progress == nil {
		return
	}
	fmt.Println("Starting...")
	for {
		select {
		case update, ok := <-s.Progress:
			if !ok {
				return
			}
			s := ""
			pct := float64(update.TotalMarked) / float64(update.GridSize)
			for i := 0.05; i <= 1.0; i += 0.05 {
				if pct >= i {
					s += "="
				} else {
					s += "."
				}
			}
			s = "[" + s + "]"
			s += fmt.Sprintf(" %d/%d (%s)", update.TotalMarked, update.GridSize, update.CurrentAction)
			fmt.Print("\033[1A\033[K")
			fmt.Printf("%s\n", s)
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}
