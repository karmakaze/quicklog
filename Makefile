# Go parameters
BINARY_NAME=quicklog
BINARY_LINUX=$(BINARY_NAME)_linux

all: test build

build:
	go build -o $(BINARY_NAME) -v

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_LINUX)

deploy: build-linux
	echo ssh quicklog "mv /opt/quicklog/$(BINARY_NAME) /opt/quicklog/$(BINARY_NAME)-old"
	echo scp $(BINARY_LINUX) quicklog:/opt/quicklog/$(BINARY_NAME)
	echo ssh quicklog "sudo service quicklog restart"

run: deps
	go run -race main.go

deps:
	go get github.com/lib/pq

build-linux:
	CGO_ENABLED= GOOS=linux GOARCH=amd64 go build -o $(BINARY_LINUX) -v
