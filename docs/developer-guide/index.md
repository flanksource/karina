# Developer Guide

## Quickstart

#### Clone the repo

```bash
git clone git@github.com:flanksource/karina.git
cd karina
```

Run the following to get going:

```bash
make setup        # make sure esc and github-release are installed
make pack         # pack templates and manifests into go sources
make              # do a local build
make compress     # compress the built executable
sudo make install # install the executable to /usr/local/bin/
```

## Documentation

Documentation is done using [MkDocs](https://github.com/mkdocs/mkdocs).

Documentation files are MarkDown formatted text files. These are in the `doc/` hierarchy.

Run the following to view and edit the documentation locally:

```sh
make serve-docs
```

Navigate to [http://localhost:8000](http://localhost:8000)

Update the documentation sources located in the repository in `doc/` (and its subdirectories) and the mkdocs development server will live-reload the pages as soon as changed.

## Common Issues

* `build command-line-arguments: cannot load github.com/moshloop/karina/manifests: no matching versions for query "latest"`

You didn't run `make pack` to generate the golang sources embedding the template and manifest files.

Run:
```sh
make pack
```

This generates the `static.go` files in the `manifests/` and `templates/` directories.
