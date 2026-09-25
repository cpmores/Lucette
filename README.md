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
+ State and location awareness
  + `Lucette` runtime will get the context state and network location and `Placer` places `WorkUnit` according to those messages.
+ Network partitioning
  + *HetNets* will be devided into different network partitions, including edge net and cloud.
  + *Partitioning* ensures not only **usability**, but also **accuracy**.

## Non-goals

what we will not do

+ a general agent workflow engine, like `LangGraph` or `CrewAI`
+ a agent system micro-kernel like what `dsh` is doing now, we only use basic interfaces.
+ a real-time feedback control loop, for the efficiency of LLMs.

## The name

Why I choose `Lucette`? Yes, it comes from `Lucinda` who is a female lead in **Nabokov**'s book *Ada or Ardor,* almost my favorite character in this noval. Anyway, back to the point,  when I started to build `Lucinda`, I find it hard to satisfy every agent requirement in real world, problems such as context sync or toolbox management are not a thing that we can solve immediately, so I choose to start with something smaller. I tried my best to make it small, at least not a runtime consist of multiple stateful components.

## Getting Started

## Where things live

Every directory below names a layer. Only `cmd/node` and `docs` hold anything yet; the rest are
empty placeholders waiting for the layer they are named after.

| Path | Layer | Holds |
| --- | --- | --- |
| `cmd/node` | entrypoint | `main` and wiring, no logic |
| `core` | placement, work | `Place(WorkUnit, ResourceGraph, Residency) → Decision`, the invariants, fenced idempotent work |
| `wire` | contract | the envelope, the codec, and the versioned wire messages, kept separate from the domain types and converted explicitly at the boundary; an unrecognized type is an error, never a silent drop |
| `contract` | contract | the frozen interfaces a plugin author depends on: model provider, transport, telemetry |
| `transport` | edge | libp2p on LAN/P2P, MQTT in the factory, HTTPS to cloud |
| `provider` | edge | how a plugin author declares requirements: model, VRAM, locality, affinity, cost |
| `internal` | edge | implementation that is not part of the public surface |
| `sim` | benchmark | the simulator, and the only place the efficiency claim can be tested |
| `configs` | ops | runtime configuration |
| `scripts` | ops | developer scripts |
| `docs` | docs | design notes, starting from `docs/architecture.md` and `docs/target-layout.md` |

Three dependency rules are what keep this from turning into a runtime, and they are conventions
for now, not something a machine checks:

+ `core` and `wire` import nothing else in this repo, only the standard library.
+ `sim` depends on `core` alone, so benchmarks still run when the edges are broken.
+ `core` does no I/O and starts no goroutine.

## Documentation

## Contributing

## License

MIT, see [LICENSE](LICENSE).

Copyright (c) 2026 Chaoran Hu.

## Related work
