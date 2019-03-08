export GOPATH = $(shell pwd)
export GOARCH=arm
export GOOS=linux

SRC=ui.go names.go const.go getchar.go pnokeys.go input.go input-subcmd.go display.go logscale.go
BUILD_DATE=$(shell (date +"%y%m%d.$$(git rev-list -1 HEAD)" | cut -b 1-9))

ui: $(SRC)
	go get -d -x
	go build -ldflags='-X main.buildDate=$(BUILD_DATE)' $^

install: ui Makefile
	-ssh pi@raspberrypi 'killall ui'
	scp ui pi@raspberrypi:/home/pi/

clean:
	go clean
	rm -rf src
	rm -rf ui
