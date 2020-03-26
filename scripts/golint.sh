#!/bin/bash

set -ex

GO_MOD_CHECKSUM=$(sha1sum go.mod | awk -F " " '{print $1}')
GO_SUM_CHECKSUM=$(sha1sum go.sum | awk -F " " '{print $1}')

# Check that go.sum and go.mod are correctly checked out
make pack
go mod tidy

NEW_GO_MOD_CHECKSUM=$(sha1sum go.mod | awk -F " " '{print $1}')
NEW_GO_SUM_CHECKSUM=$(sha1sum go.sum | awk -F " " '{print $1}')

if [ "$GO_MOD_CHECKSUM" != "$NEW_GO_MOD_CHECKSUM" ]; then
  echo "go.mod is not up to date. Checksum before 'go mod tidy' was $GO_MOD_CHECKSUM and after was $NEW_GO_MOD_CHECKSUM"
fi

if [ "$GO_SUM_CHECKSUM" != "$NEW_GO_SUM_CHECKSUM" ]; then
  echo "go.sum is not up to date. Checksum before 'go mod tidy' was $GO_GO_SUM_CHECKSUM and after was $NEW_GO_SUM_CHECKSUM"
fi

# Check that no *.yml files are present, only use *.yaml as standard suggests
go run test/linter/main.go

# Run golanci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.24.0
make
GOGC=20 golangci-lint run --verbose --concurrency 1


