.PHONY: setup
setup:
	go get -u github.com/gobuffalo/packr/v2/packr2

.PHONY: build
build:
	go build -o ./.bin/platform-cli -ldflags "-X \"main.version=$(shell date "+%Y-%m-%d %H:%M:%S")\""  main.go

.PHONY: pack
pack:
	packr2 build -o ./.bin/platform-cli -ldflags "-X \"main.version=$(shell date "+%Y-%m-%d %H:%M:%S")\""  main.go


.PHONY: linux
linux:
	GOOS=linux packr2 build -o ./.bin/platform-cli -ldflags "-X \"main.version=$(shell date "+%Y-%m-%d %H:%M:%S")\""  main.go

.PHONY: install
install: build
	cp ./.bin/platform-cli /usr/local/bin/

.PHONY: docker
docker:
	docker build ./ -t platform-cli

