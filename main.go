package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/cjdenio/crossword/puz"
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

func RenderPuzzle(puzzle string, width, height int, selectedClueCells []int, selectedCell int) string {
	b := strings.Builder{}

	b.WriteRune('┌')
	for range (width * 2) + 1 {
		b.WriteRune('─')
	}
	b.WriteString("┐\n")

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
			b.WriteString(" │\n")
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
	b.WriteString("┘\n")

	return b.String()
}

func main() {
	// fmt.Print("\x1b[?1049h")
	// fmt.Print("\x1b[?25l")

	// defer func() {
	// 	fmt.Print("\x1b[?1049l")
	// 	fmt.Print("\x1b[?25h")
	// }()

	filename := os.Args[1]
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	puzzle, err := puz.ParsePuz(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("TITLE: %s\n", puzzle.Title)
	fmt.Printf("AUTHOR: %s\n", puzzle.Author)

	fmt.Println(RenderPuzzle(puzzle.State, puzzle.Width, puzzle.Height, puzzle.Clues[0].Cells, 1))

	// j, _ := json.Marshal(puzzle)
	// fmt.Println(string(j))
	// for _, clue := range puzzle.Clues {
	// 	if clue.Direction == puz.DirectionAcross {
	// 		fmt.Printf(". %d-across: %s: %s\n", clue.Number, clue.Clue, clue.Solution)
	// 	} else {
	// 		fmt.Printf(". %d-down: %s: %s\n", clue.Number, clue.Clue, clue.Solution)
	// 	}
	// }

	fmt.Printf("%s\n", puzzle.Copyright)
	// for {
	// 	n, _ := os.Stdin.Read(make([]byte, 1))
	// 	if n > 0 {
	// 		return
	// 	}
	// }
}
