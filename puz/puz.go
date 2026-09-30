package puz

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

func startOfDownClue(puzzle string, width, height, index int) bool {
	if puzzle[index] == '.' {
		return false
	}
	if int(math.Floor(float64(index/width)))+1 == height || puzzle[index+width] == '.' {
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
	if (index+1)%width == 0 || puzzle[index+1] == '.' {
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

func readText(b *bufio.Reader, version string) (string, error) {
	text, err := b.ReadString(0x00)
	if err != nil {
		return "", err
	}

	if version == "1" {
		decoded, err := charmap.ISO8859_1.NewDecoder().String(text[:len(text)-1])
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(decoded), nil
	} else {
		return strings.TrimSpace(string(text[:len(text)-1])), nil
	}
}

type ExtraBlock struct {
	Name string
	Data []byte
}

func readExtraBlock(r *bufio.Reader) (*ExtraBlock, error) {
	name := make([]byte, 4)
	_, err := r.Read(name)
	if err != nil {
		return nil, err
	}

	length := make([]byte, 2)
	_, err = r.Read(length)
	if err != nil {
		return nil, err
	}

	_, err = r.Discard(2)
	if err != nil {
		return nil, err
	}

	data := make([]byte, binary.LittleEndian.Uint16(length))
	_, err = r.Read(data)
	if err != nil {
		return nil, err
	}

	// consume trailing null byte
	_, err = r.Discard(1)
	if err != nil {
		return nil, err
	}

	return &ExtraBlock{
		Name: string(name),
		Data: data,
	}, nil
}

func ParsePuz(file []byte) (*Puzzle, error) {
	r := bytes.NewReader(file)

	_, err := r.Seek(2, io.SeekStart) // consume the checksum
	if err != nil {
		return nil, err
	}

	// read magic bytes
	magic := make([]byte, 12)
	_, err = r.Read(magic)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(magic, []byte("ACROSS&DOWN\x00")) {
		return nil, errors.New("not a valid .puz file")
	}

	puzzle := new(Puzzle)

	puzzle.RebusCells = make(map[int]string)

	_, err = r.Seek(10, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	// read header
	version := make([]byte, 4)
	_, err = r.Read(version)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(16, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	width, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	height, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	puzzle.Width = int(width)
	puzzle.Height = int(height)

	clueCount := make([]byte, 2)
	_, err = r.Read(clueCount)
	if err != nil {
		return nil, err
	}
	puzzle.ClueCount = int(binary.LittleEndian.Uint16(clueCount))

	puzzleSize := puzzle.Width * puzzle.Height

	_, err = r.Seek(4, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	solution := make([]byte, puzzleSize)
	_, err = r.Read(solution)
	if err != nil {
		return nil, err
	}
	state := make([]byte, puzzleSize)
	_, err = r.Read(state)
	if err != nil {
		return nil, err
	}

	puzzle.Solution = string(solution)
	puzzle.State = string(state)

	// read text fields
	buf := bufio.NewReader(r)

	title, err := readText(buf, string(version[0]))
	if err != nil {
		return nil, err
	}
	puzzle.Title = title

	author, err := readText(buf, string(version[0]))
	if err != nil {
		return nil, err
	}
	puzzle.Author = author

	copyright, err := readText(buf, string(version[0]))
	if err != nil {
		return nil, err
	}
	puzzle.Copyright = copyright

	// read clues
	clues := make([]string, 0, puzzle.ClueCount)
	for range puzzle.ClueCount {
		clue, err := readText(buf, string(version[0]))
		if err != nil {
			return nil, err
		}
		clues = append(clues, clue)
	}

	// read the "note" field, currently unused
	_, err = readText(buf, string(version[0]))
	if err != nil {
		return nil, err
	}

	extraBlocks := make(map[string]*ExtraBlock)

	for {
		extra, err := readExtraBlock(buf)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		extraBlocks[extra.Name] = extra
	}

	grbs, grbsOk := extraBlocks["GRBS"]
	trbl, trblOk := extraBlocks["RTBL"]
	if grbsOk && trblOk {
		puzzle.HasRebus = true
		rebusCellMap := make(map[int][]int)
		for i, b := range grbs.Data {
			if b > 0 {
				rebusCellMap[int(b)-1] = append(rebusCellMap[int(b)-1], i)
			}
		}

		rebusData := regexp.MustCompile(`\s*(\d+):(\w+);\s*`).FindAllStringSubmatch(string(trbl.Data), -1)
		for _, rebus := range rebusData {
			cell, err := strconv.Atoi(rebus[1])
			if err != nil {
				continue
			}

			for _, c := range rebusCellMap[cell] {
				puzzle.RebusCells[c] = rebus[2]
			}
		}
	}

	// assign clue numbers
	puzzle.Clues = make([]*Clue, 0, puzzle.ClueCount)
	puzzle.Cells = make([][2]*Clue, len(puzzle.Solution))

	clueNumber := 1
	clueIndex := 0

	for i, _ := range puzzle.Solution {
		cellGetsNumber := false
		if startOfAcrossClue(puzzle.Solution, puzzle.Width, i) {
			cellGetsNumber = true
			clue := Clue{
				Cells:     []int{i},
				Direction: DirectionAcross,
				Number:    clueNumber,
				Clue:      clues[clueIndex],
			}
			puzzle.Clues = append(puzzle.Clues, &clue)
			puzzle.Cells[i][0] = &clue
			clueIndex++
		}

		if startOfDownClue(puzzle.Solution, puzzle.Width, puzzle.Height, i) {
			cellGetsNumber = true
			clue := Clue{
				Cells:     []int{i},
				Direction: DirectionDown,
				Number:    clueNumber,
				Clue:      clues[clueIndex],
			}
			puzzle.Clues = append(puzzle.Clues, &clue)
			puzzle.Cells[i][1] = &clue
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
				switch clue.Direction {
				case DirectionAcross:
					puzzle.Cells[currentCell][0] = clue
				case DirectionDown:
					puzzle.Cells[currentCell][1] = clue
				}
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
