package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/cjdenio/crossword/puz"
	"golang.org/x/term"
)

const (
	AnsiInvert          string = "\x1b[7m"
	AnsiReset           string = "\x1b[m"
	AnsiWhiteBackground string = "\x1b[100m"
)

func AnsiInverted(s string) string {
	return AnsiInvert + s + AnsiReset
}
func AnsiWhiteBackgrounded(s string) string {
	return AnsiWhiteBackground + s + AnsiReset
}

func RenderPuzzle(puzzle []rune, width, height int, selectedClueCells []int, selectedCell int) string {
	b := strings.Builder{}

	b.WriteRune('┌')
	for range (width * 2) + 1 {
		b.WriteRune('─')
	}
	b.WriteString("┐\r\n")

	for index, char := range puzzle {
		if index%width == 0 {
			b.WriteString("│ ")
		}

		cell := ""
		switch char {
		case '.':
			cell = string(rune(0x2588))
		case '-':
			cell = "_"
		default:
			cell = string(char)
		}

		if selectedCell == index {
			cell = AnsiInverted(cell)
		} else if slices.Contains(selectedClueCells, index) {
			cell = AnsiWhiteBackgrounded(cell)
		}

		b.WriteString(cell)

		if (index+1)%width == 0 {
			b.WriteString(" │\r\n")
		} else if char == '.' && puzzle[index+1] == '.' {
			b.WriteRune(0x2588)
		} else {
			// inefficient
			if slices.Contains(selectedClueCells, index) && slices.Contains(selectedClueCells, index+1) {
				b.WriteString(AnsiWhiteBackgrounded(" "))
			} else {
				b.WriteRune(' ')
			}
		}
	}

	b.WriteRune('└')
	for range (width * 2) + 1 {
		b.WriteRune('─')
	}
	b.WriteString("┘\r\n")

	return b.String()
}

type State struct {
	Puzzle       *puz.Puzzle
	PuzzleState  []rune
	SelectedCell int
	SelectedClue *puz.Clue
	Goodbye      bool
}

func RenderUI(state *State) int {
	selectedClueCells := []int{}
	if state.SelectedClue != nil {
		selectedClueCells = state.SelectedClue.Cells
	}

	uiHeight := 0

	fmt.Printf("\r\nTITLE: %s\r\n", state.Puzzle.Title)
	uiHeight += 2
	fmt.Printf("AUTHOR: %s\r\n", state.Puzzle.Author)
	uiHeight += 1
	fmt.Print(RenderPuzzle(state.PuzzleState, state.Puzzle.Width, state.Puzzle.Height, selectedClueCells, state.SelectedCell) + "\r\n")
	uiHeight += state.Puzzle.Height + 3
	if state.SelectedClue != nil {
		fmt.Printf("%d: %s\r\n\r\n", state.SelectedClue.Number, state.SelectedClue.Clue)
		uiHeight += 2
	}
	fmt.Printf("%s\r\n", state.Puzzle.Copyright)
	uiHeight += 1

	if state.Goodbye {
		fmt.Print("\r\nsee ya\r\n")
		uiHeight += 2
	}

	return uiHeight
}

func main() {
	fmt.Print("\x1b[?25l")

	defer func() {
		fmt.Print("\x1b[?25h")
	}()
	termState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	defer term.Restore(int(os.Stdin.Fd()), termState)

	filename := os.Args[1]
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	puzzle, err := puz.ParsePuz(file)
	if err != nil {
		log.Fatal(err)
	}

	state := State{
		Puzzle:       puzzle,
		PuzzleState:  []rune(puzzle.State),
		SelectedClue: puzzle.Clues[0],
		SelectedCell: puzzle.Clues[0].Cells[0],
	}

	uiHeight := RenderUI(&state)

	for {
		buffer := make([]byte, 3)
		_, err = os.Stdin.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}

		if buffer[0] == 3 {
			state.SelectedClue = nil
			state.SelectedCell = -1
			state.Goodbye = true
			fmt.Printf("\r\x1b[%dA", uiHeight)
			fmt.Print("\x1b[J")
			RenderUI(&state)
			return
		}

		if (buffer[2] == 68 || buffer[2] == 67) && state.SelectedClue != nil && state.SelectedClue.Direction == puz.DirectionDown {
			state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][0]
		} else if (buffer[2] == 65 || buffer[2] == 66) && state.SelectedClue != nil && state.SelectedClue.Direction == puz.DirectionAcross {
			state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][1]
		} else {
			switch buffer[2] {
			case 68: // left
				state.SelectedCell--
			case 67: // right
				state.PuzzleState[state.SelectedCell] = 'a'
				state.SelectedCell++
			case 65: // up
				state.SelectedCell -= puzzle.Width
			case 66: // down
				state.SelectedCell += puzzle.Width
			}

			switch state.SelectedClue.Direction {
			case puz.DirectionAcross:
				state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][0]
			case puz.DirectionDown:
				state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][1]
			}
		}

		fmt.Printf("\r\x1b[%dA", uiHeight)
		fmt.Print("\x1b[J")

		uiHeight = RenderUI(&state)
	}
}
