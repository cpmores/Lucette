# 目标文件结构（Target Layout）

> **这是目标，不是现状。** 下面绝大多数文件还不存在。仓库里目前只有 `cmd/node/main.go`、
> `docs/architecture.md` 和这份地图本身。**不要把这份地图当成进度报告读。**

这份地图是终点，不是施工顺序。按纪律 A/C，实现时每个文件都必须由**消费者**拽出来：
先有一个失败的测试，再写让它通过的最小代码。没有消费者的类型面一律不写，或写成
`// TODO(<分支名>): 无消费者`。

里程碑按"并发何时进入"切，不按文件多少切：

| 里程碑       | 范围                                                             | 新增的能力                          | 首次引入的复杂度           |
| ------------ | ---------------------------------------------------------------- | ----------------------------------- | -------------------------- |
| **M1** | `core/` + `sim/`                                             | 用仿真证明`Place` 与代价模型      | 无：零 I/O、零 goroutine   |
| **M2** | `internal/` + `wire/` + `contract/` + `transport/inproc` | 执行面：租约、围栏、幂等、reconcile | 并发、`-race` 从这里开始 |
| **M3** | `transport/{libp2p,mqtt,https}`、真实 provider、装配           | 真实边缘与云                        | 真实网络、真实硬件         |

## 结构

```
cmd/
├── node/main.go            装配与启动顺序。没有逻辑。
└── simbench/main.go        跑基准、输出 CSV 与图表

core/                       ★ 冻结核心:仅 stdlib,零 I/O,零 goroutine
├── id.go                   NodeID / WorkID / SessionKey / LocalityLabel
├── capability.go           Capability / ModelResidency / CostModel
├── work.go                 WorkUnit / Requirements / HardConstraints /
│                           SoftPreferences / WorkKind
│                           ← 需求声明 DSL 在这里(Requires/Affinity/Locality/Fallback/CostCeiling)
├── decision.go             Decision / Outcome{Placed, Deferred, Rejected}
├── graph.go                ResourceGraph / ResidencyState(只读视图)
├── place.go                Placer 接口
├── place_naive.go          基线:RoundRobin / CapabilityOnly
├── place_locality.go       主策略:LocalityAware
├── cost.go                 代价模型 —— 全项目唯一的核心假设
└── invariants_test.go      六条不变量(纯函数可测:幂等、围栏、分区自持、
                            收敛非重放、位置强制、丢事件不损正确性)

wire/                       ★ 跨线的东西:不 import core
├── envelope.go             Envelope{ID, Type, Version, From, To, TS, Payload raw}
├── codec.go                按 (Type, Version) 注册;未识别类型 → 显式 error
├── messages.go             各版本的线消息类型(AdvertiseV1 / OfferV1 / ResultV1 …)
│                           ← 与领域类型刻意分开,边界处显式转换
└── codec_test.go           未知类型必须报错、版本不匹配必须报错

contract/                   ★ 四个冻结接口:插件作者的唯一依赖(只 import core)
├── provider.go             ModelProvider
├── transport.go            Transport
├── tool.go                 Tool
└── telemetry.go            TelemetryCollector

transport/                  edge:实现 contract.Transport(用 wire 编解码)
├── inproc/                 单机与测试 —— 让"自放置"和真实远端走同一条 offer 路径
├── libp2p/                 LAN / P2P
├── mqtt/                   工厂
└── https/                  云

provider/                   edge:实现 contract.ModelProvider
├── fake/                   ★ 确定性 —— 仿真与可复现实验的前提,不是凑数
├── vllm/
├── ollama/
└── cloud/

internal/                   runtime:不属于公开接口
├── plane/
│   ├── advertise.go        周期发布自身能力(带 TTL)
│   └── graph.go            剔除过期广告 → 产出资源图视图
├── work/
│   ├── lease.go            租约 FSM(单一租约权威,不设双重机制)
│   ├── fence.go            fencing token 签发与校验
│   ├── store.go            带围栏的状态存储
│   └── idempotency.go      幂等键
├── authority/              拥有者角色
│   ├── plan.go             持有 plan/DAG,派生后继提升(不是命令式级联)
│   ├── reconcile.go        ★ 只扫"我拥有的工作",会调用 core.Place
│   └── offer.go            定向 offer + TTL
└── worker/                 持有者角色
    ├── admit.go            准入判定(不含 Place)
    ├── reconcile.go        ★ 只扫"我持有的 offer",绝不调 Place
    └── execute.go          执行与回报

sim/                        benchmark:只依赖 core
├── topology.go             合成异构拓扑
├── workload.go             多步会话负载(ReAct 式)
├── chaos.go                分区/延迟/丢包/重启注入(M2 才引入)
├── metrics.go              机制指标(prefill miss / model load / eviction /
│                            locality hit)+ 换算指标(latency / cost)
├── bench_test.go           五条基线对比(含 oracle 上界)
└── sensitivity_test.go     代价比例扫描 → 交叉点 → 论文图表

configs/{home.yaml,factory.yaml}     ops:运行时配置示例
scripts/{check_deps.sh,two_nodes.sh} ops:开发脚本
docs/                               docs:architecture.md / adr/ / placement.md /
                                    protocol.md / scenarios.md / thesis/
```

