all: bin/crossword bin/ssh

bin/crossword: main.go puz/puz.go game/game.go
	go build -o $@ .
bin/ssh: ssh/main.go puz/puz.go game/game.go
	go build -o $@ ./ssh
