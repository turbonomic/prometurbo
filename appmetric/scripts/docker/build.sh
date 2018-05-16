#!/bin/bash

tag=vmturbo/appmetric:6.2dev

make product
ret=$?
if [ $ret -ne 0 ] ; then
    echo "[`date`] build binary file failed"
    exit 1
fi

export GIT_COMMIT=$(git rev-parse HEAD)
sh scripts/docker/gen.build.info.sh
docker build -t $tag --build-arg GIT_COMMIT=${GIT_COMMIT} .
ret=$?
if [ $ret -ne 0 ] ; then
    echo "[`date`] build docker image failed"
    exit 1
fi

docker push $tag
