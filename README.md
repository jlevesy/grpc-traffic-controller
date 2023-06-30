# gTC: gRPC Traffic Controller

gRPC Traffic Controller (or gTC) is a Kubernetes controller that allows through a simple manifest to set up and describe routing configuration between gRPC clients and servers. It provides confiuguration to the clients using the gRPC xDS service discovery integration available in most of the major gRPC distributions.

## Usage Examples

See [the example setup](./example/k8s/echo-server/1-grpc-service.yaml).

## Current Status

I'm currently rebuilding this API.

xDS features implemented in gRPC are listed [here](https://grpc.github.io/grpc/cpp/md_doc_grpc_xds_features.html), the table tracks their support in gTC.

| gRFC  | Status |
| ------------- | ------------- |
| [A27](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md) | Supported (except LRS) | N/A (initial implementation) |
| [A28](https://github.com/grpc/proposal/blob/master/A28-xds-traffic-splitting-and-routing.md)  | Supported |
| [A31](https://github.com/grpc/proposal/blob/master/A31-xds-timeout-support-and-config-selector.md)  | Supported: MaxStreamDuration on routes and HTTPConnManager. |
| [A32](https://github.com/grpc/proposal/blob/master/A32-xds-circuit-breaking.md)  | Supported: Cluster MaxRequests |
| [A33](https://github.com/grpc/proposal/blob/master/A33-Fault-Injection.md)  | TODO |
| [A42](https://github.com/grpc/proposal/blob/master/A42-xds-ring-hash-lb-policy.md) | TODO |
| [A44](https://github.com/grpc/proposal/blob/master/A44-xds-retry.md)  | TODO |
| [A29](https://github.com/grpc/proposal/blob/master/A29-xds-tls-security.md)  | TODO |
| [A41](https://github.com/grpc/proposal/blob/master/A41-xds-rbac.md)  | TODO |
| [A36](https://github.com/grpc/proposal/blob/master/A36-xds-for-servers.md)  | TODO |
| [A40](https://github.com/grpc/proposal/blob/master/A40-csds-support.md)  | TODO, Not directly related but it highlight the need of supporting CSDS on gTC's end? |
| [A39](https://github.com/grpc/proposal/blob/master/A39-xds-http-filters.md)  | TODO |

- I indend to suport xDS enabled gRPC servers, yet it might require a slight API change, or even a new CRD. More thinking is needed here.
- LRS server side is left out of scope at the moment, though it could be an interesting thing to elaborate (expose load metrics?) I am unsure of what to do with for now.

## Getting Started

### Required Tools

- [go1.20](https://go.dev/learn/)
- [k3d](https://github.com/k3d-io/k3d)
- [ko](https://github.com/google/ko)
- [helm](https://helm.sh/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)

### Running the development environment

```bash
make dev
```

This command:

1. Creates a k3d cluster
2. Installs the gTC controller
3. Deploy an example server
4. Deploys an example client.

From there you can run

```bash
make client_shell

/ko-app/client --addr xds:///echo-server hello there
```
