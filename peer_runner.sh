#!/bin/sh
set -eu

peer=$1
dispatcher=$2

if [ ! -d log ]; then
  mkdir log
fi

name=$(printf 'peer%02d' "$peer")
date=$(date '+%Y%m%d_%H:%M:%S')
cd cmd/peer &&
  air --tmp_dir "../../../tmp" --build.bin "../../../tmp/main" --build.cmd "go build -o ../../../tmp/main ." \
    --build.args_bin "--name $name --port $((peer + 3000)) --dispatcher $dispatcher" >../../log/"${name}"_"${date}".log 2>&1
