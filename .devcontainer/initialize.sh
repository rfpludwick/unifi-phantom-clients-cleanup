#!/usr/bin/env bash

set -e
set -o pipefail

# Local initializer?
if [ -x .devcontainer/initialize-local.sh ]; then
	.devcontainer/initialize-local.sh
fi
