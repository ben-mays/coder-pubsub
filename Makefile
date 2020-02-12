BIN = $(CURDIR)/bin
PKGS := $(shell go list ./... | grep -v /vendor)
TESTPKGS = $(shell go list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))

build:
	go build -o $(BIN)/pubsubhttp

run: build
	./bin/pubsubhttp

test:
	go test $(TESTPKGS)
	go test -race $(TESTPKGS)