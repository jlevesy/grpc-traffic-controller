# KxDS

## What ?

This is an attempt to write an xDS control plane for gRPC that provides configuration to gRPC clients based on custom
resources describing the service topology.

A naive CRD could be:

```yaml
---
apiVersion: api.kxds.dev/v1alpha1
kind: Service
metadata:
  name: some-service
  namespace: some-namespace
spec:
   listener: awesome-service
   destination:
     name: some-k8s-service
     port: 1997
```

## Current Status

At the moment, this is a dumb implementation that tells any client to contact a server on localhost:3333

Running the demo

```bash
make run_kxds
make run_server
make send_echo
```
