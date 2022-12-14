#!/bin/bash

set -e
trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG
trap 'CMD=${last_command} RET=$?; if [[ $RET -ne 0 ]]; then echo "\"${CMD}\" command failed with exit code $RET."; fi' EXIT
SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
cd ${SCRIPTPATH}

go get -u golang.org/x/lint/golint
go test . -count=1

# gofmt always returns 0... so we need to capture the output and test it

FMT=$(gofmt -l -s *.go)
if ! [ -z ${FMT} ]; then
    echo ''
    echo "Some files have bad style. Run the following commands to fix them:"
    for LINE in ${FMT}
    do
        echo "  gofmt -s -w `pwd`/${LINE}"
    done
    echo ''
    exit 1
fi
