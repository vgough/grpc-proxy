#!/usr/bin/env bash

: ${IMG:=vgough/protoc-go:go1.10-proto351}
: ${DIRECT:=false} # Don't set unless running in a container.
: ${DEBUG:=false} # Set true to see command lines.
: ${CMD:=/build/gen-pb.sh}
: ${PLUGIN:=gogoslick}
: ${EXCLUDE:=vendor}

if [[ "${DEBUG}" == "true" ]]; then
  set -x
fi

generate () {
  PTYPES=github.com/gogo/protobuf/types
  GENGO=github.com/golang/protobuf/protoc-gen-go
  MAP="Mgoogle/protobuf/any.proto=${PTYPES}"

  echo "Compiling protos in $1"
  PATH=$PATH:/root/go/bin protoc -I. --${PLUGIN}_out=${MAP},plugins=grpc:. ${1}/*.proto
}

# prefix_env takes all environment variables in the form env__key=value and
# turns them into a space-separated list in the form [prefix][key]=[value].
prefix_env () {
  local out=''
  prefix=$1
  for key in ${!env__*}
  do
    k=${key#env__}
    value=${!key}
    out="${prefix}${k}=${value} ${out}"
  done
  echo "$out"
}

bootstrap_docker () {
  # Run script from within a docker container.
  echo "Starting docker container ${IMG}"
  env=$(prefix_env "-e ")
  docker run --rm \
    ${env} \
    -w /build/ \
    -v `pwd`:/build \
    -it ${IMG} ${CMD}
}

if [[ "${DIRECT}" == true ]]; then
  for d in $(find ${BASE} -name "*.proto" | xargs -n1 dirname | sort -u); do
    if [[ $d =~ ${EXCLUDE} ]]; then
      continue
    fi
    generate $d
  done

else
  # Use flattened values instead of associate array, for bash3 compatibility.
  env__BASE=${BASE}
  env__DEBUG=${DEBUG}
  env__DIRECT=true
  env__PLUGIN=${PLUGIN}
  env__EXCLUDE=${EXCLUDE}

  bootstrap_docker
fi
