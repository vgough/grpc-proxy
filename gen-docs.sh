#!/bin/bash
# Script that checks the code for errors.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"

function generate_markdown {
    echo "Generating markdown"
    oldpwd=$(pwd)
    for i in $(find . -path ./vendor -prune -o -iname 'doc.go'); do
        dir=${i%/*}
        echo "$dir"
        cd ${dir}
        ${GOPATH}/bin/godocdown -heading=Title -o DOC.md
        ln -s DOC.md README.md 2> /dev/null # can fail
        cd ${oldpwd}
    done;
}

generate_markdown
echo "returning $?"
