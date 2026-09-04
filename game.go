package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/cjdenio/crossword/puz"
)

type State struct {
	Puzzle          *puz.Puzzle
	PuzzleState     []rune
	SelectedCell    int
	SelectedClue    *puz.Clue
	Goodbye         bool
	LastKeySequence string
	DebugMode       bool
	SolveState      int
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

func (state *State) ClueFilled(clue *puz.Clue) bool {
	for _, cell := range clue.Cells {
		if state.PuzzleState[cell] == '-' {
			return false
		}
	}

	return true
}

func (state *State) FirstUnfilledCellForClue(clue *puz.Clue) int {
	for _, cell := range clue.Cells {
		if state.PuzzleState[cell] == '-' {
			return cell
		}
	}
	return clue.Cells[0]
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

		if foundSelectedClue && clue.Direction == state.SelectedClue.Direction && !state.ClueFilled(clue) {
			state.SelectedClue = clue
			state.SelectedCell = state.FirstUnfilledCellForClue(clue)
			return
		}
	}

	for _, clue := range state.Puzzle.Clues {
		if clue.Direction != state.SelectedClue.Direction && !state.ClueFilled(clue) {
			state.SelectedClue = clue
			state.SelectedCell = state.FirstUnfilledCellForClue(clue)
			return
		}
	}

	for _, clue := range state.Puzzle.Clues {
		if clue == state.SelectedClue {
			return
		} else if clue.Direction == state.SelectedClue.Direction && !state.ClueFilled(clue) {
			state.SelectedClue = clue
			state.SelectedCell = state.FirstUnfilledCellForClue(clue)
			return
		}
	}
}

func (state *State) PreviousWord() {
	if state.SelectedClue == nil {
		return
	}

	foundSelectedClue := false

	for i := len(state.Puzzle.Clues) - 1; i >= 0; i-- {
		clue := state.Puzzle.Clues[i]

		if !foundSelectedClue && clue == state.SelectedClue {
			foundSelectedClue = true
			continue
		}

		if foundSelectedClue && clue.Direction == state.SelectedClue.Direction && !state.ClueFilled(clue) {
			state.SelectedClue = clue
			state.SelectedCell = state.FirstUnfilledCellForClue(clue)
			return
		}
	}

	for i := len(state.Puzzle.Clues) - 1; i >= 0; i-- {
		clue := state.Puzzle.Clues[i]

		if clue.Direction != state.SelectedClue.Direction && !state.ClueFilled(clue) {
			state.SelectedClue = clue
			state.SelectedCell = state.FirstUnfilledCellForClue(clue)
			return
		}
	}

	for i := len(state.Puzzle.Clues) - 1; i >= 0; i-- {
		clue := state.Puzzle.Clues[i]

		if clue == state.SelectedClue {
			return
		} else if clue.Direction == state.SelectedClue.Direction && !state.ClueFilled(clue) {
			state.SelectedClue = clue
			state.SelectedCell = state.FirstUnfilledCellForClue(clue)
			return
		}
	}
}

func (state *State) GridFilled() bool {
	return !slices.Contains(state.PuzzleState, '-')
}

func (state *State) CheckPuzzle() bool {
	for i, cell := range state.PuzzleState {
		if cell == '.' {
			continue
		}
		if cell != rune(state.Puzzle.Solution[i]) {
			return false
		}
	}
	return true
}

func (state *State) RenderUI() int {
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

	switch state.SolveState {
	case 1:
		fmt.Printf("\r\n%s\r\n", AnsiRed("The puzzle was filled, but at least 1 letter is incorrect..."))
		uiHeight += 2
	case 2:
		fmt.Printf("\r\n%s\r\n", AnsiGreen("The puzzle was solved!"))
		uiHeight += 2
	}

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
func AnsiGreen(s string) string {
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", s)
}
func AnsiRed(s string) string {
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", s)
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

type SaveFile struct {
	State string `json:"state"`
}

func (state *State) CreateSaveFile() ([]byte, error) {
	return json.Marshal(SaveFile{
		State: string(state.PuzzleState),
	})
}

func (state *State) LoadSaveFile(f []byte) error {
	var save SaveFile
	err := json.Unmarshal(f, &save)
	if err != nil {
		return err
	}

	// verify the save file matches the shape of the puzzle
	if len(save.State) != state.Puzzle.Width*state.Puzzle.Height {
		return errors.New("save is invalid")
	}

	for i, cell := range state.Puzzle.Solution {
		if (save.State[i] == '.' && cell != '.') || (save.State[i] != '.' && cell == '.') {
			return errors.New("save is invalid")
		}
	}

	state.PuzzleState = []rune(save.State)

	return nil
}
