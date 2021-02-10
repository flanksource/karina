
default: serve
NAME:=karina
CONTAINER:=mkdocs


.PHONY: container
container:
	docker build -t $(CONTAINER) ./ --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)


.PHONY: serve
serve: container
	docker run --rm -it -p 8000:8000 -v $(PWD):/docs -w /docs $(CONTAINER)

.PHONY: build
build:
	which mkdocs || pip3 install -r requirements.txt
	mkdocs build -d build/docs


.PHONY: deploy
deploy: build
	netlify deploy --site b7d97db0-1bc2-4e8c-903d-6ebf3da18358 --prod --dir build/docs


