package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode"

	"github.com/cjdenio/crossword/puz"
)

type SolveState int

const (
	Unsolved SolveState = iota
	FilledNotSolved
	Solved
)

type UIState int

const (
	StateGame UIState = iota
	StateRebusConfirmation
	StateMenu
	StateRevealMenu
	StateCheckMenu
)

type State struct {
	Puzzle          *puz.Puzzle
	PuzzleState     []rune
	SelectedCell    int
	SelectedClue    *puz.Clue
	LastKeySequence string
	DebugMode       bool
	SolveState      SolveState
	CheckState      []rune
	UIState         UIState
}

func RenderWithLineWrap(s string, width int) (out string, lines int) {
	builder := strings.Builder{}
	runes := []rune(s)

	counter := 0
	lines = 1

	for i := 0; i < len(runes); i++ {
		char := runes[i]

		if char == '\n' {
			builder.WriteString("\r\n")
			lines++
			counter = 0
			continue
		}

		if char == '\r' && i < len(runes)-1 && runes[i+1] == '\n' {
			builder.WriteString("\r\n")
			lines++
			counter = 0
			i++
			continue
		}

		builder.WriteRune(char)
		counter++

		if counter >= width && i < len(runes)-1 {
			lines++
			builder.WriteString("\r\n")
			counter = 0
		}
	}

	builder.WriteString("\r\n")

	out = builder.String()
	return
}

