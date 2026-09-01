package puz

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

type Direction int

const (
	DirectionAcross Direction = iota
	DirectionDown
)

type Clue struct {
	Clue      string
	Solution  string
	Cells     []int
	Direction Direction
	Number    int
}

type Puzzle struct {
	ClueCount int
	Width     int
	Height    int

	Title     string
	Author    string
	Copyright string

	Solution string
	State    string

	Clues []Clue
}

func startOfDownClue(puzzle string, width, index int) bool {
	if puzzle[index] == '.' {
		return false
	}
	if index < width {
		return true
	}
	if puzzle[index-width] == '.' {
		return true
	}
	return false
}

func startOfAcrossClue(puzzle string, width, index int) bool {
	if puzzle[index] == '.' {
		return false
	}
	if index%width == 0 {
		return true
	}
	if puzzle[index-1] == '.' {
		return true
	}
	return false
}

func readText(b *bufio.Reader) (string, error) {
	text, err := b.ReadString(0x00)
	if err != nil {
		return "", err
	}

	decoded, err := charmap.ISO8859_1.NewDecoder().String(text[:len(text)-1])
	if err != nil {
		return "", err
	}

	return decoded, nil
}

func ParsePuz(file []byte) (*Puzzle, error) {
	// read magic bytes
	magic := string(file[2:14])
	if magic != "ACROSS&DOWN\x00" {
		return nil, errors.New("not a valid .puz file")
	}

	puzzle := new(Puzzle)

	// read header
	puzzle.Width = int(file[0x2C])
	puzzle.Height = int(file[0x2D])
	puzzle.ClueCount = int(binary.LittleEndian.Uint16(file[0x2E:0x30]))

	puzzleSize := puzzle.Width * puzzle.Height
	solutionStart := 0x34
	solutionEnd := 0x34 + puzzleSize

	stateStart := solutionEnd
	stateEnd := stateStart + puzzleSize

	puzzle.Solution = string(file[solutionStart:solutionEnd])
	puzzle.State = string(file[stateStart:stateEnd])

	// read text
	buf := bufio.NewReader(bytes.NewReader(file[stateEnd:]))

	title, err := readText(buf)
	if err != nil {
		return nil, err
	}
	puzzle.Title = title

	author, err := readText(buf)
	if err != nil {
		return nil, err
	}
	puzzle.Author = author

	copyright, err := readText(buf)
	if err != nil {
		return nil, err
	}
	puzzle.Copyright = copyright

	// read clues
	clues := make([]string, 0, puzzle.ClueCount)
	for range puzzle.ClueCount {
		clue, err := readText(buf)
		if err != nil {
			return nil, err
		}
		clues = append(clues, clue)
	}

	// assign clue numbers
	puzzle.Clues = make([]Clue, 0, puzzle.ClueCount)

	clueNumber := 1
	clueIndex := 0

	for i, _ := range puzzle.Solution {
		cellGetsNumber := false
		if startOfAcrossClue(puzzle.Solution, puzzle.Width, i) {
			cellGetsNumber = true
			puzzle.Clues = append(puzzle.Clues, Clue{
				Cells:     []int{i},
				Direction: DirectionAcross,
				Number:    clueNumber,
				Clue:      clues[clueIndex],
			})
			clueIndex++
		}

		if startOfDownClue(puzzle.Solution, puzzle.Width, i) {
			cellGetsNumber = true
			puzzle.Clues = append(puzzle.Clues, Clue{
				Cells:     []int{i},
				Direction: DirectionDown,
				Number:    clueNumber,
				Clue:      clues[clueIndex],
			})
			clueIndex++
		}

		if cellGetsNumber {
			clueNumber++
		}
	}

	// get solutions
	for i, clue := range puzzle.Clues {
		var solution strings.Builder
		currentCell := clue.Cells[0]

		for {
			solution.WriteByte(puzzle.Solution[currentCell])

			if currentCell != clue.Cells[0] {
				clue.Cells = append(clue.Cells, currentCell)
			}

			if clue.Direction == DirectionAcross {
				if ((currentCell+1)%puzzle.Width == 0) || puzzle.Solution[currentCell+1] == '.' {
					break
				} else {
					currentCell++
				}
			} else if clue.Direction == DirectionDown {
				if int(math.Floor(float64(currentCell/puzzle.Width)))+1 == puzzle.Height || puzzle.Solution[currentCell+puzzle.Width] == '.' {
					break
				} else {
					currentCell += puzzle.Width
				}
			}
		}

		puzzle.Clues[i].Solution = solution.String()
		puzzle.Clues[i].Cells = clue.Cells
	}

	return puzzle, nil
}
