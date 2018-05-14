#!/bin/bash

##NOTE: should use the ip address, instead of 'localhost' or '127.0.0.1'
url=10.10.2.3:9090

docker run -d -p 18081:8081 beekman9527/appmetric:v2 --promUrl=$url --v=3 --logtostderr
sleep 1
docker ps
