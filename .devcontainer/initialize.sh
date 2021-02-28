#!/usr/bin/env bash

set -e

# Local initializer?
if [ -x ./initialize-local.sh ]; then
	./initialize-local.sh
fi
