COMMIT_SHA=$(shell git rev-parse HEAD)

.PHONY: build
build:
	CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-X github.com/ayildirim21/numaflow-perfman/logging.CommitSHA=$(COMMIT_SHA)" -v -o perfman main.go

.PHONY: clean
clean:
	-rm perfman