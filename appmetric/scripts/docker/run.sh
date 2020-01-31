#!/bin/bash
## start the docker container in local for testing

##NOTE: should use the ip address, instead of 'localhost' or '127.0.0.1'
url=10.10.200.46:19090

docker run -d -p 18081:8081 turbonomic/appmetric:6.4dev --v=3 --logtostderr
sleep 1
docker ps
