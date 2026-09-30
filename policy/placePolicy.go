package policy

import (
	core "github.com/cpmores/lucette/core"
)

// type Policy interface {
// 	GetID() string
// 	GetHash() []byte
// }
//
// type PlacePolicy interface {
// 	Policy
// 	Score(wu WorkUnit, sm StarMap, n NodeID, now int64) (score uint64, feasible bool)
// }

type DefaultPlacePolicy struct{}

// ── Policy Interface ──────────────────────────────────────────────────────────

func (p *DefaultPlacePolicy) GetID() string {
	return "default"
}

func (p *DefaultPlacePolicy) GetHash() []byte {
	return []byte("default")
}

// ── PlacePolicy Interface ──────────────────────────────────────────────────────────

// WARN: only for test, similiar to VRAM in placement policy
func (p *DefaultPlacePolicy) Score(wu *core.WorkUnit,
	sm *core.StarMap, n *core.Node, now int64) (score float64, feasible bool) {
	// check if the node has enough resource
	if n.FreeResource < wu.ResourceUsage {
		return 0, false
	}

	nodeID := n.ID
	// check if the node is reachable
	link, ok := sm.UpLink[nodeID]
	if !ok {
		return 0, false
	}

	// WARN: bad score function, only for test
	s := 100.0 * (link.Bandwidthbps()) / (link.Cost() + link.Latencyms() + n.Cost)
	return s, true
}
