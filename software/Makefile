export GOPATH = $(shell pwd)
export GOARCH=arm
export GOOS=linux

SRC_BTN=ui.go names.go const.go getchar.go pnokeys.go input-btn.go input-subcmd.go display.go logscale.go
SRC_CLI=ui.go names.go const.go getchar.go pnokeys.go input-cli.go input-subcmd.go display.go logscale.go

all: ui-cli ui-btn

ui-btn: $(SRC_BTN)
	go get -d -x
	go build $^
	mv ui ui-btn
	# -ssh pi@raspberrypi 'killall ui-btn'
	# scp $@ pi@raspberrypi:/home/pi/

ui-cli: $(SRC_CLI)
	go get -d -x
	go build $^
	mv ui ui-cli
	# scp $@ pi@raspberrypi:/home/pi/

clean:
	go clean
	rm -rf src
	rm -rf ui-btn ui-cli
