package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cjdenio/crossword/game"
	"github.com/cjdenio/crossword/puz"
	"github.com/gliderlabs/ssh"
)

func main() {
	ssh.Handle(func(s ssh.Session) {
		file, err := os.ReadFile(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		puzzle, err := puz.ParsePuz(file)
		if err != nil {
			log.Fatal(err)
		}

		state := game.State{
			Puzzle:       puzzle,
			PuzzleState:  []rune(puzzle.State),
			SelectedClue: puzzle.Clues[0],
			SelectedCell: puzzle.Clues[0].Cells[0],
		}

		io.WriteString(s, "\x1b[?25l")

		defer func() {
			io.WriteString(s, "\x1b[?25h")
		}()

		uiHeight := state.RenderUI(s)

		scanner := bufio.NewScanner(s)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if len(data) == 0 {
				return 0, nil, bufio.ErrFinalToken
			}

			if data[0] == 0x1b {
				if len(data) < 3 {
					return 0, nil, nil
				}

				return 3, data[0:3], nil
			}

			return 1, data[0:1], nil
		})

		for {
			scanner.Scan()
			buffer := scanner.Bytes()

			exited := state.HandleInput(buffer)
			if exited {
				state.SelectedClue = nil
				state.SelectedCell = -1
				state.Goodbye = true
				fmt.Fprintf(s, "\r\x1b[%dA", uiHeight)
				fmt.Fprint(s, "\x1b[J")
				state.RenderUI(s)
				return
			}

			fmt.Fprintf(s, "\r\x1b[%dA", uiHeight)
			fmt.Fprint(s, "\x1b[J")

			uiHeight = state.RenderUI(s)
		}
	})
	log.Fatal(ssh.ListenAndServe(":1234", nil))
}
