package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
)

func main() {
	brd := generateBoard(37)
	brd.print()
	for brd.solveHiddenSingles() != 0 {
		continue
	}
	brd.print()
	fmt.Printf("Puzzle validity: %t\n", brd.valid())
	fmt.Printf("Puzzle filled: %t\n", brd.filled())
}

type board struct {
	box [81]int // 0 for empty spaces
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
			houseOcc := b.houseOccupants(i, j)
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
			houseOcc := b.houseOccupants(i, j)
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
				b.box[i] = poss[0]
				solved++
			}
		}
	}
	return solved
}

func (b *board) evalPossible(idx int) (possibilities []int) {
	colIdx := idx % 9
	rowIdx := (idx - (colIdx)) / 9
	houseRowIdx := (rowIdx - (rowIdx % 3)) / 3
	houseColIdx := (colIdx - (colIdx % 3)) / 3
	occupants := b.houseOccupants(houseRowIdx, houseColIdx)
	occupants = append(occupants, b.rowOccupants(rowIdx)...)
	occupants = append(occupants, b.colOccupants(colIdx)...)
	uniqueOccupants := make(map[int]struct{})
	for _, occupant := range occupants {
		if _, ok := uniqueOccupants[occupant]; !ok {
			uniqueOccupants[occupant] = struct{}{}
		}
	}
	for i := 1; i <= 9; i++ {
		if _, ok := uniqueOccupants[i]; !ok {
			possibilities = append(possibilities, i)
		}
	}
	return possibilities
}

func (b *board) rowOccupants(rowIdx int) (occupants []int) {
	for i := rowIdx * 9; i < (rowIdx*9)+9; i++ {
		if b.box[i] != 0 {
			occupants = append(occupants, b.box[i])
		}
	}
	return occupants
}

func (b *board) colOccupants(colIdx int) (occupants []int) {
	for i := colIdx; i < 81; i += 9 {
		if b.box[i] != 0 {
			occupants = append(occupants, b.box[i])
		}
	}
	return occupants
}

// get the occupants of a house given any index that falls inside that house
func (b *board) houseOccupants(houseRowIdx, houseColIdx int) (occupants []int) {
	startingRow := houseRowIdx * 3
	startingCol := houseColIdx * 3
	for rowOffset := range 3 {
		startIdx := (startingRow + rowOffset) * 9
		for colOffset := range 3 {
			idxToCheck := startIdx + startingCol + colOffset
			if val := b.box[idxToCheck]; val != 0 {
				occupants = append(occupants, val)
			}
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
