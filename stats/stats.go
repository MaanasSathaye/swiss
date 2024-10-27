package stats

import (
	"fmt"
	"time"
)

// https://github.com/ThePrimeagen/vim-arcade/blob/8c99725866086cf2797973db37d65573c9674d39/pkg/game-server-stats/stats.go
type ConnectionStats struct {
	Connections        int
	ConnectionsAdded   int
	ConnectionsRemoved int
	Load               float32
	UpdatedAt          time.Time
}

func (cs *ConnectionStats) String() string {
	return fmt.Sprintf("Connectionss=%d Added=%d Removed=%d", cs.Connections, cs.ConnectionsAdded, cs.ConnectionsRemoved)
}

func (cs *ConnectionStats) Reset() {
	cs.Connections = 0
	cs.ConnectionsAdded = 0
	cs.ConnectionsRemoved = 0
}

func (cs *ConnectionStats) Equal(rhs *ConnectionStats) bool {
	return cs.Connections == rhs.Connections &&
		cs.ConnectionsRemoved == rhs.ConnectionsRemoved &&
		cs.ConnectionsAdded == rhs.ConnectionsAdded
}
