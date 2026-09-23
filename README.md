Very silly terminal-based crossword solver. It can play .puz files from a variety of sources, including [Crosshare](https://crosshare.org).

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
- rebus (at least add message that rebus is not supported)
- .ipuz files
