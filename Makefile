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
	ssh prod "mv /opt/quicklog/$(BINARY_NAME) /opt/quicklog/$(BINARY_NAME)-old || true"
	scp $(BINARY_LINUX) prod:/opt/quicklog/$(BINARY_NAME)
	ssh prod "sudo service quicklog restart"

run: deps
	go run -race main.go

deps:
	go get github.com/lib/pq
	go get github.com/kuangchanglang/graceful

build-linux:
	CGO_ENABLED= GOOS=linux GOARCH=amd64 go build -o $(BINARY_LINUX) -v
