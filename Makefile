# SUBDIRS := data mesh messanger otto server station utils
PIENV	= env GOOS=linux GOARCH=arm GOARM=7
BINARY_NAME=otto
VERSION?=0.1.0

all: test build

init:
	git update --init 

fmt:
	gofmt -s -w .

vet:
	go vet ./...

build:
	go build -o ${BINARY_NAME} -ldflags "-X github.com/rustyeddy/otto/cmd.version=${VERSION}" ./cmd/otto

test:
	rm -f cover.out
	go test -coverprofile=cover.out -cover ./...

verbose:
	rm -f cover.out
	go test -v -coverprofile=cover.out -cover ./...

coverage: test
	go tool cover -func=cover.out

html: test
	rm -f coverage.html
	go tool cover -html=cover.out -o coverage.html

clean:
	rm -f ${BINARY_NAME}
	rm -f cover.out coverage.html

ci: fmt vet test build

.PHONY: all test build fmt vet clean ci $(SUBDIRS)
