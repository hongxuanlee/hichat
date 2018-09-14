PKG := github.com/hongxuanlee/hichat
VERSION := $(shell git describe --always --long --dirty)

deps:
	@echo $@
	@go get -u gopkg.in/abiosoft/ishell.v2 

build:
	@go build -o ${GOPATH}/bin/hichat ./
	@cp config.json ~/.hichat_conf.json

clean:
	@go clean
	@rm -f ~/.hichat_conf.json
 
run:
	@go run ./message main.go

.PHONY: deps build run
