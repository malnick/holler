.PHONY: all build test clean

all: clean build test

build:
	bash -c "./scripts/build.sh proxy"

test:
	bash -c "./scripts/test.sh proxy unit"

clean:
	rm -rf ./build
