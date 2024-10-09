package server

type Server struct {
	Alive      bool
	StatusChan chan struct{}
}
