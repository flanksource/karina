.PHONY: setup
	go get -u github.com/gobuffalo/packr/v2/packr2

.PHONY: build
build:
	packr2 build -o ./.bin/platform-cli -ldflags "-X \"main.version=$(shell date "+%Y-%m-%d %H:%M:%S")\""  main.go


.PHONY: install
install: build
	cp ./.bin/platform-cli /usr/local/bin/

.PHONY: docker
docker:
	docker build ./ -t platform-cli

