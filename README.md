# KxDS

KxDS is an xDS control plane for gRPC xDS service discovery. It is a kubernetes controller that reads service definition through a `Service` custom resources.

## Current thoughts and ideas

- This implementation rebuilds the whole configuration state each time it either receives an endpoints or a XDSService update. This is inneficient and causes the SnapshotCache to resent the whole connfig to each connectec client for any (even non relevant) endpoint update, which is bad. I believe the correct way  to do this is to only update CDS and EDS when we receive a k8s endpoint update. (matching by service name in cache could be super efficient?)
- I'm still very much confused by this idea of nodeID in the case of gRPC clients: the config is the same, independently of who is calling (which is not the case for a standard service mesh here). Buuut, am I shooting myself in the foot by ignoring this dimention? Probably!

## What Could the CRD looks like?

I'm writing this down as I'm reading gRPC-go's implementation of xDS suppoert. I don't know what I'm doing here!

Minimal example

```yaml
apiVersion: api.kxds.dev/v1alpha1
kind: XDSService
metadata:
  name: echo-server-grpc-v1
  namespace: echo-server
spec:
  listener: echo-server-v1
  routes:
  - clusters:
    - name: main-cluster
      localities:
      - service:
          name: echo-server-v1
          port:
            name: xds
```

Full example with all the fields sets.

```yaml
apiVersion: api.kxds.dev/v1alpha1
kind: XDSService
metadata:
  name: echo-server-grpc-v1
  namespace: echo-server
spec:
  listener: echo-server-v1
  maxStreamDuration: 20s
  retry:
    codes:
      - 300
      - 400
    maxAttempts: 10
    backoff:
      baseInterval: 10s
      maxInterval: 60m
  filters:
    - todo?
  routes:
    - matcher:
        path: /foo
        headers:
      maxStreamDuration: 20s
      retry:
        codes:
          - 300
          - 400
        maxAttempts: 10
        backoff:
          baseInterval: 10s
          maxInterval: 60m
      filtersOverrides:
        - todo?
      clusters:
        - name: some-cluster
          security:
            todo: clarify
          maxRequest: 30 // circuit breaking
          lbPolicy: round_robin
          localities:
            - priority 1
              weight 2
              service:
                name: echo-server-v1
                 port:
                name: xds
            - weight: 1
              externalName: https://google.com
```

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
