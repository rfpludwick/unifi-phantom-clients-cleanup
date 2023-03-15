#!/usr/bin/env bash

set -exo pipefail

# Local initializer?
if [ -x .devcontainer/initialize-local.sh ]; then
	.devcontainer/initialize-local.sh
fi
