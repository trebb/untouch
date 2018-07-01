export GOPATH = $(shell pwd)

SRC=ui.go names.go const.go getchar.go

all: ui

ui: $(SRC)
	go get -d -x
	go build $^

clean:
	go clean
	rm -rf src

