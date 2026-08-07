package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
)

func main() {
	// 17 filled spaces required to necessitate some number of solutions
	emptySpaces := 81 - 25
	b := generateBoard(emptySpaces)
	b.print()
	b.backtrackSolve()
	b.print()
	fmt.Printf("Puzzle validity: %t\n", b.valid())
	fmt.Printf("Puzzle filled: %t\n", b.filled())
}

func (b *board) backtrackSolve() {
	for {
		b.deductiveSolve()
		if !b.valid() {
			// clear all the current guess step spots
			b.clearGuesses(b.guessStep)
			// get the current guess in the guess step before that
			b.guessStep--
			guessIdx, guessVal := b.findGuess(b.guessStep)
			b.box[guessIdx] = 0
			b.shadow[guessIdx] = 0
			fmt.Println("previous guess:", guessVal, "index:", guessVal)
			b.guessStep--
			// guess the next possible
			poss := b.evalPossible(guessIdx)
			if len(poss) == 0 {
				panic("no guesses possible at index")
				// break
			}
			for _, p := range poss {
				if p > guessVal {
					b.shadow[guessIdx] = b.guessStep
					b.box[guessIdx] = guessVal
					b.guessStep++
					break
				}
			}
		}
		if b.filled() {
			return
		}
		if !b.valid() {
			panic("bad after backtrack re-guess")
		}

		fmt.Println("trying backtrack")
		// time to guess
		b.guessStep++
		// find the first open spot
		for i := range 81 {
			if b.box[i] == 0 {
				poss := b.evalPossible(i)
				if len(poss) == 0 {
					b.print()
					fmt.Println("no possibilities at index", i)
					continue
					//break // dead end here
				}
				// guess the first possible value
				b.box[i] = poss[0]
				// set the guess in the shadow
				b.shadow[i] = b.guessStep
				// increment guess step again for next deductions
				b.guessStep++
			}
		}
	}
}

func (b *board) clearGuesses(guessStep int) {
	for i := range 81 {
		if b.shadow[i] == guessStep {
			b.box[i] = 0
			b.shadow[i] = 0
		}
	}
}

// returns index, value
func (b *board) findGuess(guessStep int) (int, int) {
	for i := range 81 {
		if b.shadow[i] == guessStep {
			return i, b.box[i]
		}
	}
	return -1, 0
}

func (b *board) deductiveSolve() {
	for {
		// if !b.valid() {
		// 	panic("hit invalid state in deductive solve")
		// }
		for b.solveHiddenSingles() != 0 {
			continue
		}
		if b.filled() {
			return
		}
		// if !b.valid() {
		// 	panic("hit invalid state in deductive solve 2")
		// }
		solvedOne := b.solveHiddenDouble()
		// if !b.valid() {
		// 	panic("hit invalid state in deductive solve 3")
		// }
		if solvedOne && b.filled() {
			return
		} else if !solvedOne {
			return
		} // else, continue
	}
}

type board struct {
	box [81]int // 0 for empty spaces

	// 0 for deduction only,
	// 1 for guess #1, 2 for guess #1 deduction,
	// 3 for guess #2, 4 for guess #2 deduction, etc
	guessStep int
	shadow    [81]int
}

func (b *board) solveHiddenDouble() bool {
	for i := range 81 {
		if b.box[i] == 0 {
			possible := b.evalPossible(i)
			if len(possible) > 0 && len(possible) == 2 {
				b.box[i] = possible[0]
				if b.guessStep > 0 {
					b.shadow[i] = b.guessStep
				}
				if !b.valid() {
					panic("invalid double hidden solve")
				}
				return true
			}
		}
	}
	return false
}

func (b *board) filled() bool {
	for i := range 9 {
		rowOcc := b.rowOccupants(i)
		colOcc := b.colOccupants(i)
		if len(rowOcc) < 9 || len(colOcc) < 9 {
			return false
		}
	}
	for i := range 3 {
		for j := range 3 {
			// todo stop doing this cursed thing
			houseOcc := b.houseOccupants(absoluteIndiciesByHouse(i, j)[0])
			if len(houseOcc) < 9 {
				return false
			}
		}
	}
	return true
}

func (b *board) valid() bool {
	for i := range 9 {
		rowOcc := b.rowOccupants(i)
		rowFreq := freq(rowOcc)
		if hasOverlap(rowFreq) {
			return false
		}
		colOcc := b.colOccupants(i)
		colFreq := freq(colOcc)
		if hasOverlap(colFreq) {
			return false
		}
	}
	for i := range 3 {
		for j := range 3 {
			// todo stop doing this cursed thing
			houseOcc := b.houseOccupants(absoluteIndiciesByHouse(i, j)[0])
			houseFreq := freq(houseOcc)
			if hasOverlap(houseFreq) {
				return false
			}
		}
	}

	return true
}

