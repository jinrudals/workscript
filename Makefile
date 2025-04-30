.PHONY: all clean

all:
		go build

clean:
		git clean -d -x -f
