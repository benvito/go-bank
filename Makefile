build:
	go build -o bin/gobank.exe

run: build
	./bin/gobank
