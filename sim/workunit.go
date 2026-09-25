package sim

import (
	"fmt"

	"github.com/cpmores/lucette/core"
)

// The default fixtures live on the logical clock, which is in seconds: that is
// what core.RETRY_INTERVAL's trailing comment says, and it is the only place the
// unit is written down. WorkUnit.CreatedAt and WorkUnit.Deadline are plain int64s
// carrying no unit of their own, so every value below is expressed in seconds to
// match Place's now.
const (
	// DefaultWorkUnitBudget is how long a work unit stays legal. It is chosen so
	// that a caller passing now = 0 leaves room for several RETRY_INTERVAL
	// retries before Place stops deferring and rejects.
	DefaultWorkUnitBudget = int64(30) // seconds

	// DefaultWorkUnitID is the ID of the first work unit, so that
	// DefaultWorkUnit and DefaultWorkUnits()[0] describe the same piece of work.
	DefaultWorkUnitID = core.WorkUnitID("wu-1")

	// DefaultWorkUnitResource is what one default work unit occupies, in whatever
	// unit the NodeSimulator accounts in -- read it as GB of VRAM.
	//
	// The absolute value means nothing on its own: only its ratio to a node's
	// capacity matters, and node capacity does not exist yet. It is uniform across
	// the defaults so the sequence stays predictable, and a varying footprint is a
	// sweep axis rather than something baked into the fixture.
	DefaultWorkUnitResource = 4.0

	// DefaultWorkUnitTime is how long one default work unit holds its slot, in the
	// seconds the logical clock uses. It has to stay below DefaultWorkUnitBudget,
	// or a work unit would expire before it finished and the
	// release-on-completion path would never be reached.
	DefaultWorkUnitTime = int64(1)
)

// Timeline is a logical clock: the events that happen, grouped by the click they
// happen at, plus the click the scan stops at.
//
// Events sharing a click are kept in insertion order, so a caller that builds the
// timeline deterministically also traverses each click deterministically.
//
// NOTE: iterating Events itself is still unordered -- Go randomises map iteration
// -- so anything walking the timeline has to sort the clicks first, or two runs
// of the same simulation visit the clicks in different orders.
type Timeline struct {
	Events map[int64][]core.Event // timeline click -> events happened at that click

	// ScanUntil is the last click the scan covers, inclusive.
	//
	// It cannot be derived from Events. The last event is the last arrival, and a
	// work unit that arrives last is still alive after it: a scan stopping there
	// would end the run with work placed but never observed to finish, or waiting
	// and never rejected. See DefaultTimeline for how it is computed.
	ScanUntil int64
}

// DefaultWorkUnit is a single work unit, for exercising Place on its own:
// created at the origin, with a budget long enough that now = 0 is neither
// expired nor immediately rejected.
func DefaultWorkUnit() *core.WorkUnit {
	return &core.WorkUnit{
		ID:        DefaultWorkUnitID,
		CreatedAt: 0,
		Deadline:  DefaultWorkUnitBudget,

		ResourceUsage: DefaultWorkUnitResource,
		TimeUsage:     DefaultWorkUnitTime,
	}
}

// DefaultWorkUnits is a small sequence of work units sharing one budget and one
// footprint, arriving at staggered clicks.
//
// The staggering is the point. Work units that all arrived at the same click
// would present a placement policy with a single decision rather than a sequence
// of them, and nothing would test what happens when later work competes with the
// state earlier work left behind.
//
// The footprint is set here too, not left at zero: a work unit that occupies
// nothing would let every node admit everything, so eviction would never happen
// and the resource accounting would be exercised nowhere. A zero here fails
// silently rather than loudly.
func DefaultWorkUnits(count int, stride int64) []*core.WorkUnit {
	wus := make([]*core.WorkUnit, 0, count)
	for i := range count {
		createdAt := int64(i) * stride
		wus = append(wus, &core.WorkUnit{
			ID:        core.WorkUnitID(fmt.Sprintf("wu-%d", i+1)),
			CreatedAt: createdAt,
			Deadline:  createdAt + DefaultWorkUnitBudget,

			ResourceUsage: DefaultWorkUnitResource,
			TimeUsage:     DefaultWorkUnitTime,
		})
	}
	return wus
}

func DefaultTimeline() *Timeline {
	events := make(map[int64][]core.Event)
	defaultWUs := DefaultWorkUnits(10, 10)

	// The scan has to outlive the last work unit, not the last event: the last
	// event is the last arrival, and that unit is still alive afterwards.
	//
	// A unit can be placed no later than the click before its deadline -- at the
	// deadline itself Place rejects rather than defers -- and once placed it holds
	// its slot for TimeUsage. Carrying a full TimeUsage past the deadline is a safe
	// upper bound on that, so every unit is either finished or rejected by the
	// time the scan stops.
	scanUntil := int64(0)

	for index, wu := range defaultWUs {
		createdAtClick := wu.CreatedAt
		createdEvent := core.Event{
			ID:       core.EventID(fmt.Sprintf("created-%d", index)),
			Click:    createdAtClick,
			Category: "workunit-created",
			Where:    "sim",
			Instance: wu,
		}

		// insert events into the timeline
		if _, ok := events[createdAtClick]; !ok {
			events[createdAtClick] = make([]core.Event, 0)
		}
		events[createdAtClick] = append(events[createdAtClick], createdEvent)

		if end := wu.Deadline + wu.TimeUsage; end > scanUntil {
			scanUntil = end
		}
	}

	return &Timeline{
		Events:    events,
		ScanUntil: scanUntil,
	}
}
