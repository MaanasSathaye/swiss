package stats

import (
	"context"
	"fmt"
	"time"
)

type ConnectionStats struct {
	Connections        int
	ConnectionsAdded   int
	ConnectionsRemoved int
}

func (cs *ConnectionStats) String() string {
	return fmt.Sprintf("Connectionss=%d Added=%d Removed=%d", cs.Connections, cs.ConnectionsAdded, cs.ConnectionsRemoved)
}

func (cs *ConnectionStats) Equal(rhs *ConnectionStats) bool {
	return cs.Connections == rhs.Connections &&
		cs.ConnectionsRemoved == rhs.ConnectionsRemoved &&
		cs.ConnectionsAdded == rhs.ConnectionsAdded
}

type ServerConfig struct {
	State              int
	Id                 string
	Connections        int
	ConnectionsAdded   int
	ConnectionsRemoved int
	UpdatedAt          time.Time
	Load               float32
	Host               string
	Port               int
}

func (sc *ServerConfig) Equal(rhs *ServerConfig) bool {
	return sc.Id == rhs.Id &&
		sc.Connections == rhs.Connections &&
		sc.ConnectionsAdded == rhs.ConnectionsAdded &&
		sc.ConnectionsRemoved == rhs.ConnectionsRemoved
}

func (sc *ServerConfig) String() string {
	return fmt.Sprintf("Server(%s): Addr=%s Conns=%d Load=%f", sc.Id, sc.Addr(), sc.Connections, sc.Load)
}

func (sc *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", sc.Host, sc.Port)
}

type ServerFuncs interface {
	GetId(string) *ServerConfig
	GetAllServerConfigs() ([]ServerConfig, error)
	Run(ctx context.Context)
	GetServersByUtilization(maxLoad float64) []ServerConfig
	Update(stats ServerConfig) error
	GetServerCount() int
	GetTotalConnectionCount() ConnectionStats
}
