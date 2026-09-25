// Package core introduce WorkUnit, StarMap
package core

// ── Event ──────────────────────────────────────────────────────────

// Event presents when and where the event happened
type Event struct {
	ID       EventID
	Click    int64
	Category string
	Where    string
	Instance any
}

// ── WorkUnit ──────────────────────────────────────────────────────────

// WorkUnit Status

type WorkUnitStatus uint8

const (
	Waiting   WorkUnitStatus = iota // waiting for placed
	InPlan                          // after placed, waiting for settled
	Working                         // settled, successfully working on the workunit
	Cancelled                       // cancelled by user before settled
	OutOfDate                       // cancelled by user, but already settled
	Expired                         // Deadline reached
)

// WorkUnit contains information needed for a single work
type WorkUnit struct {
	// TODO: finish agent message for workunit
	// only implement ID for now
	ID     WorkUnitID
	Status WorkUnitStatus

	CreatedAt int64 // when wu is created
	Deadline  int64 // when wu is illegal

	// WARN: only for test
	ResourceUsage float64 // only for test, similiar to VRAM in placement policy
	SettleTime    int64   // only for test, when wu is settled
	TimeUsage     int64   // omly for test
}

// StarMap contains information needed for a network
type StarMap struct {
	// TODO: finish agent message for starmap
	// only implement ID for now
	// NOTE: every node has its unique StarMap for its unique center
	ID       StarMapID
	Center   NodeID
	Nodes    map[NodeID]*Node
	UpLink   map[NodeID]Link // from this node to center
	DownLink map[NodeID]Link // from center to this node
}

type Link interface {
	// cost
	Cost() float64      // TODO: change the return from float64 to something specific
	Latencyms() float64 // TODO: change the return from float64 to something specific
	Bandwidthbps() float64

	// idenity
	Policy() string      // return name of the Policy
	PolicyHash() []byte  // return hash of the Policy
	Network() string     // return name of the Network
	NetworkHash() []byte // return hash of the Network
}

type Policy interface {
	GetID() string
	GetHash() []byte
}

type PlacePolicy interface {
	Policy
	Score(wu WorkUnit, sm StarMap, n *Node, now int64) (score uint64, feasible bool)
}
