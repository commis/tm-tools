#!/bin/bash

set -e

ROOT_PATH=$(cd $(dirname $0) && pwd)

go build -o "bin/tm_tools" cmd/*.go

if [ -f "bin/tm_tools" ]; then
    ./bin/tm_tools
fi
