<p align="center">
  <img src="https://cloud.githubusercontent.com/assets/4391815/26681386/05b857c4-46ab-11e7-8c71-15a46d886834.png">
</p>


<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2018 Turbonomic

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# prometurbo

<img width="717" alt="prometurbo" src="https://user-images.githubusercontent.com/27221807/41005210-91c0c8f2-68ea-11e8-95be-7599610383aa.png">

## Overview
Prometurbo is a framework to get metrics from Prometheus for Turbonomic.
It is implemented as a GO SDK probe that aims to discover applications and nodes from [Prometheus](https://prometheus.io/) for the Turbonomic Operations Manager.


As of currently, this probe supports:
* Creating Application entities based on the Prometheus [istio](https://istio.io/docs/reference/config/adapters/prometheus.html)
and the [redis](https://github.com/oliver006/redis_exporter) exporters.  More will be gradually added in the future.
* Collecting app response time and transaction data.  More will be gradually added in the future.

## Prerequisites
* Turbonomic 6.2+ 
* Kubernetes 1.7.3+
* Istio 0.3+ (with Prometheus addon)
* supported exporters (as listed above).

## Prometurbo Installation
* [Deploy Prometurbo](https://github.com/turbonomic/prometurbo/prometurbo/tree/master/deploy)
* Once deployed, corresponding targets will show up in Turbonomic UI

