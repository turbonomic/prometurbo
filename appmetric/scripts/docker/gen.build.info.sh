#!/bin/sh

fname=./build.info

echo "Name: appmetric" > $fname
echo "GIT_COMMIT: $GIT_COMMIT" >> $fname
#echo "GIT_Head: `git rev-parse head`" >> $fname
echo "Build_date: `date` " >> $fname
