package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"unicode"

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
func AnsiDimmed(s string) string {
	return fmt.Sprintf("\x1b[2m%s\x1b[0m", s)
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
			cell = AnsiDimmed("_")
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
	Puzzle          *puz.Puzzle
	PuzzleState     []rune
	SelectedCell    int
	SelectedClue    *puz.Clue
	Goodbye         bool
	LastKeySequence string
	DebugMode       bool
}

func (state *State) MoveCursor(direction int) {
	switch direction {
	case 0: // up
		if state.SelectedCell < state.Puzzle.Width { // if already in the top row, do nothing
			return
		}

		for i := state.SelectedCell - state.Puzzle.Width; i >= 0; i -= state.Puzzle.Width {
			if state.PuzzleState[i] != '.' {
				state.SelectedCell = i
				return
			}
		}
	case 1: // right
		if (state.SelectedCell+1)%state.Puzzle.Width == 0 { // if already in the right column, do nothing
			return
		}

		for i := state.SelectedCell + 1; i%state.Puzzle.Width != 0; i += 1 {
			if state.PuzzleState[i] != '.' {
				state.SelectedCell = i
				return
			}
		}
	case 2: // down
		if (state.SelectedCell) >= (state.Puzzle.Width*state.Puzzle.Height)-state.Puzzle.Width { // if already in the bottom row, do nothing
			return
		}

		for i := state.SelectedCell + state.Puzzle.Width; i < (state.Puzzle.Width * state.Puzzle.Height); i += state.Puzzle.Width {
			if state.PuzzleState[i] != '.' {
				state.SelectedCell = i
				return
			}
		}
	case 3: // left
		if state.SelectedCell%state.Puzzle.Width == 0 { // if already in the left column, do nothing
			return
		}

		for i := state.SelectedCell - 1; (i+1)%state.Puzzle.Width != 0; i -= 1 {
			if state.PuzzleState[i] != '.' {
				state.SelectedCell = i
				return
			}
		}
	}
}

func (state *State) NextWord() {
	if state.SelectedClue == nil {
		return
	}

	foundSelectedClue := false

	for _, clue := range state.Puzzle.Clues {
		if !foundSelectedClue && clue == state.SelectedClue {
			foundSelectedClue = true
			continue
		}

		if foundSelectedClue && clue.Direction == state.SelectedClue.Direction {
			state.SelectedClue = clue
			state.SelectedCell = clue.Cells[0]
			return
		}
	}

	for _, clue := range state.Puzzle.Clues {
		if clue.Direction != state.SelectedClue.Direction {
			state.SelectedClue = clue
			state.SelectedCell = clue.Cells[0]
			return
		}
	}
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
		if state.SelectedClue.Direction == puz.DirectionAcross {
			fmt.Printf("%d-across: %s\r\n\r\n", state.SelectedClue.Number, state.SelectedClue.Clue)
		} else {
			fmt.Printf("%d-down: %s\r\n\r\n", state.SelectedClue.Number, state.SelectedClue.Clue)
		}
		uiHeight += 2
	}
	fmt.Printf("%s\r\n", state.Puzzle.Copyright)
	uiHeight += 1

	if state.LastKeySequence != "" && state.DebugMode {
		fmt.Printf("\r\n%s\r\n", AnsiDimmed(state.LastKeySequence))
		uiHeight += 2
	}

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

	state := State{
		Puzzle:       puzzle,
		PuzzleState:  []rune(puzzle.State),
		SelectedClue: puzzle.Clues[0],
		SelectedCell: puzzle.Clues[0].Cells[0],
	}
	if debugMode != nil {
		state.DebugMode = *debugMode
	}

	uiHeight := RenderUI(&state)

	for {
		buffer := make([]byte, 3)
		_, err = os.Stdin.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}

		state.LastKeySequence = fmt.Sprintf("%v", buffer)

		if buffer[0] == 3 {
			state.SelectedClue = nil
			state.SelectedCell = -1
			state.Goodbye = true
			fmt.Printf("\r\x1b[%dA", uiHeight)
			fmt.Print("\x1b[J")
			RenderUI(&state)
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
			state.PuzzleState[state.SelectedCell] = unicode.ToUpper(rune(buffer[0]))
			// jump to next unfilled cell in clue
			i := slices.Index(state.SelectedClue.Cells, state.SelectedCell)
			for x := i + 1; x < len(state.SelectedClue.Cells); x++ {
				if state.PuzzleState[state.SelectedClue.Cells[x]] == '-' {
					state.SelectedCell = state.SelectedClue.Cells[x]
					break
				}
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

		fmt.Printf("\r\x1b[%dA", uiHeight)
		fmt.Print("\x1b[J")

		uiHeight = RenderUI(&state)
	}
}