两处结构上的要点值得保护，改动前先想清楚：

+ `internal/authority/reconcile.go` **会**调用 `core.Place`，`internal/worker/reconcile.go` **绝不**调用。
  这个不对称就是"位置强制"不变量落在文件上的形状：调度器不可信，节点侧必须自己过闸门。
  一旦 worker 也去 `Place`，双闸门就退化成了双重派发。
+ `sim/bench_test.go` 里有 **oracle 上界**。没有它，结果只能自己跟自己比；有它，才能说
  "locality-aware 拿到了多少最优差距"。这条要一直留着。

## M1：`core/` + `sim/`

只碰这两个包，其余一行不写。真正要交付的不是"一个能跑的 placer"，而是一个**交叉点条件**。

+ `core/`：`id.go`、`capability.go`、`work.go`、`graph.go`、`decision.go`、`cost.go`、
  `place.go`、`place_naive.go`、`place_locality.go`
+ `sim/`：`topology.go`、`workload.go`、`metrics.go`、`bench_test.go`、`sensitivity_test.go`
+ 顺序：**先写失败的测试，再让编译器把 `core` 拽出来**。即先 `sim/workload.go`（三步会话，带
  seed）→ `sim/topology.go`（一个节点持有该会话的 KV cache 且模型常驻，一个两样都没有）→
  `sim/metrics.go` → `sim/bench_test.go`（失败：step 2 必须落在持有 KV cache 的节点上，
  naive 不会）→ 然后才逐个写 `core/*`，每个都只因编译错误而存在

`core/cost.go` 是"全项目唯一的核心假设"，所以这里藏着项目最大的智力风险：系数若是拍的，
实验就在演示"模拟器相信 locality 有用"，而不是"locality 有用"。让 claim 诚实起来的办法是
把它**条件化**——"当 KV 复用的节省 / 模型加载开销 > R 时，locality-aware 优于朴素放置"，
于是 R 就是贡献，交叉点就是图表。因此 `sensitivity_test.go` 是主结果，不是收尾工作；
写 `cost.go` 之前应当先能回答每个系数从哪来（实测 / 公开数据 / 可辩护的区间）。

## M2：执行面

`wire/`、`contract/`、`transport/inproc/`、`internal/` 全部，加上 `sim/chaos.go` 与
`provider/fake/`。六条不变量的**完整**测试到这一层才成立（见下方修正一）。

`provider/fake/` 放在 M2 而非 M1：M1 里"模型是否常驻"是 `core.Capability` 的数据，不需要
插件接口；等到真有 worker 去执行时，它才成为必要。若仿真发现确实更早需要，再提前。

## M3：真实边缘

`transport/{libp2p,mqtt,https}`、`provider/{vllm,ollama,cloud}`、`cmd/node/main.go` 的装配、
`configs/`、`scripts/`、`cmd/simbench/main.go`。`cmd/simbench` 的 CSV 与图表也要等到这里——
在指标稳定之前 `go test -bench` 就是那个接口。

## 对原建议的修正

**修正一：`core/invariants_test.go` 声称测六条不变量，但其中五条在纯 `core` 里测不了。**

+ 幂等（副作用只生效一次）、围栏、分区自持、收敛非重放、丢事件不损正确性——全都需要租约、
  fence、状态存储、reconcile 循环与 transport。它们在 M2 才可测。
+ 连"位置强制"也不能在 `core` 里证：该不变量明确要求**节点侧**强制、不信任调度器，而节点侧
  闸门是 `internal/worker/admit.go`。`core` 里能证的只是调度器侧那一半，即弱的那一半。
+ 纯 `core` 真正能证明的是另外四条：**`Place` 的确定性**（同输入同输出，这是仿真可复现的
  前提，与工作幂等是两件事）、**陈旧只退化为重试而绝不错误放置**、**`Rejected` 的终局性**
  （重新 `Place` 仍是 `Rejected`，不翻烧饼）、**`core` 内无时钟**（时间必须显式入参，
  否则 `Deferred` 不可测）。
+ 建议把该文件拆成 `place_test.go` / `graph_test.go` / `cost_test.go`，只放上面四条。

**修正二：`contract/tool.go` 的 `Tool` 没有出处。** 简报通篇未把"工具"当作概念，这是无需求
支撑的抽象（纪律 A）。删掉，或标成 `// TODO: 无消费者`。

**修正三：`core/place.go` 的 `Placer` 接口可以推迟。** 在 `sim` 真的需要遍历多个实现之前，
具名函数（或函数值）就够。等第二个消费者出现再引入接口。

## 待定

+ `wire/` 与 `contract/` 的边界：现在按"`wire/` 放 Envelope + codec + 线消息；`contract/` 放
  插件接口"理解。`contract/` 需不需要 import `wire/`，等 M2 有真实调用点时再定。
+ `docs/architecture.md` 目前只有五行，且把项目称作 "plugin-style **agent task manager**"，
  与 README 的"一个放置函数 + 一个收敛循环"重心不同。两份文档必须收敛到同一个重心，否则
  读者不知道该信哪一份。
