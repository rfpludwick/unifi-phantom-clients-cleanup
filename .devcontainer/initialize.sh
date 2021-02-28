#!/usr/bin/env bash

set -e

# Local initializer?
if [ -x .devcontainer/initialize-local.sh ]; then
	.devcontainer/initialize-local.sh
fi
