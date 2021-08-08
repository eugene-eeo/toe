COMMIT = $(shell git rev-list -1 HEAD)

build:
	python ./tool/make_ast.py
	go generate ./...
	go build ./...
	go build -ldflags "-X main.VERSION=$(COMMIT)" .

test:
	gotest ./...
