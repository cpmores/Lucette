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
	ID        SimulatorID
	Click     int64
	StarMap   *core.StarMap
	WorkUnits map[core.WorkUnitID]*core.WorkUnit
}

func (s *NodeSimulator) GetID() SimulatorID {
	return s.ID
}

// Methods
func InitNodeSimulator(id SimulatorID, starMap *core.StarMap) *NodeSimulator {
	return &NodeSimulator{
		ID:      SimulatorID(id),
		StarMap: starMap,
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

		decision := nil
		if status == core.Waiting {
			// TODO: finish node and placepolicy
			decision = core.Place(wu, *s.StarMap)
		}
	}

	// update StarMap

	// update click
}

// Place a new WokrUnit into this Node cluster
func (s *NodeSimulator) Place(wu *core.WorkUnit) {
}
