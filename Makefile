EXECUTABLE=gnmi-streamer
VERSION=$(shell git describe --tags --always --long)

.PHONY: all test clean server client

all: test build

upstream: check-env ## Make sure you TAG correctly. E.g. export TAG=0.1.0
	git add .
	git commit -m "Bump to version ${TAG}"
	git tag -a -m "Bump to version ${TAG}" v${TAG}
	git push --follow-tags

check-env: ## Check if TAG variable is set. Brought to you by https://stackoverflow.com/a/4731504
ifndef TAG
	$(error TAG is undefined)
endif
	@echo "TAG is ${TAG}"

tag:
	git tag <tagname>

test:
	go test ./... -v

fmt:
	go fmt ./...

run: fmt
	go build && ./$(EXECUTABLE)

server: fmt
	go build && ./$(EXECUTABLE)

client: fmt
	cd client
	go run main.go
	cd ..

build: fmt test
	@echo version: $(VERSION)