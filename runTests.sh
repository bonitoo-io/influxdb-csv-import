#!/bin/bash
set -e
GOROOT=~/sdk/go1.13.9
PATH=$GOROOT/bin:$PATH
go vet
# tests
go test ./...
go tool cover -html=cover.out -o=cover.html
#xdg-open ./cover.html

# editor checker
GO111MODULE=on go run github.com/editorconfig-checker/editorconfig-checker/cmd/editorconfig-checker
GO111MODULE=on go run github.com/editorconfig-checker/editorconfig-checker/cmd/editorconfig-checker
# static analysis
GO111MODULE=on go install honnef.co/go/tools/cmd/staticcheck
GO111MODULE=on ~/go/bin/staticcheck ./write
# ensure tidy module dependendencies
export GO111MODULE=on
go mod tidy
if ! git --no-pager diff --exit-code -- go.mod go.sum; then
  >&2 echo "modules are not tidy, please run 'go mod tidy'"
  exit 1
fi
