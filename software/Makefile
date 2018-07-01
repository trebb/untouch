export GOPATH = $(shell pwd)

SRC_BTN=ui.go names.go const.go getchar.go input-btn.go
SRC_CLI=ui.go names.go const.go getchar.go input-cli.go

all: ui-btn ui-cli

ui-btn: $(SRC_BTN)
	go get -d -x
	go build $^
	mv ui ui-btn

ui-cli: $(SRC_CLI)
	go get -d -x
	go build $^
	mv ui ui-cli
clean:
	go clean
	rm -rf src
	rm -rf ui-btn ui-cli
