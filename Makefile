# SUBDIRS := data mesh messanger otto server station utils
PIENV	= env GOOS=linux GOARCH=arm GOARM=7

all: test $(SUBDIRS)

init:
	git update --init 

test:
	go test -cover ./...

verbose:
	go test -v -cover ./...

# go test -coverprofile=./cover.out ./...
# go tool cover -func=cover.out
# go tool cover -html=cover.out

.PHONY: all test build $(SUBDIRS)
