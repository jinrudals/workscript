.PHONY: all build clean-all clean-output

all: build

build:
		go build

clean: clean-all

clean-all:
		git clean -d -x -f

clean-output:
		rm workscript
