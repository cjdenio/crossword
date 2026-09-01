package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cjdenio/crossword/puz"
)

func RenderPuzzle(puzzle string, width, height int) string {
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

		switch char {
		case '.':
			b.WriteRune(0x2588)
		case '-':
			b.WriteRune('_')
		default:
			b.WriteRune(char)
		}

		if (index+1)%width == 0 {
			b.WriteString(" │\n")
		} else if char == '.' && puzzle[index+1] == '.' {
			b.WriteRune(0x2588)
		} else {
			b.WriteRune(' ')
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

	fmt.Println(RenderPuzzle(puzzle.State, puzzle.Width, puzzle.Height))

	j, _ := json.Marshal(puzzle)
	fmt.Println(string(j))
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
