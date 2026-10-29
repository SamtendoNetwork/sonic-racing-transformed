BINARY := build/sonic-nex

.PHONY: all build run test e2e fmt vet clean

all: build

build:
	go build -o $(BINARY) .

run:
	go run .

test:
	go test ./...

e2e:
	python tests/e2e_test.py

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf build
