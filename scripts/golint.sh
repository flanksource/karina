#!/bin/bash

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.24.0

make
mkdir -p test-results
GOGC=1 golangci-lint run --verbose --print-resources-usage  --out-format=junit-xml > test-results/lint.xml
