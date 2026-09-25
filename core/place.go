package core

const RETRY_INTERVAL = 3 // seconds

// Place decides where to place a workunit
// according to the starmap it receives.
//
// returns Decision with Outcome Placed, Deferred, or Rejected
func Place(wu WorkUnit, sm StarMap, p PlacePolicy, now int64) Decision {
	// NOTE: for now we just return one node that satified the policy,
	// but in the future we may return multiple nodes and let the agent choose one
	// like top-k nodes with highest score, or top-k nodes with lowest cost, etc.
	oneFeasible := false
	bestScore := uint64(0)
	bestNode := NodeID("")
	for nodeID, node := range sm.Nodes {
		score, feasible := p.Score(wu, sm, node, now)
		oneFeasible = oneFeasible || feasible
		if feasible {
			if bestScore < score {
				bestScore = score
				bestNode = nodeID
			} else if bestScore == score && nodeID < bestNode {
				bestNode = nodeID
			}
		}
	}

	// choose return Outcomes
	if oneFeasible && bestNode != "" {
		// if we have at least one feasible node, we return the best one
		return Decision{
			Outcome: Placed,
			Node:    bestNode,
			Score:   bestScore,
		}
	} else if now < wu.Deadline {
		// if no candidat, but still have time, we return Deferred
		retry := now + RETRY_INTERVAL
		retry = min(retry, wu.Deadline)

		return Decision{
			Outcome: Deferred,
			RetryAt: retry,
			Reason:  ReasonNoFeasibleNode,
			Detail:  "no feasible node found, will retry later",
		}
	}

	// everything is gone
	return Decision{
		Outcome: Rejected,
		Reason:  ReasonDeadlinePassed,
		Detail:  "deadline passed, no feasible node found",
	}
}

type Outcome uint8

const (
	Placed Outcome = iota
	Deferred
	Rejected
)

type Reason uint8

const (
	ReasonUnknown Reason = iota
	ReasonNodeNotVisible
	ReasonVRAMShort
	ReasonCostCeiling
	ReasonContextTooLong
	ReasonNoFeasibleNode
	ReasonDeadlinePassed
)

type Decision struct {
	Outcome Outcome

	// PLaced
	Node  NodeID
	Score uint64

	// Deferred: when to ask again
	RetryAt int64

	Reason Reason
	Detail string
}
