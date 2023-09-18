---
categories: ["Guide", "First Steps"]
tags: ["first steps"]
title: "Exposing your first gRPC service"
linkTitle: "First Steps"
weight: 2
description: >
    Expose your first gRPC service.
---

## Setting up the scene

For the context of this page, let's imagine that we're running in our production environment a very important gRPC service called `Echo`. This service, as its name indicates, writes back the content of any incoming requests as payload. As we said, very important.

This gRPC service is served by single go application called `echo-server`. This application is running in our production Kubernetes cluster inside a dedicated namespace called `echo-server`. Also the app ir managed by a Kubernetes Deployment configured to run 2 replicas and is exposed by a Kubernetes service also called `echo-server`.

## Exposing the echo-server in grpc-traffic-controller

To expose a workload in gRPC Traffic Controller, the first step is to define a `GRPCListener` alongside the application Kubernetes service.

The gRPC listener looks like the following:

```yaml
---
apiVersion: api.gtc.dev/v1alpha1
kind: GRPCListener
metadata:
  name: basic
  namespace: echo-server
spec:
  routes:
    - backends:
        - service:
             name: echo-server-v1
             port:
                name: grpc
```

If you're familliar with the
