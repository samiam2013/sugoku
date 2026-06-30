package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	brd := generateBoard(60)
	brd.print()
	// fmt.Printf("%#v\n", brd)
	fmt.Printf("is board valid?: %t\n", brd.validate())
}

type board struct {
	box [81]int
}

func (b *board) print() {
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

func printLine() {
	for range 9 {
		fmt.Print("+---")
	}
	fmt.Println("+")
}

func (b *board) validate() bool {
	startT := time.Now()
	// check each line
	// fmt.Println("checking each line")
	for i := range 9 {
		if !validateSeries(b.box[(i * 9):((i * 9) + 9)]) {
			return false
		}
	}
	// check each column
	// fmt.Println("checking each column")
	for i := range 9 {
		col := [9]int{}
		for j := range 9 {
			col[j] = b.box[i+(j*9)]
		}
		if !validateSeries(col[:]) {
			return false
		}
	}
	// check each house
	// fmt.Println("checking each house")
	for i := range 3 {
		for j := range 3 {
			house := []int{}
			for k := range 3 {
				startIdx := (k * 9) + (j * 9) + (i * 3)
				endIdx := startIdx + 3
				house = append(house, b.box[startIdx:endIdx]...)
			}
			if !validateSeries(house) {
				return false
			}
		}
	}
	fmt.Printf("validated in %s\n", time.Since(startT))
	return true
}

func validateSeries(nineBoxes []int) bool {
	if len(nineBoxes) < 9 {
		panic("less than 9 boxes in series to validate!")
	}
	if len(nineBoxes) > 9 {
		panic("more than 9 boxes in series to validate!")
	}
	freq := make(map[int]int, 9)
	for _, val := range nineBoxes {
		if f, ok := freq[val]; !ok && val != 0 {
			freq[val] = 1
		} else if f > 1 && val != 0 {
			return false
		}
	}
	// fmt.Printf("%#v\n", freq)
	return true
}

func generateBoard(emptyCount int) board {
	startT := time.Now()
	cmd := exec.Command("python3", "./g4gGenerator/main.py", strconv.Itoa(emptyCount))
	b, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to run python sudoku generator", "error", err)
	}
	lines := bytes.Split(bytes.TrimRight(b, "\n"), []byte("\n"))
	brd := board{}
	i := 0
	for _, line := range lines {
		values := bytes.Split(line, []byte(","))
		for _, value := range values {
			if len(value) > 1 {
				panic("value longer than a single character")
			}
			brd.box[i] = int(value[0]) - 48
			i++
		}
	}
	fmt.Printf("board generation took %s\n", time.Since(startT))
	return brd
}
