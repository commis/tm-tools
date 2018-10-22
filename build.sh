#!/bin/bash

# set -eux

ROOT_PATH=$(cd $(dirname $0) && pwd)

# copy code to tendermint dir
dir="${GOPATH}/src/github.com/tendermint/tendermint/consensus"
if [ -d "${dir}" ]; then
    newFile=${dir}/wal_up.go
    cp -f ${ROOT_PATH}/libs/cli/wal_up.new ${newFile}
    chmod 777 ${newFile}
fi

go build -o "bin/tm_tools" cmd/*.go

# test tools command
if [ -f "bin/tm_tools" ]; then
    ./bin/tm_tools --help 2>&1 |grep -v 'duplicate proto'
fi

# copy to test tools
testdir=${ROOT_PATH}/../tm-network/tools/0.23.1
if [ -d ${testdir} ]; then
    cp -f ${ROOT_PATH}/bin/tm_tools ${testdir}/tm_tools
    chmod 777 ${testdir}/tm_tools
fi
