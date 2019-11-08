
default: build
NAME:=$(shell basename $(PWD))

VERSION:=v$(shell git tag --points-at HEAD ) $(shell date "+%Y-%m-%d %H:%M:%S")
.PHONY: setup
setup:
	which packr2 2>&1 > /dev/null || go get -u github.com/gobuffalo/packr/v2/packr2

.PHONY: build
build: setup
	go build -o ./.bin/$(NAME) -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: pack
pack:
	packr2 build -o ./.bin/$(NAME) -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: linux
linux: setup
	GOOS=linux packr2 build -o ./.bin/$(NAME) -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: darwin
darwin: setup
	GOOS=darwin packr2 build -o ./.bin/$(NAME)_osx -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: install
install: build
	cp ./.bin/$(NAME) /usr/local/bin/

.PHONY: docker
docker:
	docker build ./ -t $(NAME)

