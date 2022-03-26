.DEFAULT: all

all: build-win build-linux

build-win:
	GOOS=windows GOARCH=amd64 go build ./cmd/pandacapture

build-linux:
	GOOS=linux GOARCH=amd64 go build ./cmd/pandacapture

clean:
	rm -f pandacapture pandacapture.exe

.PHONY: all build-win build-linux
