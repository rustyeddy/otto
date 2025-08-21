SUBDIRS := examples
PIENV	= env GOOS=linux GOARCH=arm GOARM=7

all: test $(SUBDIRS)

init:
	git update --init 

test:
	go test -cover ./...
	go test -coverprofile=./cover.out ./...
	go tool cover -func=cover.out
	go tool cover -html=cover.out

$(SUBDIRS):
	$(MAKE) -C $@

.PHONY: all test build $(SUBDIRS)