func NewState(puzzle *puz.Puzzle) *State {
	state := StateGame
	if puzzle.HasRebus {
		state = StateRebusConfirmation
	}
	return &State{
		Puzzle:       puzzle,
		PuzzleState:  []rune(puzzle.State),
		SelectedClue: puzzle.Clues[0],
		SelectedCell: puzzle.Clues[0].Cells[0],
		CheckState:   make([]rune, len(puzzle.State)),
		UIState:      state,
	}
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

func (state *State) PuzzleSolved() bool {
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

func (state *State) CheckPuzzle() {
	result := make([]rune, len(state.PuzzleState))

	for i, cell := range state.PuzzleState {
		if cell == '.' || cell == '-' {
			continue
		}
		if cell == rune(state.Puzzle.Solution[i]) {
			result[i] = 'y'
		} else {
			result[i] = 'n'
		}
	}

	state.CheckState = result
}

func (state *State) CheckWord(clue *puz.Clue) {
	for i, cellIdx := range clue.Cells {
		cellState := state.PuzzleState[cellIdx]
		if cellState == '-' || cellState == '.' {
			continue
		}

		if cellState == rune(clue.Solution[i]) {
			state.CheckState[cellIdx] = 'y'
		} else {
			state.CheckState[cellIdx] = 'n'
		}
	}
}

func (state *State) RevealWord(clue *puz.Clue) {
	for i, cellIdx := range clue.Cells {
		state.PuzzleState[cellIdx] = rune(clue.Solution[i])
		state.CheckState[cellIdx] = 'y'
	}
}

func (state *State) RevealPuzzle() {
	for i, cell := range state.PuzzleState {
		if cell == '.' {
			continue
		}
		state.PuzzleState[i] = rune(state.Puzzle.Solution[i])
		state.CheckState[i] = 'y'
	}
}

func (state *State) RenderUI(w io.Writer, width int) int {
	if state.UIState == StateRebusConfirmation {
		fmt.Fprint(w, "This puzzle contains a rebus. This tool does not support rebuses.\r\nPress 'y' to continue anyway, or press any other key to exit.\r\n")
		return 2
	}

	uiHeight := 0

	fmt.Fprintf(w, "\r\nTITLE: %s\r\n", state.Puzzle.Title)
	uiHeight += 2
	fmt.Fprintf(w, "AUTHOR: %s\r\n", state.Puzzle.Author)
	uiHeight += 1
	fmt.Fprint(w, state.RenderPuzzle()+"\r\n")
	uiHeight += state.Puzzle.Height + 3
	if state.SelectedClue != nil {
		directionStr := "down"
		if state.SelectedClue.Direction == puz.DirectionAcross {
			directionStr = "across"
		}

		clueStr, clueLines := RenderWithLineWrap(fmt.Sprintf("%d-%s: %s", state.SelectedClue.Number, directionStr, state.SelectedClue.Clue), width)
		w.Write([]byte(clueStr))
		uiHeight += clueLines
		io.WriteString(w, "\r\n")
		uiHeight += 1
	}
	fmt.Fprintf(w, "%s\r\n", state.Puzzle.Copyright)
	uiHeight += 1

	switch state.UIState {
	case StateGame:
		fmt.Fprintln(w, AnsiDimmed("\r\n[enter]/[tab]: next word | [ctrl+a]: menu | [ctrl+c]: exit"))
		uiHeight += 2
	case StateMenu:
		fmt.Fprintln(w, AnsiDimmed("\r\n[c]: check puzzle/word | [r]: reveal puzzle/word | [x]: clear grid | [q]: exit menu"))
		uiHeight += 2
	case StateCheckMenu:
		fmt.Fprintln(w, AnsiDimmed("\r\n[w]: check word | [p]: check puzzle | [q]: exit menu"))
		uiHeight += 2
	case StateRevealMenu:
		fmt.Fprintln(w, AnsiDimmed("\r\n[w]: reveal word | [p]: reveal puzzle | [q]: exit menu"))
		uiHeight += 2
	}

	switch state.SolveState {
	case FilledNotSolved:
		fmt.Fprintf(w, "\r\n%s\r\n", AnsiRed("The puzzle was filled, but at least 1 letter is incorrect..."))
		uiHeight += 2
	case Solved:
		fmt.Fprintf(w, "\r\n%s\r\n", AnsiGreen("The puzzle was solved!"))
		uiHeight += 2
	}

	if state.LastKeySequence != "" && state.DebugMode {
		fmt.Fprintf(w, "\r\n%s\r\n", AnsiDimmed(state.LastKeySequence))
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

func (state *State) RenderPuzzle() string {
	selectedClueCells := []int{}
	if state.SelectedClue != nil {
		selectedClueCells = state.SelectedClue.Cells
	}

	b := strings.Builder{}

	b.WriteRune('┌')
	for range (state.Puzzle.Width * 2) + 1 {
		b.WriteRune('─')
	}
	b.WriteString("┐\r\n")

	for index, char := range state.PuzzleState {
		if index%state.Puzzle.Width == 0 {
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

		if state.SelectedCell == index {
			cell = AnsiInverted(cell)
		} else if slices.Contains(selectedClueCells, index) {
			cell = AnsiWhiteBackgrounded(cell)
		}

		if len(state.CheckState) == len(state.PuzzleState) {
			switch state.CheckState[index] {
			case 'y':
				cell = AnsiGreen(cell)
			case 'n':
				cell = AnsiRed(cell)
			}
		}

		b.WriteString(cell)

		if (index+1)%state.Puzzle.Width == 0 {
			b.WriteString(" │\r\n")
		} else if char == '.' && state.PuzzleState[index+1] == '.' {
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
	for range (state.Puzzle.Width * 2) + 1 {
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

func (state *State) SwitchDirections() bool {
	switch state.SelectedClue.Direction {
	case puz.DirectionAcross:
		if state.Puzzle.Cells[state.SelectedCell][1] != nil {
			state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][1]
		}
		return true
	case puz.DirectionDown:
		if state.Puzzle.Cells[state.SelectedCell][0] != nil {
			state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][0]
		}
		return true
	}

	return false
}

func (state *State) ClearGrid() {
	for i, cell := range state.PuzzleState {
		if cell != '.' {
			state.PuzzleState[i] = '-'
			state.CheckState[i] = '\x00'
		}
	}
}

func (state *State) HandleInput(buffer []byte) (exited bool) {
	if len(buffer) == 0 {
		return false
	}

	state.LastKeySequence = fmt.Sprintf("%v", buffer)

	if buffer[0] == 3 {
		return true
	}

	switch state.UIState {
	case StateRebusConfirmation:
		if buffer[0] == 'y' {
			state.UIState = StateGame
			return false
		} else {
			return true
		}
	case StateMenu:
		switch buffer[0] {
		case 'c':
			state.UIState = StateCheckMenu
		case 'r':
			state.UIState = StateRevealMenu
		case 'x':
			state.ClearGrid()
			state.UIState = StateGame
		default:
			state.UIState = StateGame
		}
		return false
	case StateCheckMenu:
		switch buffer[0] {
		case 'w':
			state.CheckWord(state.SelectedClue)
		case 'p':
			state.CheckPuzzle()
		}
		state.UIState = StateGame
		return false
	case StateRevealMenu:
		switch buffer[0] {
		case 'w':
			state.RevealWord(state.SelectedClue)
		case 'p':
			state.RevealPuzzle()
			return true
		}
		state.UIState = StateGame
		return false
	}

	if buffer[0] == '\x01' {
		state.UIState = StateMenu
		return false
	}

	if buffer[0] == ' ' {
		state.SwitchDirections()
	}

	if buffer[0] >= 0x61 && buffer[0] <= 0x7a {
		cellWasFilled := state.PuzzleState[state.SelectedCell] != '-'
		state.PuzzleState[state.SelectedCell] = unicode.ToUpper(rune(buffer[0]))
		state.CheckState[state.SelectedCell] = 0x00

		i := slices.Index(state.SelectedClue.Cells, state.SelectedCell)

		if !cellWasFilled {
			// jump to next unfilled cell in clue
			for x := (i + 1) % len(state.SelectedClue.Cells); x != i; x = ((x + 1) % len(state.SelectedClue.Cells)) {
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
			if state.PuzzleSolved() {
				state.SolveState = Solved
				return true
			} else {
				state.SolveState = FilledNotSolved
			}
		} else {
			state.SolveState = Unsolved
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

	if buffer[0] == '\r' || buffer[0] == '\t' {
		state.NextWord()
	}
	if buffer[0] == '~' {
		state.PreviousWord()
	}

	if string(buffer[0:2]) == "\x1b[" {
		if (buffer[2] == 68 || buffer[2] == 67) && state.SelectedClue != nil && state.SelectedClue.Direction == puz.DirectionDown {
			state.SwitchDirections()
		} else if (buffer[2] == 65 || buffer[2] == 66) && state.SelectedClue != nil && state.SelectedClue.Direction == puz.DirectionAcross {
			state.SwitchDirections()
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
				if state.Puzzle.Cells[state.SelectedCell][0] != nil {
					state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][0]
				} else {
					state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][1]
				}
			case puz.DirectionDown:
				if state.Puzzle.Cells[state.SelectedCell][1] != nil {
					state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][1]
				} else {
					state.SelectedClue = state.Puzzle.Cells[state.SelectedCell][0]
				}
			}
		}
	}
	return false
}
