# Lucette

**Locality-aware task placement** and durable execution for stateful AI inference across intermittently-connected edge and cloud.

Iterally, `Lucette` is a simplified version of my previous project `Lucinda`.

## Status

1. On our first step, we will get started to simulate an basic environment for `Lucette`, under simlulator, we will be able to varify the correctness of our first and most important function `Place`.

## Why

Why `Lucette` is important? Well, for those most popular orchestration layer, they assume that our devices are under a stable environment with sufficient resources, because now most of us are using inference services from cloud providers, yes, it is true, for now. But when the hardwares get faster and better in edge devices, this assumption may lead to serious problems.

## What it does

Core of `Lucinda` is simple, just *a place function* and *a reconcile loop*

+ Core
  + Place: place `WorkUnit` on the node most suitable according to `ResourceGraph`
  + Reconcile: this loop will check the status of placable nodes whether they have right status, if not, force them to.

## Non-goals

## The name

## Getting Started

## Where things live

## Documentation

## Contributing

## License

## Related work
