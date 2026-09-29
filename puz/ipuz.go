package puz

import (
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"
)

type iPuzDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
type iPuzClue struct {
	Number int    `json:"number"`
	Clue   string `json:"clue"`
}
type iPuzFile struct {
	Kind       []string              `json:"kind"`
	Title      string                `json:"title"`
	Author     string                `json:"author"`
	Copyright  string                `json:"copyright"`
	Puzzle     [][]any               `json:"puzzle"`
	Solution   [][]string            `json:"solution"`
	Dimensions iPuzDimensions        `json:"dimensions"`
	Clues      map[string][]iPuzClue `json:"clues"`
}

func ParseIPuz(file []byte) (*Puzzle, error) {
	var parsed iPuzFile
	err := json.Unmarshal(file, &parsed)
	if err != nil {
		return nil, err
	}

	if len(parsed.Kind) == 0 || !slices.ContainsFunc(parsed.Kind, func(k string) bool {
		return strings.HasPrefix(k, "http://ipuz.org/crossword")
	}) {
		return nil, errors.New("ipuz file is not a crossword")
	}

	puzzle := new(Puzzle)

	puzzle.Title = parsed.Title
	puzzle.Author = parsed.Author
	puzzle.Copyright = parsed.Copyright

	puzzle.Width = parsed.Dimensions.Width
	puzzle.Height = parsed.Dimensions.Height

	puzzle.ClueCount = len(parsed.Clues["Across"]) + len(parsed.Clues["Down"])

	clueToCellMap := make(map[int][2]int)

	state := strings.Builder{}
	for y, row := range parsed.Puzzle {
		for x, cell := range row {
			switch c := cell.(type) {
			case string:
				if c == "#" {
					state.WriteRune('.')
				} else {
					if i, err := strconv.Atoi(c); err == nil {
						clueToCellMap[i] = [2]int{x, y}
					}
					state.WriteRune('-')
				}
			case float64:
				clueToCellMap[int(c)] = [2]int{x, y}
				state.WriteRune('-')
			default:
				state.WriteRune('-')
			}
		}
	}
	puzzle.State = state.String()

	solution := strings.Builder{}
	for _, row := range parsed.Solution {
		for _, cell := range row {
			if cell == "#" {
				solution.WriteRune('.')
			} else if len(cell) > 1 {
				solution.WriteByte(cell[0])
				puzzle.HasRebus = true
			} else if len(cell) == 1 {
				solution.WriteString(cell)
			} else {
				return nil, errors.New("invalid ipuz file")
			}
		}
	}
	puzzle.Solution = solution.String()

	puzzle.Clues = make([]*Clue, 0, puzzle.ClueCount)
	puzzle.Cells = make([][2]*Clue, puzzle.Width*puzzle.Height)

	for _, clue := range parsed.Clues["Across"] {
		cellCoords := clueToCellMap[clue.Number]

		clueCells := make([]int, 0)
		clueSolution := strings.Builder{}

		clueStruct := &Clue{
			Clue:      clue.Clue,
			Direction: DirectionAcross,
			Number:    clue.Number,
		}

		for i := cellCoords[0]; i < puzzle.Width; i++ {
			cell := parsed.Solution[cellCoords[1]][i]
			if cell == "#" {
				break
			}

			cellIdx := cellCoords[1]*puzzle.Width + i

			clueCells = append(clueCells, cellIdx)
			puzzle.Cells[cellIdx][0] = clueStruct
			clueSolution.WriteByte(cell[0])
		}

		clueStruct.Solution = clueSolution.String()
		clueStruct.Cells = clueCells

		puzzle.Clues = append(puzzle.Clues, clueStruct)
	}

	for _, clue := range parsed.Clues["Down"] {
		cellCoords := clueToCellMap[clue.Number]

		clueCells := make([]int, 0)
		clueSolution := strings.Builder{}

		clueStruct := &Clue{
			Clue:      clue.Clue,
			Direction: DirectionDown,
			Number:    clue.Number,
		}

		for i := cellCoords[1]; i < puzzle.Height; i++ {
			cell := parsed.Solution[i][cellCoords[0]]
			if cell == "#" {
				break
			}

			cellIdx := i*puzzle.Width + cellCoords[0]

			clueCells = append(clueCells, cellIdx)
			puzzle.Cells[cellIdx][1] = clueStruct
			clueSolution.WriteByte(cell[0])
		}

		clueStruct.Solution = clueSolution.String()
		clueStruct.Cells = clueCells

		puzzle.Clues = append(puzzle.Clues, clueStruct)
	}

	return puzzle, nil
}
