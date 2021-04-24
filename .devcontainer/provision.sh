#!/usr/bin/env bash

set -e

wget https://golang.org/dl/go1.16.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.16.linux-amd64.tar.gz
rm go1.16.linux-amd64.tar.gz
export GOPATH=/usr/lib/go
go get -v golang.org/x/tools/gopls
go get -v github.com/uudashr/gopkgs/v2/cmd/gopkgs
go get -v github.com/ramya-rao-a/go-outline
go get -v github.com/go-delve/delve/cmd/dlv
go get -v golang.org/x/lint/golint
go get -v golang.org/x/tools/gopls
unset GOPATH
go mod tidy

# Local provisioner?
if [ -x .devcontainer/provision-local.sh ]; then
	.devcontainer/provision-local.sh
fi
