#!/bin/bash

TM_ROOT=$(cd `dirname $(readlink -f "$0")`/.. && pwd)
TM_DATA=${TM_ROOT}/tm-data
TM_TOOL="${TM_ROOT}/bin/tm_tools"
TM_VIEW=${TM_DATA}/result

OLD_VER="v0.18.0"
NEW_VER="v0.23.1"

function view_all() {
    ver=$1
    database=$2
    output=${TM_VIEW}/${ver}/${database}_all.txt

    dbPath=${TM_DATA}/{VER_DIR}/tendermint/data/${database}
    db=$(echo $dbPath |sed "s/{VER_DIR}/$ver/g")
    ${TM_TOOL} view --db ${db} --a getall |sort -n -k2 -t: > ${output} 
    
    echo "view all ${database} for ${ver} finished."
}

function view_detail_info() {
    ver=$1
    database=$2
    params=$3
    output=${TM_VIEW}/${ver}/${database}
    mkdir -p ${output}
    
    dbPath=${TM_DATA}/{VER_DIR}/tendermint/data/${database}
    db=$(echo $dbPath |sed "s/{VER_DIR}/$ver/g")

    srcfile=${TM_VIEW}/${ver}/${database}_all.txt
    while read line; do
        outfile=${output}/$(echo $line |sed 's/:/_/g').txt
        ${TM_TOOL} view --db ${db} --q $line ${params} |jq . > ${outfile} 
    done < ${srcfile}
    
    echo "view all of ${database} for ${ver} finished."
}

function view_evidence() {
    echo "not impl"
}

function migrate_all() {
    oldPath=${TM_DATA}/${1}/tendermint
    newPath=${TM_DATA}/${2}/tendermint

    ${TM_TOOL} migrate --old ${oldPath} --new ${newPath}
    
    echo "migrate all finished."
}

function view_version_data() {
    verdir=$1
    params="--d"
    if [ "$2" == "new" ]; then params="${params} --v"; fi

    view_all ${verdir} "blockstore"
    view_all ${verdir} "state"
    view_detail_info ${verdir} "blockstore" "${params}"
    view_detail_info ${verdir} "state" "${params}"
}

# main function
function main() {
    if [ -d "${TM_VIEW}/ " ]; then rm -rf ${TM_VIEW}/*; fi
    mkdir -p ${TM_VIEW} ${TM_VIEW}/${OLD_VER} ${TM_VIEW}/${NEW_VER}

    view_version_data ${OLD_VER} "old"
    migrate_all ${OLD_VER} ${NEW_VER}
    view_version_data ${NEW_VER} "new"
   
    echo "execute success."
}
main 2>&1 |grep -v 'duplicate proto'
