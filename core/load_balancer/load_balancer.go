package load_balancer

// Load Balancer interface
type LoadBalancer interface {
	GetBackendsNum() int
	GetAliveBackendsNum() int
	GetBackend(i int) *Backend
	GetNextValidPeer() *Backend
	LeastConnValidPeer() *Backend
	AddBackend(backend *Backend) error
	RemoveBackend(backend *Backend) error
}