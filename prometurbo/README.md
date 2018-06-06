

# prometurbo

<img width="800" alt="probe" src="https://user-images.githubusercontent.com/27221807/41060588-0ae09bdc-699e-11e8-9a3e-a07e26d07491.png">


## Overview

This is a SDK probe which lies between [`appMetric`](../appmetric) and Turbonomic Operations Manager server. 
From the one hand, it communicates with Turbonomic server to do registration/validation/discovery. From the other hand,
it will talk with [`appMetric`](../appmetric) to get entity metrics on receiving `discovery` command from Turbonomic server.

In current implementation, it will not discovery the topology among the entities. Instead, it only generates (proxy) entities with `ResponseTime` and `Transaction` sold commodities.


## Prerequisites
* Turbonomic 6.2+ installation
* Kubernetes 1.7.3+
* Install [`appMetric`](../appmetric)

