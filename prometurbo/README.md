

# prometurbo

<img width="800" alt="probe" src="https://user-images.githubusercontent.com/27221807/41060588-0ae09bdc-699e-11e8-9a3e-a07e26d07491.png">


## Overview

This is a SDK probe which lies between [`appMetric`](../appmetric) and Turbonomic Operations Manager server. 
It communicates with Turbonomic server to do registration/validation/discovery via a websocket connection. It communicates with [`appMetric`](../appmetric) via a REST interface to get entity metrics on receiving `discovery` command from Turbonomic server.

It does not yet support the discovery of the topological relationships among entities. Instead, it only discovers individual entities and their metrics with support for some hardcoded relationships.


## Prerequisites
* Turbonomic 6.2+ installation
* Install [`appMetric`](../appmetric)

