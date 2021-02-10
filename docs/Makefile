
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


.PHONY: scratchpad
scratchpad:
	mkdir -p .scratchpad
	# save existing value
	cp  .scratchpad/index.md $(SCRATCH) || true
	cp -R mkdocs.yml overrides .scratchpad/

	yq d -i .scratchpad/mkdocs.yml nav
	yq w -i .scratchpad/mkdocs.yml nav[0].Index. ./index.md

	cp $(SCRATCH) .scratchpad/index.md
	cp -R overrides .scratchpad/
	open  .scratchpad/index.md
	open http://localhost:8000
	docker run --rm  -p 8000:8000 -v $(PWD)/.scratchpad:/docs -w /docs $(CONTAINER)

