#!/bin/bash

# set -e

ROOT_PATH=$(cd $(dirname $0) && pwd)

# copy code to tendermint dir
function copy_code() {
    curr=$(pwd)
    tagdir=$1

    if [ -d "${tagdir}" ]; then
        newFile=${tagdir}/wal_up.go
        cp -f ${ROOT_PATH}/libs/cli/wal_up.new ${newFile}
        chmod 777 ${newFile}
    fi
}
copy_code ${GOPATH}/src/github.com/tendermint/tendermint/consensus
copy_code ${GOPATH}/src/github.com/commis/tm-tools/vendor/github.com/tendermint/tendermint/consensus

go build -o "bin/tm_tools" cmd/*.go

# test tools command
if [ -f "bin/tm_tools" ]; then
    ./bin/tm_tools --help 2>&1 |grep -v 'duplicate proto'
fi

# copy to test tools
function copy_dist() {
    testdir=$1

    if [ -d ${testdir} ]; then
        cp -f ${ROOT_PATH}/bin/tm_tools ${testdir}/tm_tools
        chmod 777 ${testdir}/tm_tools
    fi
}
copy_dist "${ROOT_PATH}/../tm-network/tools/0.23.1"