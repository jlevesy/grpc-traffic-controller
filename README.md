# gTC: gRPC Traffic Controller

gRPC Traffic Controller (or gTC) is a Kubernetes controller that allows through a manifest to describe routing configuration between gRPC clients and servers. It provides configuration to the clients using the gRPC xDS service discovery integration available in all major gRPC implemetations.

Some of the gRPC features supported by gTC:

- Traffic Splitting and Routing
- Weighted Load Balancing
- Circuit breaking
- Retries
- Fault injection
- Locality Fallback
- Hash Ring Load Balancing
- Topology Aware Routing, if a destination service has [TAR enabled](https://kubernetes.io/docs/concepts/services-networking/topology-aware-routing/), gTC will serve the hinted endpoints with a higher priority.

Some features I wish to add:

- Prometheus Metrics
- Integration of SPIFFE and SPIRE, for both TLS and RBAC.

## Documentation

Please refer to the [official website](https://jlevesy.github.io/grpc-traffic-controller).

## Usage Examples

See [the example setup](./example/k8s/echo-server/1-grpc-service.yaml).

## Current Status

xDS features implemented in gRPC are listed [here](https://grpc.github.io/grpc/cpp/md_doc_grpc_xds_features.html), this table tracks their support in gTC.

| gRFC  | Status |
| ------------- | ------------- |
| [A27](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md) | Supported (except LRS) | N/A (initial implementation) |
| [A28](https://github.com/grpc/proposal/blob/master/A28-xds-traffic-splitting-and-routing.md)  | Supported |
| [A29](https://github.com/grpc/proposal/blob/master/A29-xds-tls-security.md)  | TODO |
| [A31](https://github.com/grpc/proposal/blob/master/A31-xds-timeout-support-and-config-selector.md)  | Supported: MaxStreamDuration on routes and HTTPConnManager. |
| [A32](https://github.com/grpc/proposal/blob/master/A32-xds-circuit-breaking.md)  | Supported: Cluster MaxRequests |
| [A33](https://github.com/grpc/proposal/blob/master/A33-Fault-Injection.md)  | Supported: delay and abort injection |
| [A36](https://github.com/grpc/proposal/blob/master/A36-xds-for-servers.md)  | TODO |
| [A39](https://github.com/grpc/proposal/blob/master/A39-xds-http-filters.md)  | Supported filters at listener, route and backend level |
| [A40](https://github.com/grpc/proposal/blob/master/A40-csds-support.md)  | TODO, Not directly related but it highlight the need of supporting CSDS on gTC's end? |
| [A41](https://github.com/grpc/proposal/blob/master/A41-xds-rbac.md)  | TODO |
| [A42](https://github.com/grpc/proposal/blob/master/A42-xds-ring-hash-lb-policy.md) | Supported: Route Hash Policies and LB Policy on backend |
| [A44](https://github.com/grpc/proposal/blob/master/A44-xds-retry.md)  | Supported, both on route and listener |

- I indend to suport xDS enabled gRPC servers, yet it might require a slight API change, or even a new CRD. More thinking is needed here.
- LRS server side is left out of scope at the moment, though it could be an interesting thing to elaborate (expose load metrics?) I am unsure of what to do with for now.

## Getting Started

### Required Tools

- [go1.21](https://go.dev/learn/)
- [k3d](https://github.com/k3d-io/k3d)
- [ko](https://github.com/google/ko)
- [helm](https://helm.sh/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)

### Running the development environment

For the first time you need to install the code generation tools

```bash
make install_code_generator install_controller_tools
```

Then you can run

```bash
make dev
```

This command:

1. Creates a k3d cluster
2. Installs the gTC controller
3. Deploy an example server
4. Deploys an example client.

From there you can run a few example commands

```bash
make client_shell_0 CMD='/ko-app/client -period=100ms --addr xds:///echo-server/basic  "hello there"'
make client_shell_0 CMD='/ko-app/client -premium -period=100ms --addr xds:///echo-server/abort-fault-injection-backend-override  "hello there"'
```

Feel free to try out [all the examples available](./example/k8s/echo-server/1-grpc-service.yaml)

If you wish to run the test suite

```bash
make test
```
