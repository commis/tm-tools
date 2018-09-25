#!/bin/bash

TM_ROOT=$(cd `dirname $(readlink -f "$0")`/.. && pwd)
TM_DATA=${TM_ROOT}/tm-data
TM_TOOL="${TM_ROOT}/bin/tm_tools"
TM_VIEW=${TM_DATA}/result

OLD_VER="v0.18.0"
NEW_VER="v0.23.1"

function view_all() {
    rstdir=$1
    verdir=$2
    database=$3
    output=${TM_VIEW}/${rstdir}/${database}_all.txt

    dbPath=${TM_DATA}/{VER_DIR}/tendermint/data/${database}
    db=$(echo $dbPath |sed "s/{VER_DIR}/$verdir/g")

    ${TM_TOOL} view --db ${db} --a getall |sort -n -k2 -t: > ${output} 
    
    echo "view all ${database} for ${verdir} finished."
}

function view_detail_info() {
    rstdir=$1
    verdir=$2
    database=$3
    params=$4
    output=${TM_VIEW}/${rstdir}/${database}
    mkdir -p ${output}
    
    dbPath=${TM_DATA}/{VER_DIR}/tendermint/data/${database}
    db=$(echo $dbPath |sed "s/{VER_DIR}/$verdir/g")

    srcfile=${TM_VIEW}/${rstdir}/${database}_all.txt
    while read line; do
        outfile=${output}/$(echo $line |sed 's/:/_/g').txt
        ${TM_TOOL} view --db ${db} --q $line ${params} |jq . > ${outfile} 
        # echo "${TM_TOOL} view --db ${db} --q $line ${params} |jq ."
    done < ${srcfile}
    
    echo "view all of ${database} for ${verdir} finished."
}

function migrate_all() {
    oldPath=${TM_DATA}/${1}/tendermint
    newPath=${TM_DATA}/${2}/tendermint

    ${TM_TOOL} migrate --old ${oldPath} --new ${newPath}
    # echo "${TM_TOOL} migrate --old ${oldPath} --new ${newPath}"
    
    echo "migrate all finished."
}

function recover_height() {
    verdir=$1
    height=$3
    params=""
    if [ "$2" == "new" ]; then params="${params} --v"; fi

    dbPath=${TM_DATA}/{VER_DIR}/tendermint
    db=$(echo $dbPath |sed "s/{VER_DIR}/$verdir/g")

    ${TM_TOOL} recover --db ${db} --h ${height} ${params}
    echo "${TM_TOOL} recover --db ${db} --h ${height} ${params}"
    
    echo "recover block height finished."
}

function view_version_data() {
    rstdir=$1
    verdir=$2
    params="--d"
    if [ "$3" == "new" ]; then params="${params} --v"; fi
    mkdir -p ${TM_VIEW}/${rstdir}

    view_all ${rstdir} ${verdir} "blockstore"
    view_detail_info ${rstdir} ${verdir} "blockstore" "${params}"
    
    view_all ${rstdir} ${verdir} "state"
    view_detail_info ${rstdir} ${verdir} "state" "${params}"
    
    # view_all ${rstdir} ${verdir} "evidence"
    # view_all ${rstdir} ${verdir} "trusthistory"
}

function do_upgrade() {
    if [ -d "${TM_VIEW} " ]; then rm -rf ${TM_VIEW}/*; fi
    mkdir -p ${TM_VIEW}

    view_version_data "${OLD_VER}" ${OLD_VER} "old"
    migrate_all ${OLD_VER} ${NEW_VER}
    view_version_data "${NEW_VER}" ${NEW_VER} "new"
   
    echo "do upgrade success."
}

function do_recover() {
    if [ -d "${TM_VIEW} " ]; then rm -rf ${TM_VIEW}/*; fi
    mkdir -p ${TM_VIEW}
    
    view_version_data "${OLD_VER}-r" ${OLD_VER} "old"
    recover_height ${OLD_VER} "old" 85
    view_version_data "${OLD_VER}-t" ${OLD_VER} "old"
    
    echo "do recover block success."
}

# main function
function main() {
    # do_upgrade
    do_recover
}
main 2>&1 |grep -v 'duplicate proto'
