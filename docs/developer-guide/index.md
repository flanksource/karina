# Developer Guide

## Quickstart

```bash
make setup
make
sudo make install
```

## Documentation

Documentation is done using [MkDocs](https://github.com/mkdocs/mkdocs). 

Documentation files are MarkDown formatted text files.

### View and Edit Documentation Locally

Run

```sh
docker run --rm -it -p 8000:8000 -v $PWD:/docs -w /docs \
    squidfunk/mkdocs-material serve -a 0.0.0.0:8000
```

Navigate to [http://localhost:8000](http://localhost:8000)

The documentation sources are located in the repository in `doc` the mkdocs development server will live-reload the pages as soon as changed.