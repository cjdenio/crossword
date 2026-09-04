package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"unicode"

	"github.com/cjdenio/crossword/puz"
	"golang.org/x/term"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Print("!!! PROGRAM CRASHED !!! Error:\r\n", r, "\r\n")
		}
	}()

	debugMode := flag.Bool("debug", false, "")
	flag.Parse()

	filename := flag.Arg(0)
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	puzzle, err := puz.ParsePuz(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\x1b[?25l")

	defer func() {
		fmt.Print("\x1b[?25h")
	}()
	termState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	defer term.Restore(int(os.Stdin.Fd()), termState)

	state := State{
		Puzzle:       puzzle,
		PuzzleState:  []rune(puzzle.State),
		SelectedClue: puzzle.Clues[0],
		SelectedCell: puzzle.Clues[0].Cells[0],
	}
	if debugMode != nil {
		state.DebugMode = *debugMode
	}

	uiHeight := state.RenderUI()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if len(data) == 0 {
			return 0, nil, bufio.ErrFinalToken
		}

		if data[0] == 0x1b {
			if len(data) < 3 {
				return 0, nil, nil
			}

			return 3, data[0:3], nil
		}

		return 1, data[0:1], nil
	})

	for {
		scanner.Scan()
		buffer := scanner.Bytes()

		state.LastKeySequence = fmt.Sprintf("%v", buffer)

		if buffer[0] == 3 {
			state.SelectedClue = nil
			state.SelectedCell = -1
			state.Goodbye = true
			fmt.Printf("\r\x1b[%dA", uiHeight)
			fmt.Print("\x1b[J")
			state.RenderUI()
			return
		}

		if buffer[0] == ' ' {
			switch state.SelectedClue.Direction {
			case puz.DirectionAcross:
				if state.Puzzle.Cells[state.SelectedCell][1] != nil {
					state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][1]
				}
			case puz.DirectionDown:
				if state.Puzzle.Cells[state.SelectedCell][0] != nil {
					state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][0]
				}
			}
		}

		if buffer[0] >= 0x61 && buffer[0] <= 0x7a {
			cellWasFilled := state.PuzzleState[state.SelectedCell] != '-'
			state.PuzzleState[state.SelectedCell] = unicode.ToUpper(rune(buffer[0]))

			i := slices.Index(state.SelectedClue.Cells, state.SelectedCell)

			if !cellWasFilled {
				// jump to next unfilled cell in clue
				for x := i + 1; x < len(state.SelectedClue.Cells); x++ {
					if state.PuzzleState[state.SelectedClue.Cells[x]] == '-' {
						state.SelectedCell = state.SelectedClue.Cells[x]
						break
					}
				}
			} else {
				// jump to next cell
				if i < len(state.SelectedClue.Cells)-1 {
					state.SelectedCell = state.SelectedClue.Cells[i+1]
				}
			}

			if state.GridFilled() {
				if state.CheckPuzzle() {
					state.SolveState = 2
				} else {
					state.SolveState = 1
				}
			} else {
				state.SolveState = 0
			}
		}

		if buffer[0] == 0x7f {
			// is there a filled cell underneath the cursor?
			if state.PuzzleState[state.SelectedCell] != '-' {
				state.PuzzleState[state.SelectedCell] = '-'
			} else {
				i := slices.Index(state.SelectedClue.Cells, state.SelectedCell)
				if i > 0 {
					state.PuzzleState[state.SelectedClue.Cells[i-1]] = '-' // clear the previous cell
					state.SelectedCell = state.SelectedClue.Cells[i-1]
				}
			}
		}

		if buffer[0] == 0x0d {
			state.NextWord()
		}

		if string(buffer[0:2]) == "\x1b[" {
			if (buffer[2] == 68 || buffer[2] == 67) && state.SelectedClue != nil && state.SelectedClue.Direction == puz.DirectionDown {
				state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][0]
			} else if (buffer[2] == 65 || buffer[2] == 66) && state.SelectedClue != nil && state.SelectedClue.Direction == puz.DirectionAcross {
				state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][1]
			} else {
				switch buffer[2] {
				case 68: // left
					state.MoveCursor(3)
				case 67: // right
					state.MoveCursor(1)
				case 65: // up
					state.MoveCursor(0)
				case 66: // down
					state.MoveCursor(2)
				}

				switch state.SelectedClue.Direction {
				case puz.DirectionAcross:
					state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][0]
				case puz.DirectionDown:
					state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][1]
				}
			}
		}

		fmt.Printf("\r\x1b[%dA", uiHeight)
		fmt.Print("\x1b[J")

		uiHeight = state.RenderUI()
	}
}
