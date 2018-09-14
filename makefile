PKG := github.com/hongxuanlee/hichat
VERSION := $(shell git describe --always --long --dirty)

deps:
	@echo $@
	@go get -u gopkg.in/abiosoft/ishell.v2 

build:
	@GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/hichat ./
	@GOOS=darwin GOARCH=amd64 go build -o bin/darwin-amd64/hichat ./
	@cp config.json ~/.hichat_conf.json

clean:
	@go clean
	@rm -f ~/.hichat_conf.json
 
run:
	@go run ./message main.go

.PHONY: deps build run
