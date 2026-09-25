package sim

import (
	"testing"

	"github.com/cpmores/lucette/core"
)

// TestDefaultStarMapIsTotal: every node in Nodes must appear in both link maps.
// A missing key reads back as a nil interface whose methods panic when called,
// so the map being total is what keeps lookups safe.
func TestDefaultStarMapIsTotal(t *testing.T) {
	sm := DefaultStarMap()

	if _, ok := sm.Nodes[sm.Center]; !ok {
		t.Fatalf("Center %q is not in Nodes", sm.Center)
	}

	for id := range sm.Nodes {
		up, ok := sm.UpLink[id]
		if !ok {
			t.Errorf("no uplink for %s: a lookup would panic, not fail loudly", id)
			continue
		}
		down, ok := sm.DownLink[id]
		if !ok {
			t.Errorf("no downlink for %s", id)
			continue
		}
		if up == nil || down == nil {
			t.Errorf("%s: nil link in the map", id)
		}
	}
}

// TestDefaultStarMapLinksAreUsable calls every accessor. A nil policy inside
// Ecost would panic only when Policy() is reached, and nothing else notices.
//
// It also pins that bandwidth is positive. A link reporting zero bandwidth is
// read as "infinitely slow" by any transfer-time calculation, and one reporting
// a negative value turns time negative -- both are silent wrong answers rather
// than errors.
func TestDefaultStarMapLinksAreUsable(t *testing.T) {
	sm := DefaultStarMap()

	for id := range sm.Nodes {
		for _, dir := range []struct {
			name string
			l    core.Link
		}{
			{"up", sm.UpLink[id]},
			{"down", sm.DownLink[id]},
		} {
			if dir.l == nil {
				t.Errorf("%s %s: nil link", id, dir.name)
				continue
			}
			if dir.l.Policy() == "" {
				t.Errorf("%s %s: empty policy name", id, dir.name)
			}
			if len(dir.l.PolicyHash()) == 0 {
				t.Errorf("%s %s: empty policy hash", id, dir.name)
			}
			if dir.l.Network() == "" {
				t.Errorf("%s %s: empty network name", id, dir.name)
			}
			if len(dir.l.NetworkHash()) == 0 {
				t.Errorf("%s %s: empty network hash", id, dir.name)
			}
			if dir.l.Cost() < 0 {
				t.Errorf("%s %s: negative cost %v", id, dir.name, dir.l.Cost())
			}
			if dir.l.Latencyms() < 0 {
				t.Errorf("%s %s: negative latency %v", id, dir.name, dir.l.Latencyms())
			}
			if dir.l.Bandwidthbps() <= 0 {
				t.Errorf("%s %s: bandwidth %v, want positive or the link is infinitely slow",
					id, dir.name, dir.l.Bandwidthbps())
			}
		}
	}
}

// TestDefaultStarMapCentreWorkerIsLocal pins what makes the worker inside the
// centre's own network different: nothing crosses a network boundary, so it is
// free in both directions, it sits on the centre's network, and it beats every
// remote site on both latency and bandwidth.
//
// That last part is the substance. A worker that were merely "another free link"
// would be indistinguishable from the centre's own entry, and the fixture would
// have added a node that no placement policy could reason about.
func TestDefaultStarMapCentreWorkerIsLocal(t *testing.T) {
	sm := DefaultStarMap()

	if _, ok := sm.Nodes[CentreWorkerID]; !ok {
		t.Fatalf("%s is not in Nodes, so Place would never consider it", CentreWorkerID)
	}

	workerUp := sm.UpLink[CentreWorkerID]
	workerDown := sm.DownLink[CentreWorkerID]

	for _, dir := range []struct {
		name string
		l    core.Link
	}{
		{"up", workerUp},
		{"down", workerDown},
	} {
		if dir.l == nil {
			t.Fatalf("%s: nil link", dir.name)
		}
		if got := dir.l.Cost(); got != 0 {
			t.Errorf("%s: traffic inside the centre's network should cost nothing, got %v", dir.name, got)
		}
		if got, want := dir.l.Network(), sm.UpLink[CentreID].Network(); got != want {
			t.Errorf("%s: on network %q, want the centre's network %q", dir.name, got, want)
		}
		if got := dir.l.Policy(); got == sm.UpLink[CentreID].Policy() {
			t.Errorf("%s: policy %q is the same as the centre's own entry, which means two different situations share one name",
				dir.name, got)
		}
	}

	for _, remote := range []core.NodeID{SiteAID, SiteBID} {
		remoteUp := sm.UpLink[remote]

		if workerUp.Latencyms() >= remoteUp.Latencyms() {
			t.Errorf("worker latency %v is not below %s's %v; being in the centre's network should be closer",
				workerUp.Latencyms(), remote, remoteUp.Latencyms())
		}
		if workerUp.Bandwidthbps() <= remoteUp.Bandwidthbps() {
			t.Errorf("worker bandwidth %v is not above %s's %v",
				workerUp.Bandwidthbps(), remote, remoteUp.Bandwidthbps())
		}

		// Moving state to the worker must not cost money, while moving it out to
		// a remote site must.
		if workerDown.Cost() >= sm.DownLink[remote].Cost() {
			t.Errorf("moving state to the worker costs %v, not less than %s's %v",
				workerDown.Cost(), remote, sm.DownLink[remote].Cost())
		}
	}
}

// TestDefaultStarMapDirectionsDiffer pins the asymmetry the fixture exists for:
// leaving the centre is priced, entering it is not, and a site's uplink is not
// its downlink.
func TestDefaultStarMapDirectionsDiffer(t *testing.T) {
	sm := DefaultStarMap()

	for _, id := range []core.NodeID{SiteAID, SiteBID} {
		if got := sm.UpLink[id].Cost(); got != 0 {
			t.Errorf("%s: entering the centre should be free, got %v", id, got)
		}
		if got := sm.DownLink[id].Cost(); got <= 0 {
			t.Errorf("%s: leaving the centre should be priced, got %v", id, got)
		}
	}

	// A home-style site uploads slowly and downloads fast. If these ever match,
	// the fixture has quietly become a symmetric link and the reason a live
	// session is expensive to move has been modelled away.
	if up, down := sm.UpLink[SiteAID].Bandwidthbps(), sm.DownLink[SiteAID].Bandwidthbps(); up >= down {
		t.Errorf("%s: uplink %v is not slower than downlink %v", SiteAID, up, down)
	}

	if got := sm.UpLink[CentreID].Cost(); got != 0 {
		t.Errorf("the centre should reach itself for free, got %v", got)
	}
	if got := sm.DownLink[CentreID].Cost(); got != 0 {
		t.Errorf("the centre should reach itself for free, got %v", got)
	}
}
