.PHONY: build install test clean

BINARY_NAME=gitusr

build:
	go build -o bin/$(BINARY_NAME) ./cmd/gitusr

test:
	go test ./... -v

install: build
	cp bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	ln -sf /usr/local/bin/$(BINARY_NAME) /usr/local/bin/gu

clean:
	rm -rf bin/
