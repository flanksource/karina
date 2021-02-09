
default: serve
NAME:=karina
CONTAINER:=mkdocs


.PHONY: container
container:
	docker build -t $(CONTAINER) ./


.PHONY: serve
serve: container
	docker run --rm -it -p 8000:8000 -v $(PWD):/docs -w /docs $(CONTAINER)

.PHONY: build
build: container
	docker run --rm  -v $(PWD):/docs -w /docs $(CONTAINER) build -d build/docs


.PHONY: deploy
deploy: build
	which netlify 2>&1 > /dev/null || sudo
	netlify deploy --site b7d97db0-1bc2-4e8c-903d-6ebf3da18358 --prod --dir build/docs


