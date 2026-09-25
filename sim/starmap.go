// Package sim contains several simulation for Lucette
package sim

import (
	"crypto/sha256"

	"github.com/cpmores/lucette/core"
)

// EgressPerByte is the classic cloud figure, about $0.09 per GB. Traffic into
// the centre is free; traffic leaving it is priced.
//
// Cost() carries money per byte, which is the quantity the claim measures
// directly.
const EgressPerByte = 0.09 / (1 << 30)

// Link bandwidth is quoted in bits per second, matching core.Link's
// Bandwidthbps. Payload sizes are counted in bytes, so anything that turns a
// bandwidth into a transfer time has to divide by 8. There is deliberately no
// byte-per-second constant here, so that conversion cannot be applied twice by
// accident.
const (
	Mbps = 1_000_000
	Gbps = 1_000_000_000
)

// SelfLinkBandwidth is what a node's link to itself reports. A self-link is not a
// link -- there is nothing to move -- so this value is chosen to make the copy
// instant. Reporting zero would read as "infinitely slow" to any transfer-time
// calculation, which is the exact opposite of the truth.
const SelfLinkBandwidth = 1e12 // 1 Tbps

// Node IDs of the default star, as constants so that DefaultNodes and the link
// maps cannot drift apart. A link keyed by an ID missing from Nodes would never
// be read, and a node missing from a link map reads back as a nil interface
// whose methods panic on call.
const (
	CentreID       = core.NodeID("node1")
	CentreWorkerID = core.NodeID("node4")
	SiteAID        = core.NodeID("node2")
	SiteBID        = core.NodeID("node3")

	centreNet = "centre-net"
	siteANet  = "site-a-net"
	siteBNet  = "site-b-net"
)

// Node resource capacity, in the same unit WorkUnit.ResourceUsage is counted in
// -- read as GB of VRAM available for session state, after the model's weights
// have taken their share.
//
// The numbers are sized against the load, not in the abstract. DefaultWorkUnits
// produces ten units that each stay legal for DefaultWorkUnitBudget clicks and
// arrive ten clicks apart, so at most four are alive at once: a concurrent
// working set of roughly 4 x DefaultWorkUnitResource = 16. Three of these four
// nodes hold less than that, so concentrating work on any of them fills it and
// forces an eviction.
//
// The centre alone can absorb the whole default load. That is deliberate: it
// stands in for cloud elasticity.
const (
	CentreResource       = 32.0
	CentreWorkerResource = 16.0
	SiteAResource        = 12.0
	SiteBResource        = 8.0
)

// Node price, in dollars per second of occupancy -- Node.Cost.
//
// This is what makes the centre non-trivial. It is roomy, free to send into
// (ingress is unmetered) and instant to reach itself, so it never waits and never
// pays for a transfer -- but it charges by the second, while the two sites are
// machines that are already bought. The trade the claim is about lives exactly
// there: keep a session on the free site that already holds its state, or move it
// to the expensive roomy one.
//
// The centre's worker sits between the two: rented like the centre, inside the
// centre's network, but smaller.
const (
	CentreCost       = 0.0005 // about $1.80 per GPU-hour
	CentreWorkerCost = 0.0002
	SiteACost        = 0 // owned; the marginal cost is electricity
	SiteBCost        = 0
)

// ── Nodes ──────────────────────────────────────────────────────────

// emptyNode is a node with nothing placed on it yet: free resource equals total.
// Built in one place so the two cannot start out disagreeing.
func emptyNode(id core.NodeID, totalResource float64) core.Node {
	return core.Node{
		ID:            id,
		TotalResource: totalResource,
		FreeResource:  totalResource,
	}
}

// NOTE: DefaultNodes use only four for test
//
// Cost is assigned separately from the resources rather than folded into
// emptyNode. emptyNode's job is keeping Free equal to Total, and a price has
// nothing to do with how full a node is; folding it in would also put two
// adjacent float64s in one signature -- a capacity in the tens and a price in the
// ten-thousandths -- which a call site can swap without failing.
func DefaultNodes() map[core.NodeID]*core.Node {
	centre := emptyNode(CentreID, CentreResource)
	centre.Cost = CentreCost

	worker := emptyNode(CentreWorkerID, CentreWorkerResource)
	worker.Cost = CentreWorkerCost

	siteA := emptyNode(SiteAID, SiteAResource)
	siteA.Cost = SiteACost

	siteB := emptyNode(SiteBID, SiteBResource)
	siteB.Cost = SiteBCost

	return map[core.NodeID]*core.Node{
		centre.ID: &centre,
		worker.ID: &worker,
		siteA.ID:  &siteA,
		siteB.ID:  &siteB,
	}
}

// ── Policy ──────────────────────────────────────────────────────────

// netPolicy is a core.Policy that is nothing but a name. GetHash digests the
// name rather than returning a hand-written byte string, so two links carrying
// the same policy always agree on its hash and the value stays reproducible.
type netPolicy string

func (p netPolicy) GetID() string { return string(p) }

