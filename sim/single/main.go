// Package main provides a single placer reconcile simulation
// it uses DefaultWorkUints and DefaultTimeline to simulate a single placer reconcile scenario
package main

import (
	"slices"

	core "github.com/cpmores/lucette/core"
	sim "github.com/cpmores/lucette/sim"
)

type Placer struct {
	StarMap core.StarMap
}

func main() {
	starMap := sim.DefaultStarMap()
	timeline := sim.DefaultTimeline()
	center := starMap.Center
	timeClicks := make([]int64, 0)
	for timeClick := range timeline.Events {
		timeClicks = append(timeClicks, timeClick)
	}
	slices.Sort(timeClicks)

	// reconcile the timeline events in order of time clicks
	for timeClick := range timeClicks {
	}
}
