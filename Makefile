NAME = "fastcar"

BUILD_DATE=$(shell date '+%Y-%m-%d %H:%M:%S')

CFLAGS = -ldflags "-s -w -X \"main.BuildDate=$(BUILD_DATE)\""

build:
	go build $(CFLAGS) -o $(NAME) ./