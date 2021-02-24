# Developer Guide

## Quickstart

#### Clone the repo

```bash
git clone git@github.com:flanksource/karina.git
cd karina
```

#### Install Go

Karina requires go v1.16 or higher.

#### Compile Karina

Run the following to get going:

```bash
make              # do a local build
make compress     # compress the built executable
sudo make install # install the executable to /usr/local/bin/
```

## Documentation

Documentation is done using [MkDocs](https://github.com/mkdocs/mkdocs).

Documentation files are MarkDown formatted text files. These are in the `docs/` hierarchy.

Run the following to view and edit the documentation locally:

```sh
make serve-docs
```

Navigate to [http://localhost:8000](http://localhost:8000)

Update the documentation sources located in the repository in `docs/` (and its subdirectories) and the mkdocs development server will live-reload the pages as soon as changed.
