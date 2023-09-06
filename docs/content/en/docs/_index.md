
---
title: "gRPC Traffic Controller documentation"
linkTitle: "Documentation"
weight: 20
menu:
  main:
    weight: 20
---

Welcome to the gRPC Traffic Controller Documentation!

gRPC Traffic Controller is an xDS control plane for gRPC and Kubernetes that allows any gRPC client embedding the support of the xDS service discovery protocol to retrieve routing policies.

It brings to any application currently using gRPC (in a decently recent version) routing, loadbalancing, observability and security features comparable to what service meshes offer without the actual complexity of running a service mesh.

All you need is gRPC embeded in your client application, gTC and a few lines of YAML!

## Proxyless service mesh

A service mesh is an infrastructure component in charge of controlling and observing service to service communication in a microservice architecture. It brings major features on the table like:

- Advanced routing and Lodabalacing
- Observability
- End to end encryption of network communication.

We can distinguish two major components in service meshes architecture:

- The Control Plane: in charge of retrieving and exposing routing configuration
- Data Plane: in charge of routing actual traffic based on the configuration exposed by control planes.

Nowardays, service meshes like [Istio](https://istio.io/latest/about/service-mesh/) or [Linkerd](https://linkerd.io/2.14/overview/) are implementing the data plane using a sidecar proxy ([Envoy](https://www.envoyproxy.io/) in the case of Istio) that handles all the traffic from the local pod.

This means running a dedicated proxy alongside each application member of the mesh. This brings, amongst other issues, resources and maintainance overhead. From a security perspective, the traffic between the application and the local sidecar proxy isn't encypted as well.

In an attempt to make this technology simpler and more accessible, the gRPC community has been actively working on delivering the idea of proxyless service mesh. This idea boils down to letting the gRPC client itself handle the routing and loadbalancing of calls, instead of routing all the traffic into a sidecar proxy.

This enables routing, loadbalancing, security and observability capabilities comparable to what we can traditionally find with Service Meshes without the actual complexity and runtime costs of a service mesh.

To realize this idea of proxyless service mesh, the gRPC framework has shipped the support [xDS](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) protocol in all its major implementations.
It is now able to talk directly to an xDS enabled server to retrieve routing policies and implements a very large subset of the features previously provided by sidecars proxies to apply those routing policies.

The good news is: if you're using a recent version of a gRPC implementation, it is likely to already embed the support the xDS protocol. This [gRPC Core documentation](https://grpc.github.io/grpc/core/md_doc_grpc_xds_features.html) provides a per feature support status for every major client implementation.

## Configuring the client

But having gRPC able to route its requests only partially solves the problem: we still need a control plane to deliver the routing configuration!

Currently, only [Google Cloud's traffic director](https://cloud.google.com/traffic-director) and [Istio](https://istio.io/v1.15/blog/2021/proxyless-grpc/) offer support for the gRPC xDS service discovery. Sadly, the first one requires to run on Google Cloud and the second provides experimental support since 2021.

This is where gRPC Traffic Controller comes in!

- Simple: It aims to provide a simple-yet-powerful Kubernetes API to configure gRPC clients. It is inspired by the `networking/v1` `Ingress` resource which, despite its limitiations, has become the actual standard for defining ingress routing configuration.
- Minimal infrastructure required: only an init container and the gRPC traffic controller deployment
- Universal: As long as you run a Kubernetes 1.21+ (first version who supports the `discovery/v1` API) you can use gRPC Traffic Controller!
- Atomic: It aims to do one thing, service discovery. Instead of packing everything inside a single bundle, we prefer integrating with other components of the Kubernetes ecosystem to deliver features like observability ([via the Prometheus ecosystem](https://prometheus.io/)) or end to end encryption (using [SPIFFE and SPIRE](https://spiffe.io/)).

Interested? Please check out the following sections!
