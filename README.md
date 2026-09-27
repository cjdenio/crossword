Very silly terminal-based crossword solver. It can play .puz / .ipuz files from a variety of sources, including [Crosshare](https://crosshare.org).

- Requirements: [Go](https://go.dev)
- Build: `make`
- Usage: `./bin/crossword <file>.puz`

![demo](demo.gif)

---

todo

- correctly calculate uiHeight with multi-line clues
- timer
- reveal grid/word
- .puz checksums
- fix shift+enter / shift+tab in other terminals
- clear grid
- rebus
- detect rebus for .puz files
