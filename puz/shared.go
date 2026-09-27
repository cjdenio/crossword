package puz

import (
	"bufio"
	"errors"
	"io"
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

	Clues []*Clue

	Cells [][2]*Clue

	HasRebus bool
}

func LoadPuzzle(r io.Reader) (*Puzzle, error) {
	bufReader := bufio.NewReader(r)
	header, err := bufReader.Peek(16)
	if err != nil {
		return nil, err
	}

	if string(header[2:14]) == "ACROSS&DOWN\x00" {
		file, err := io.ReadAll(bufReader)
		if err != nil {
			return nil, err
		}
		return ParsePuz(file)
	} else if header[0] == '{' {
		file, err := io.ReadAll(bufReader)
		if err != nil {
			return nil, err
		}
		return ParseIPuz(file)
	} else {
		return nil, errors.New("could not detect file type")
	}
}
