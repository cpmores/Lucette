package sim

import (
	"log"

	core "github.com/cpmores/lucette/core"
)

// ── NodeSimulator ──────────────────────────────────────────────────────────

// Structure
type SimulatorID string

type Simulator interface {
	GetID() SimulatorID
}

type NodeSimulator struct {
	ID          SimulatorID
	Click       int64
	StarMap     *core.StarMap
	PlacePolicy core.PlacePolicy
	WorkUnits   map[core.WorkUnitID]*core.WorkUnit
	Decisions   chan *core.Decision
}

func (s *NodeSimulator) GetID() SimulatorID {
	return s.ID
}

// Methods
func InitNodeSimulator(id SimulatorID, placePolicy core.PlacePolicy, starMap *core.StarMap) *NodeSimulator {
	return &NodeSimulator{
		ID:          SimulatorID(id),
		PlacePolicy: placePolicy,
		StarMap:     starMap,
		WorkUnits:   make(map[core.WorkUnitID]*core.WorkUnit),
		Decisions:   make(chan *core.Decision, 100),
	}
}

// Time Pass
func (s *NodeSimulator) Pass(click int64) {
	if click < s.Click {
		log.Fatalf("[%s] Cannot pass to a past click: current click %d, requested click %d", s.GetID(), s.Click, click)
	}

	currentTime := click
	// scan WorkUnits, check if it is needed to update StarMap
	for index, wu := range s.WorkUnits {
		status := wu.Status
		// check if the workunit is expired
		deadline := wu.Deadline
		if click > deadline {
			wu.Status = core.Expired
		}

		decision := core.Decision{
			Outcome: core.Unknown,
		} // TODO: default decision
		if status == core.Waiting {
			// FINISHED: finish node and placepolicy
			decision := core.Place(wu, s.StarMap, s.PlacePolicy, currentTime)
		}

		log.Printf("[%s] WorkUnit %s: status %d, decision %d", s.GetID(), index, status, decision.Outcome)
	}

	// update StarMap

	// update click
	s.Click = currentTime
}

// Inject a new WorkUnit into this Node cluster
func (s *NodeSimulator) Inject(wu *core.WorkUnit) {
	// TODO: workunit injection
}

// Reconcile workunit status with the real world
func (s *NodeSimulator) Reconcile() {
}

// Decide what to do with the decision
func (s *NodeSimulator) decide(decision *core.Decision) {
}
