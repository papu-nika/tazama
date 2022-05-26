GO_FILES = $(shell find . -type f -name '*.go' -print)

BINARIE_NAME = tazama

all : set build

set:
	go mod tidy

build: 
	go build -o ./bin/ $(GO_FILES)

clean:
	rm bin/$(BINARIE_NAME)
