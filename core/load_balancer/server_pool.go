package load_balancer

import (
	"encoding/json"
	"os"
	"sync"
	"fmt"
	"reverse_proxy/CustomErrors"
)

// Server Pool struct implements LoadBalancer
type ServerPool struct {
	sync.RWMutex
	Backends []*Backend `json:"backends"`
	Current  uint64     `json:"current"` // Used for Round-Robin
}

// Constructor for the struct with string parameter of the path of the configuration file
func NewServerPool(conf_file_name string) (*ServerPool, error) {
	// Instantiating a new server pool
	var sp = ServerPool{}

	// Reading the configuration file into a byte array
	conf_file, err := os.ReadFile(conf_file_name)
	// File reading error handling
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", customerrors.ServerPoolUnmarshalErr, err)
	}

	// Instantiating a slice of backends
	var backends []Backend

	// Unmarshaling the file byte array
	err = json.Unmarshal(conf_file, &backends)
	// Json unmarshaling error handling
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", customerrors.ServerPoolUnmarshalErr, err)
	}
	
	// Initializing a slice of pointers of backends and copying the address of the unmarshaled backends
	sp.Backends = make([]*Backend, len(backends))
	for i := range backends {
		sp.Backends[i] = &backends[i]
	}
	
	return &sp, nil
}

// Getters
func (sp *ServerPool) GetBackendsNum() int {
	return len(sp.Backends)
}

func (sp *ServerPool) GetAliveBackendsNum() int {
	if len(sp.Backends) <= 0 {
		return 0
	}

	// Incrementing num if a backend is alive
	num := 0
	for _, b := range sp.Backends {
		if b.Alive {
			num++
		}
	}

	return num
}

func (sp *ServerPool) GetBackend(i int) *Backend {
	return sp.Backends[i]
}

// Method to find backend by url value
func (sp *ServerPool) FindBackendByURL(url string) *Backend {
	for i := range sp.Backends {
		if sp.Backends[i].URL.String() == url {
			return sp.Backends[i]
		}
	}
	return nil
}

// Method to Thread-safe increment the counter current
func (sp *ServerPool) increment_current() {
	sp.Lock()
	defer sp.Unlock()
	sp.Current = (sp.Current + 1) % uint64(len(sp.Backends))
}

// Method to verify if all backends are not alive
func (sp *ServerPool) all_not_alive() bool {
	if len(sp.Backends) <= 0 {
		return true
	}

	// If at least one is alive return false
	for _, b := range sp.Backends {
		if b.Alive {
			return false
		}
	}

	return true
}

// Method to get the next valid peer using the round-robin strategy
func (sp *ServerPool) GetNextValidPeer() *Backend {
	// Check if there is at least one backend alive
	if sp.all_not_alive() {
		return nil
	}

	// Increment the counter current until finding the next alive backend
	sp.increment_current()
	for !sp.Backends[sp.Current].Alive {
		sp.increment_current()
	}

	return sp.Backends[sp.Current]
}

// Method to get the next valid peer using the least-connections strategy
func (sp *ServerPool) LeastConnValidPeer() *Backend {
	// Check if there is at least one backend alive
	if sp.all_not_alive() {
		return nil
	}

	// Getting the first backend that is alive
	i := 0
	var least_conn_peer *Backend
	for !sp.Backends[i].Alive {
		i++
	}
	least_conn_peer = sp.GetBackend(i)

	// Finding the backend with minimal connections and at the same time alive
	for _, b := range sp.Backends {
		if (b.Alive) && (least_conn_peer.CurrentConns > b.CurrentConns) {
			least_conn_peer = b
		}
	}

	return least_conn_peer
}

// Thread-safe backend adding and removing
func (sp *ServerPool) AddBackend(backend *Backend) error {
	sp.Lock()
	defer sp.Unlock()
	sp.Backends = append(sp.Backends, backend)
	return nil
}

func (sp *ServerPool) RemoveBackend(backend *Backend) error {
	sp.Lock()
	defer sp.Unlock()
	targetURL := backend.URL.String()
	// backend removal is done without taking care of the order to assure fast deletion
	for i, b := range sp.Backends {
		if b.URL.String() == targetURL {
			// Swap with last
			lastIdx := len(sp.Backends) - 1
			sp.Backends[i] = sp.Backends[lastIdx]
			
			// Clean up for Garbage collector
			sp.Backends[lastIdx] = nil
			sp.Backends = sp.Backends[:lastIdx]
			
			return nil
		}
	}
	return customerrors.BackendNotFound
}