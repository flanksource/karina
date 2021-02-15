
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
	which mkdocs || pip3 install mkdocs
	pip3 install -r requirements.txt
	mkdocs build -d build/docs

.PHONY: deploy
deploy: build
	which netlify || npm install -g netlify
	netlify deploy --site b7d97db0-1bc2-4e8c-903d-6ebf3da18358 --prod --dir build/docs


.PHONY: watch
watch:
	which mkdocs || pip3 install mkdocs
	pip3 install -r requirements.txt
	-kill $(cat .pid)
	(cd build/docs && python -m SimpleHTTPServer & echo "$$!" > .pid)
	-watchexec --ignore build/ -- mkdocs build -d build/docs
	-kill $(cat .pid)
	rm .pid
