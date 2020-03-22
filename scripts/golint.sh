#!/bin/bash

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.24.0

.bin/golangci-lint run