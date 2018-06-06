<p align="center">
  <img width=300 height=150 src="https://cloud.githubusercontent.com/assets/4391815/26681386/05b857c4-46ab-11e7-8c71-15a46d886834.png">
</p>


# PromeTurbo

<img width="717" alt="prometurbo" src="https://user-images.githubusercontent.com/27221807/41005210-91c0c8f2-68ea-11e8-95be-7599610383aa.png">


## Overview
Prometurbo is a framework to get metrics from Prometheus for Turbonomic.
It is implemented as a SDK probe that aims to discover applications and nodes from [Prometheus](https://prometheus.io/) for the Turbonomic Operations Manager. 

## Two components
It has two components: [`appMetric`](./appmetric) and [`Probe`](prometurbo).

**Main functions of appMetric** : 
   * Get entity metrics from Prometheus, or other sources;
   * Expose the entity metrics via REST API;
   
**Main functions of Probe** :
   * Fetch entity metrics from appMetric, convert them into EntityDTO;
   * Probe regristration with TurboServer;
   * Execute the validation/discovery command from TurboServer;

Since these two components interact with each other via REST API, so they can be deployed in the same pod, or separately as different services.

## Metric sources and entities
Currently, Prometurbo can get metrics from [Istio](https://istio.io/docs/reference/config/adapters/prometheus.html) and [redis](https://github.com/oliver006/redis_exporter) exporters.
It will generate `ResponseTime` and `Transaction` commodities for `Application Entity`.


 More exporters will be supported, and more entity types/commodities will be gradually added in the future.

## Prerequisites
* Turbonomic 6.2+ 
* Kubernetes 1.7.3+
* Istio 0.3+ (with Prometheus addon)
* supported exporters (as listed above).

## Prometurbo Installation
* [Deploy Prometurbo](./deploy)

