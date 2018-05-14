#!/bin/bash


localport=19090
namespace=istio-system
kubectl -n $namespace port-forward $(kubectl -n $namespace get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') $localport:9090 

