build:
	python ./tool/make_ast.py
	go generate ./...
	go build ./...
