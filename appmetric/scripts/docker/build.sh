#!/bin/bash

tag=vmturbo/appmetric:v2.0

make product
ret=$?
if [ $ret -ne 0 ] ; then
    echo "[`date`] build binary file failed"
    exit 1
fi

docker build -t $tag .
ret=$?
if [ $ret -ne 0 ] ; then
    echo "[`date`] build docker image failed"
    exit 1
fi

docker push $tag
