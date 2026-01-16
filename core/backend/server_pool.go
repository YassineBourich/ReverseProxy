package backend

import (
	"sync"
	"net/url"
)

type ServerPool struct {
	Backends []*Backend `json:"backends"`
	Current  uint64     `json:"current"` // Used for Round-Robin
	mux      sync.RWMutex
}

func (sp *ServerPool) increment_current() {
	sp.mux.Lock()
	defer sp.mux.Unlock()
	sp.Current = (sp.Current + 1) % uint64(len(sp.Backends))
}

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

func (sp *ServerPool) GetNextValidPeer() *Backend {
	if sp.all_not_alive() {
		return nil
	}

	sp.increment_current()
	for !sp.Backends[sp.Current].Alive {
		sp.increment_current()
	}

	return sp.Backends[sp.Current]
}

func (sp *ServerPool) LeastConnValidPeer() *Backend {
	if sp.all_not_alive() {
		return nil
	}

	var least_conn_peer = sp.Backends[0]

	for _, b := range sp.Backends {
		if least_conn_peer.CurrentConns > b.CurrentConns {
			least_conn_peer = b
		}
	}

	return least_conn_peer
}

func (sp *ServerPool) AddBackend(backend *Backend) {
	sp.Backends = append(sp.Backends, backend)
}

func (sp *ServerPool) SetBackendStatus(uri *url.URL, alive bool) {

}