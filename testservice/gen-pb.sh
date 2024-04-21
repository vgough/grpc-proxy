#!/usr/bin/env bash

: ${IMG:=bufbuild/buf:1.30.1}
: ${DEBUG:=false} # Set true to see command lines.

if [[ "${DEBUG}" == "true" ]]; then
  set -x
fi

docker run \
	-u`id -u`:`id -g` \
	-e BUF_CACHE_DIR=/workspace/.cache/buf \
	--volume "$(pwd):/workspace" \
	--workdir /workspace \
	${IMG} generate

