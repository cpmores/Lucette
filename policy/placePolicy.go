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

// WARN: only for test, similiar to VRAM in placement policy
func (p *DefaultPlacePolicy) Score(wu core.WorkUnit,
	sm core.StarMap, n *core.Node, now int64) (score uint64, feasible bool) {
}
