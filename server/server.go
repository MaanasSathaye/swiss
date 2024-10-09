package server

import (
	"net"
	"sync"

	"github.com/MaanasSathaye/swiss/stats"
)

type Server struct {
	Alive      bool
	StatusChan chan struct{}
	db         stats.SSRetriever
	stats      stats.ServerConfig
	listener   net.Listener
	mutex      sync.Mutex
}
