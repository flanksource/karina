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
make serve-docs
```

Navigate to [http://localhost:8000](http://localhost:8000)

The documentation sources are located in the repository in `doc` the mkdocs development server will live-reload the pages as soon as changed.