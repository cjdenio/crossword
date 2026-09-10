all: bin/crossword bin/crossword-ssh

bin/crossword: main.go puz/puz.go game/game.go
	go build -o $@ .
bin/crossword-ssh: ssh/main.go puz/puz.go game/game.go
	go build -o $@ ./ssh
