.PHONY: build
build:
	go build -o ./.bin/platform-cli main.go


.PHONY: install
install: build
	cp ./.bin/platform-cli /usr/local/bin/

.PHONY: docker
docker:
	docker build ./ -t platform-cli

