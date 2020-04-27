
default: build
NAME:=platform-cli

ifeq ($(VERSION),)
VERSION := v$(shell git describe --tags --exclude "*-g*" ) built $(shell date)
endif

.PHONY: help
help:
	@cat docs/developer-guide/make-targets.md

.PHONY: setup
setup:
	which esc 2>&1 > /dev/null || go get -u github.com/mjibson/esc
	which github-release 2>&1 > /dev/null || go get github.com/aktau/github-release


.PHONY: build
build:
	go build -o ./.bin/$(NAME) -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: pack
pack: setup
	esc --prefix "manifests/" --ignore "static.go" -o manifests/static.go --pkg manifests manifests
	esc --prefix "templates/" --ignore "static.go" -o templates/static.go --pkg templates templates

.PHONY: linux
linux:
	GOOS=linux go build -o ./.bin/$(NAME) -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: darwin
darwin:
	GOOS=darwin go build -o ./.bin/$(NAME)_osx -ldflags "-X \"main.version=$(VERSION)\""  main.go

.PHONY: compress
compress:
	# upx 3.95 has issues compressing darwin binaries - https://github.com/upx/upx/issues/301
	which upx 2>&1 >  /dev/null  || (sudo apt-get update && sudo apt-get install -y xz-utils && wget -nv -O upx.tar.xz https://github.com/upx/upx/releases/download/v3.96/upx-3.96-amd64_linux.tar.xz; tar xf upx.tar.xz; mv upx-3.96-amd64_linux/upx /usr/bin )
	upx -5 ./.bin/$(NAME) ./.bin/$(NAME)_osx

.PHONY: install
install:
	cp ./.bin/$(NAME) /usr/local/bin/

.PHONY: docker
docker:
	docker build ./ -t $(NAME)

.PHONY: serve-docs
serve-docs:
	docker run --rm -it -p 8000:8000 -v $(PWD):/docs -w /docs squidfunk/mkdocs-material

.PHONY: build-api-docs
build-api-docs:
	go run main.go docs api  pkg/types/config.go pkg/types/types.go pkg/types/nsx.go  > docs/reference/config.md
	go run main.go docs cli "docs/cli"

.PHONY: build-docs
build-docs:
	which mkdocs 2>&1 > /dev/null || pip install mkdocs mkdocs-material
	mkdocs build -d build/docs

.PHONY: deploy-docs
deploy-docs:
	which netlify 2>&1 > /dev/null || sudo npm install -g netlify-cli
	netlify deploy --site b7d97db0-1bc2-4e8c-903d-6ebf3da18358 --prod --dir build/docs

.PHONY: lint
lint: pack build
	golangci-lint run --verbose --print-resources-usage
