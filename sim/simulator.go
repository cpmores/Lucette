package sim

import (
	"log"
	"sync"

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
	WuID        *SimWorkUnitID
	Decisions   chan *core.Decision
}

func (s *NodeSimulator) GetID() SimulatorID {
	return s.ID
}

// Methods
func InitNodeSimulator(id SimulatorID, placePolicy core.PlacePolicy, starMap *core.StarMap) *NodeSimulator {
	s := &NodeSimulator{
		ID:          SimulatorID(id),
		PlacePolicy: placePolicy,
		StarMap:     starMap,
		WorkUnits:   make(map[core.WorkUnitID]*core.WorkUnit),
		WuID: &SimWorkUnitID{
			ID: 0,
		},
		Decisions: make(chan *core.Decision, 100),
	}

	go s.Reconcile()

	return s
}

type SimWorkUnitID struct {
	ID core.WorkUnitID
	sync.Mutex
}

func (w *SimWorkUnitID) NewID() core.WorkUnitID {
	w.Lock()
	defer w.Unlock()
	w.ID++
	return w.ID
}

// Time Pass
func (s *NodeSimulator) Pass(click int64) {
	if click < s.Click {
		log.Fatalf("[%s] Cannot pass to a past click: current click %d, requested click %d", s.GetID(), s.Click, click)
	}

	currentTime := click
	// scan WorkUnits, place or clear WUs
	for index, wu := range s.WorkUnits {
		status := wu.Status
		// check if the workunit is expired
		deadline := wu.Deadline
		if click > deadline {
			wu.Status = core.Expired
		}

		if status == core.Waiting {
			// FINISHED: finish node and placepolicy
			decision := core.Place(wu, s.StarMap, s.PlacePolicy, currentTime)
			s.Decisions <- &decision
			log.Printf("[%s] WorkUnit %s: status %d, decision %d", s.GetID(), index, status, decision.Outcome)
		} else if status == core.Cancelled || status == core.OutOfDate {
			// clear the workunit from the node
			delete(s.WorkUnits, index)
			log.Printf("[%s] WorkUnit %s: status %d, cleared", s.GetID(), index, status)
		} else if status == core.Expired {
			// reinject the workunit to the node
			wu.Status = core.Waiting
			wu.Deadline = currentTime + 1000 // TODO: set a new deadline
			s.WorkUnits[index] = wu
			log.Printf("[%s] WorkUnit %s: status %d, reinjected", s.GetID(), index, status)
		}
	}

	// reconcile, settle placeable workunits and upadte status
	s.Reconcile()

	// update click
	s.Click = currentTime
}

// Inject a new WorkUnit into this Node cluster
func (s *NodeSimulator) Inject(wu *core.WorkUnit) {
	// FINISHED: workunit injection
	wu.ID = s.WuID.NewID()
	s.WorkUnits[wu.ID] = wu
	log.Printf("[%s] Injected WorkUnit %d: status %d, deadline %d", s.GetID(), wu.ID, wu.Status, wu.Deadline)
}

// Reconcile workunit status with the real world
func (s *NodeSimulator) Reconcile() {
	for {
		select {
		case decision := <-s.Decisions:
			wuID := decision.WuID
			outcome := decision.Outcome
		}
	}
}

func (s *NodeSimulator) settle(decision *core.Decision) {
	// Perform any necessary actions to settle the workunit on the specified node
	if decision.Outcome != core.Placed {
		log.Printf("[%s] Cannot settle WorkUnit %d: decision outcome is not Placed", s.GetID(), decision.WuID)
		return
	}

	// If placed, check workunit status and update accordingly
	// TODO: implement settle logic, from InPlan to Working
}

// Decide what to do with the decision
func (s *NodeSimulator) decide(decision *core.Decision) {
}