func hasOverlap(frequency map[int]int) bool {
	for _, ct := range frequency {
		if ct > 1 {
			return true
		}
	}
	return false
}

func freq(occupants []int) map[int]int {
	frequency := make(map[int]int)
	for _, occupant := range occupants {
		if _, ok := frequency[occupant]; !ok {
			frequency[occupant] = 1
		} else {
			frequency[occupant]++
		}
	}
	return frequency
}

func (b *board) solveHiddenSingles() (solved int) {
	for i, v := range b.box {
		if v == 0 {
			if poss := b.evalPossible(i); len(poss) == 1 {
				if b.guessStep > 0 {
					b.box[i] = b.guessStep
				}
				b.box[i] = poss[0]
				if !b.valid() {
					panic("invalid single hidden solve")
				}
				solved++
			}
		}
	}
	return solved
}

func (b *board) evalPossible(idx int) (possibilities []int) {
	occupants := b.houseOccupants(idx)
	occupants = append(occupants, b.colOccupants(idx)...)
	occupants = append(occupants, b.rowOccupants(idx)...)
	uniqueOccupants := mapUniqueOccupants(occupants)
	for i := 1; i <= 9; i++ {
		if _, ok := uniqueOccupants[i]; !ok {
			possibilities = append(possibilities, i)
		}
	}
	return possibilities
}

func (b *board) rowOccupants(idx int) (occupants []int) {
	rowIdx := (idx - (idx % 9)) / 9
	for i := rowIdx * 9; i < (rowIdx*9)+9; i++ {
		if b.box[i] != 0 {
			occupants = append(occupants, b.box[i])
		}
	}
	return occupants
}

func (b *board) colOccupants(idx int) (occupants []int) {
	colIdx := idx % 9
	for i := colIdx; i < 81; i += 9 {
		if b.box[i] != 0 {
			occupants = append(occupants, b.box[i])
		}
	}
	return occupants
}

func houseIndicesByAbsolute(idx int) (houseRow, houseCol int) {
	colIdx := idx % 9
	rowIdx := (idx - (colIdx)) / 9
	houseRow = (rowIdx - (rowIdx % 3)) / 3
	houseCol = (colIdx - (colIdx % 3)) / 3
	return
}

func absoluteIndiciesByHouse(houseRow, houseCol int) [9]int {
	absIndicies := [9]int{}
	startingRow := houseRow * 3
	startingCol := houseCol * 3
	i := 0
	for rowOffset := range 3 {
		startIdx := (startingRow + rowOffset) * 9
		for colOffset := range 3 {
			absIndicies[i] = startIdx + startingCol + colOffset
			i++
		}
	}
	return absIndicies
}

// get the occupants of a house given any index that falls inside that house
func (b *board) houseOccupants(idx int) (occupants []int) {
	absIdxs := absoluteIndiciesByHouse(houseIndicesByAbsolute(idx))
	for _, i := range absIdxs {
		if val := b.box[i]; val != 0 {
			occupants = append(occupants, val)
		}
	}
	return occupants
}

func (b *board) print() {
	printLine := func() {
		for range 9 {
			fmt.Print("+---")
		}
		fmt.Println("+")
	}
	for i := range 9 {
		printLine()
		fmt.Print("| ")
		for _, num := range b.box[(i * 9):((i + 1) * 9)] {
			if num == 0 {
				fmt.Print("  | ")
			} else {
				fmt.Printf("%d | ", num)
			}
		}
		fmt.Println()
	}
	printLine()
}

func generateBoard(emptyCount int) (brd board) {
	cmd := exec.Command("python3", "./g4gGenerator/main.py", strconv.Itoa(emptyCount))
	b, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to run python sudoku generator", "error", err)
	}
	for row, line := range bytes.Split(bytes.TrimRight(b, "\n"), []byte("\n")) {
		for col, value := range bytes.Split(line, []byte(",")) {
			brd.box[(row*9)+col] = int(value[0]) - 48
		}
	}
	return brd
}

func mapUniqueOccupants(occupants []int) map[int]struct{} {
	uniqueOccupants := make(map[int]struct{})
	for _, occupant := range occupants {
		if _, ok := uniqueOccupants[occupant]; !ok {
			uniqueOccupants[occupant] = struct{}{}
		}
	}
	return uniqueOccupants
}
