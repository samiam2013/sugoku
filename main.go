package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
)

type board struct {
	box [81]int
}

func main() {
	fmt.Printf("%#v\n", generateBoard(30))
}

func generateBoard(emptyCount int) board {
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
	return brd
}
