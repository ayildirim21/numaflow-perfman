.PHONY: build
build:
	CGO_ENABLED=0 GOARCH=amd64 go build -v -o perfman main.go