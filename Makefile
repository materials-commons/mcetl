.PHONY: bin test all fmt deploy docs server cli setup

all: fmt bin

fmt:
	-go fmt ./...

bin: cli

cli:
	(cd ./cmd/mcetl; go build)

devrun:
	-reflex -r '\.go$\' -s make run

devtest:
	-reflex -r '\.go$\' -s make test 

dep:
	dep ensure

testdb:
	-(cd ./internal/store/migration; go test -count=1)

test: 
	-(cd ./internal/store/migration; go test)
	-go test -v ./...

docs:
	./makedocs.sh

setup:
	-go get -u github.com/cespare/reflex
	-dep ensure
