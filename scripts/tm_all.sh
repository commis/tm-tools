#!/bin/bash

TM_SCRIPT=$(cd `dirname $(readlink -f "$0")` && pwd)
TM_TOOL=${TM_SCRIPT}/tm_upgrade.sh

function set_upgrade_node() {
    name=$1
    cfgFile=${TM_SCRIPT}/.env
    
    value="OLD_VER=${name}"
    lineNo=$(sed -n '/^OLD_VER/=' ${cfgFile})
    sed -i "${lineNo}c $(echo ${value})" ${cfgFile}
}

# main function
function main() {
    nodes=$(docker ps -a --format "{{.Names}}" |sort)
    for i in ${nodes}; do
        echo "upgrade node $i ..."
        set_upgrade_node ${i}
        sudo ${TM_TOOL}
        # echo "sh ${TM_TOOL}"
    done
}
main
