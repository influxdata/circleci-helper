.PHONY: all
all: mod-download circleci-helper

mod-download:
	go mod download

circleci-helper:
	CGO_ENABLED=0 go build -ldflags=-buildid= -o ./circleci-helper ./cmd/circleci-helper
	strip ./circleci-helper

.PHONY: clean
clean:
	rm -f ./circleci-helper

