#!/usr/bin/env bash

set -e

# Local provisioner?
if [ -x .devcontainer/provision-local.sh ]; then
	.devcontainer/provision-local.sh
fi