func (p netPolicy) GetHash() []byte {
	sum := sha256.Sum256([]byte(p))
	return sum[:]
}

// ── Link ───────────────────────────────────────────────────────────

// Ecost is the concrete core.Link of the default star: a network cost plus a
// network identity, for test only.
type Ecost struct {
	cost         float64 // money per byte
	latencyMs    float64
	bandwidthBps float64 // bits per second
	policy       core.Policy
	network      string
}

func (e *Ecost) Cost() float64         { return e.cost }
func (e *Ecost) Latencyms() float64    { return e.latencyMs }
func (e *Ecost) Bandwidthbps() float64 { return e.bandwidthBps }

func (e *Ecost) Policy() string     { return e.policy.GetID() }
func (e *Ecost) PolicyHash() []byte { return e.policy.GetHash() }

func (e *Ecost) Network() string { return e.network }

func (e *Ecost) NetworkHash() []byte {
	sum := sha256.Sum256([]byte(e.network))
	return sum[:]
}

// linkSpec is the argument list for link. Its fields are named rather than
// positional because cost, latency and bandwidth are three adjacent float64s,
// which are trivially swappable at a call site and would fail silently.
type linkSpec struct {
	cost         float64
	latencyMs    float64
	bandwidthBps float64
	policy       string
	network      string
}

func link(s linkSpec) *Ecost {
	return &Ecost{
		cost:         s.cost,
		latencyMs:    s.latencyMs,
		bandwidthBps: s.bandwidthBps,
		policy:       netPolicy(s.policy),
		network:      s.network,
	}
}

// ── StarMap ──────────────────────────────────────────────────────────

// DefaultUpLinks is what it costs a node to send toward the centre.
//
// The centre carries an entry for itself rather than none: UpLink is a map of
// interfaces, so a missing key reads back as a nil interface and calling a
// method on it panics. A total map means no caller has to remember to check.
//
// The centre's worker sits inside the centre's own network, so its traffic never
// crosses a network boundary: no money, 0.2 ms, 10 Gbps. Its policy is named
// "intra-net" rather than "local" on purpose -- the centre's own entry means
// "same machine, nothing to transfer", while this one means "a real transfer
// between two machines, just a cheap one". Both cost zero, so the name is what
// states which situation applies.
//
// A site's uplink is its own number and not its downlink's: home-style sites have
// a slow uplink and a fast downlink, and that asymmetry is most of the reason a
// live session is expensive to move.
func DefaultUpLinks() map[core.NodeID]core.Link {
	return map[core.NodeID]core.Link{
		CentreID: link(linkSpec{
			latencyMs: 0, bandwidthBps: SelfLinkBandwidth,
			policy: "local", network: centreNet,
		}),
		CentreWorkerID: link(linkSpec{
			latencyMs: 0.2, bandwidthBps: 10 * Gbps,
			policy: "intra-net", network: centreNet,
		}),
		SiteAID: link(linkSpec{
			latencyMs: 40, bandwidthBps: 100 * Mbps, // a consumer uplink is the bottleneck
			policy: "unmetered-up", network: siteANet,
		}),
		SiteBID: link(linkSpec{
			latencyMs: 30, bandwidthBps: 500 * Mbps,
			policy: "unmetered-up", network: siteBNet,
		}),
	}
}

// DefaultDownLinks is what it costs the centre to send toward a node.
//
// A separate map from DefaultUpLinks, because the two directions are different
// objects rather than one object read twice: entering the centre is free while
// leaving it is priced. The centre's worker is free both ways for the same
// reason -- nothing leaves the network, so nothing is charged.
//
// The two remote sites differ on every axis the fixture can express -- latency,
// bandwidth, and price -- so that a policy has something to discriminate on. With
// all three equal they would be interchangeable and no placement decision could
// tell them apart. The values themselves are placeholders.
func DefaultDownLinks() map[core.NodeID]core.Link {
	return map[core.NodeID]core.Link{
		CentreID: link(linkSpec{
			latencyMs: 0, bandwidthBps: SelfLinkBandwidth,
			policy: "local", network: centreNet,
		}),
		CentreWorkerID: link(linkSpec{
			latencyMs: 0.2, bandwidthBps: 10 * Gbps,
			policy: "intra-net", network: centreNet,
		}),
		SiteAID: link(linkSpec{
			cost:      EgressPerByte,
			latencyMs: 40, bandwidthBps: 500 * Mbps,
			policy: "metered-egress", network: siteANet,
		}),
		SiteBID: link(linkSpec{
			cost:      EgressPerByte / 4,
			latencyMs: 30, bandwidthBps: 500 * Mbps,
			policy: "metered-egress", network: siteBNet,
		}),
	}
}

func DefaultStarMap() *core.StarMap {
	return &core.StarMap{
		ID:       "default_starmap",
		Center:   CentreID,
		Nodes:    DefaultNodes(),
		UpLink:   DefaultUpLinks(),
		DownLink: DefaultDownLinks(),
	}
}
