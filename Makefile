# Go parameters
OS_MACH=$(shell uname -o; uname -m)
BINARY_NAME=quicklog
ifeq ($(OS_MACH),GNU/Linux x86_64)
BINARY_LINUX=$(BINARY_NAME)
else
BINARY_LINUX=$(BINARY_NAME)-linux
endif
BINARY_DARWIN=$(BINARY_NAME)-darwin

all: test build

rebuild: delete-binary build

rebuild-linux: delete-binary-linux build-linux

delete-binary:
	-rm $(BINARY_NAME) 2>/dev/null

delete-binary-linux:
	-rm $(BINARY_LINUX) 2>/dev/null

build: deps
	go build -o $(BINARY_NAME) -v

build-linux:
	CGO_ENABLED= GOOS=linux GOARCH=amd64 go build -o $(BINARY_LINUX) -v

build-darwin:
	CGO_ENABLED= GOOS=darwin GOARCH=amd64 go build -o $(BINARY_DARWIN) -v

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
