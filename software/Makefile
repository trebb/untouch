export GOPATH = $(shell pwd)
export GOARCH=arm
export GOOS=linux

SRC=ui.go names.go const.go getchar.go pnokeys.go input-btn.go input-subcmd.go display.go logscale.go

ui: $(SRC)
	go get -d -x
	go build $^

install: ui Makefile
	-ssh pi@raspberrypi 'killall ui'
	scp ui pi@raspberrypi:/home/pi/

clean:
	go clean
	rm -rf src
	rm -rf ui
