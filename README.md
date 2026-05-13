# O-RAN SC SubMgr: gRPC Subscription Support
Implementation of O-RAN WG3 Standardized gRPC Interface for Near-RT RIC Subscription Management.

## Overview
This project extends the O-RAN Software Community (OSC) Subscription Manager (SubMgr) to support gRPC-based Northbound Interfaces (NBI). While the current OSC implementation primarily uses REST, this update aligns with O-RAN WG3 standards to provide high-performance, bidirectional communication between xApps and SubMgr.

### Key Features
- WG3 Standard Compliance: Implemented gRPC services based on official O-RAN RIC API proto definitions.
- Dual-Stack Interface: Support for both REST and gRPC subscription requests in a single SubMgr instance.


## Architecture
- SubMgr (Go): Runs a gRPC server as a non-blocking goroutine, integrated with the existing RMR and SDL logic.
- Connectivity: Includes logic to handle RTMGR (Route Manager) endpoint addressing for gRPC-enabled xApps.

## Usage
### Build

```
docker build -t <registry>/<image_name:tag> .
```
### Deployment (Helm)
```
RELEASE_NAME=$(helm list -n ricplt -q | grep submgr)
helm upgrade ${RELEAS_NAME} ./helm_submgr -n ricplt -f ./helm_submgr/values.yaml
```
Note: Update the image information in values.yaml to match your environment before deployment.
