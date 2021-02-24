

# Make targets

* `help`           - Lists all valid targets
* `build`(default) - Build binaries
* `install`        - Installs binary locally (needs admin priviliges)
* `linux`          - Build for Linux
* `darwin`         - Build for Darwin
* `docker`         - Build docker image
* `compress`       - Uses UPX to compress the executable
* `serve-docs`     - Serves the MkDocs docs locally
* `build-api-docs` - Build golang docs
* `build-docs`     - Build MkDocs docs
* `deploy-docs`    - Deploy MkDocs to Netlify

Normal first time use:
```shell
make              # do a local build
make compress     # compress the built executable
sudo make install # install the executable to /usr/local/bin/
```

