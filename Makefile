all: bin/crossword bin/crossword-ssh

bin/crossword: cmd/main.go puz/puz.go game/game.go
	go build -o $@ ./cmd
bin/crossword-ssh: ssh/main.go puz/puz.go game/game.go
	go build -o $@ ./ssh
