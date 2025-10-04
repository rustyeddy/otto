# SUBDIRS := data mesh messanger otto server station utils
PIENV	= env GOOS=linux GOARCH=arm GOARM=7

all: test $(SUBDIRS)

init:
	git update --init 

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

# go test -coverprofile=./cover.out ./...
# go tool cover -func=cover.out
# go tool cover -html=cover.out

.PHONY: all test build $(SUBDIRS)
