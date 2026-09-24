package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cjdenio/crossword/game"
	"github.com/cjdenio/crossword/puz"
	"golang.org/x/term"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Print("!!! PROGRAM CRASHED !!! Error:\r\n", r, "\r\n")
		}
	}()

	debugMode := flag.Bool("debug", false, "")
	flag.Parse()

	filename := flag.Arg(0)
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	puzzle, err := puz.ParsePuz(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\x1b[?25l")

	defer func() {
		fmt.Print("\x1b[?25h")
	}()
	termState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	defer term.Restore(int(os.Stdin.Fd()), termState)

	state := game.NewState(puzzle)
	if debugMode != nil {
		state.DebugMode = *debugMode
	}

	// try to load a saveFile
	saveFile, err := os.ReadFile(filename + ".save")
	if err == nil {
		err = state.LoadSaveFile(saveFile)
		if state.DebugMode && err != nil {
			state.LastKeySequence = fmt.Sprintf("failed to read savefile: %s", err)
		}
	}

	uiHeight := state.RenderUI(os.Stdout)

	scanner := bufio.NewScanner(os.Stdin)
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
			saveFile, err := state.CreateSaveFile()
			if err == nil {
				err = os.WriteFile(filename+".save", saveFile, fileInfo.Mode())
				if err != nil && state.DebugMode {
					state.LastKeySequence = fmt.Sprintf("failed to write savefile: %s", err)
				}
			}
			state.SelectedClue = nil
			state.SelectedCell = -1
			fmt.Printf("\r\x1b[%dA", uiHeight)
			fmt.Print("\x1b[J")
			state.RenderUI(os.Stdout)
			return
		}

		fmt.Printf("\r\x1b[%dA", uiHeight)
		fmt.Print("\x1b[J")

		uiHeight = state.RenderUI(os.Stdout)
	}
}
