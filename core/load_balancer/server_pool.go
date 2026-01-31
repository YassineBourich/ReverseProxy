package load_balancer

import (
	"encoding/json"
	"net/url"
	"os"
	"sync"
	"fmt"
	errors "reverse_proxy/CustomErrors"
)

type ServerPool struct {
	Backends []*Backend `json:"backends"`
	Current  uint64     `json:"current"` // Used for Round-Robin
	mux      sync.RWMutex
}

// Constructor for the struct with string parameter of the path of the configuration file
func NewServerPool(conf_file_name string) (*ServerPool, error) {
	// Instantiating a new server pool
	var sp = ServerPool{}

	// Reading the configuration file into a byte array
	conf_file, err := os.ReadFile(conf_file_name)
	// File reading error handling
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", errors.ServerPoolUnmarshalErr, err)
	}

	// Instantiating a slice of backends
	var backends []Backend

	// Unmarshaling the file byte array
	err = json.Unmarshal(conf_file, &backends)
	// Json unmarshaling error handling
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", errors.ServerPoolUnmarshalErr, err)
	}
	
	// Initializing a slice of pointers of backends and copying the address of the unmarshaled backends
	sp.Backends = make([]*Backend, len(backends))
	for i := range backends {
		sp.Backends[i] = &backends[i]
	}
	
	return &sp, nil
}

// Method to Thread-safe increment the counter current
func (sp *ServerPool) increment_current() {
	sp.mux.Lock()
	defer sp.mux.Unlock()
	sp.Current = (sp.Current + 1) % uint64(len(sp.Backends))
}

// Method to verify if all backends are not alive
func (sp *ServerPool) all_not_alive() bool {
	if len(sp.Backends) <= 0 {
		return true
	}

	for _, b := range sp.Backends {
		if b.Alive {
			return false
		}
	}

	return true
}

// Method to get the next valid peer using the round-robin strategy
func (sp *ServerPool) GetNextValidPeer() *Backend {
	sp.mux.Lock()
	defer sp.mux.Unlock()
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
	sp.mux.Lock()
	defer sp.mux.Unlock()
	// Check if there is at least one backend alive
	if sp.all_not_alive() {
		return nil
	}

	// Finding the backend with minimal connections and at the same time alive
	var least_conn_peer = sp.Backends[0]

	for _, b := range sp.Backends {
		if (b.Alive) && (least_conn_peer.CurrentConns > b.CurrentConns) {
			least_conn_peer = b
		}
	}

	return least_conn_peer
}

// Thread-safe backend appending
func (sp *ServerPool) AddBackend(backend *Backend) {
	sp.mux.Lock()
	defer sp.mux.Unlock()
	sp.Backends = append(sp.Backends, backend)
}

func (sp *ServerPool) SetBackendStatus(uri *url.URL, alive bool) {

}