# KxDS

KxDS is an xDS control plane for gRPC xDS service discovery. It is a kubernetes controller that reads service definition through a `Service` custom resources.

## Getting Started

### Required Tools

- [go1.19](https://go.dev/learn/)
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
2. Installs the kxds controller
3. Deploy an example server
4. Deploys an example client.

From there you can run

```bash
k -n echo-client exec -ti echo-client-58bb8b9864-rcfpd -- /ko-app/client --addr xds:///echo-server hello there
```
