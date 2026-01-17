package load_balancer

import (
	"encoding/json"
	"net/url"
	"os"
	"sync"
	"fmt"
)

type ServerPool struct {
	Backends []*Backend `json:"backends"`
	Current  uint64     `json:"current"` // Used for Round-Robin
	mux      sync.RWMutex
}

func (sp *ServerPool) LoadConfiguration() error {
	conf_file, err := os.ReadFile("config\\backends.json")
	if err != nil {
		return err
	}

	var backends []Backend

	err = json.Unmarshal(conf_file, &backends)
	fmt.Println(err)
	if err != nil {
		return err
	}
	
	fmt.Println(backends)
	//sp.Backends = make([]*Backend, len(backends))
	for i := range backends {
		sp.Backends = append(sp.Backends, &backends[i])
		//sp.Backends[i] = &backends[i]
	}
	fmt.Println(sp.Backends)
	return nil
}

func (sp *ServerPool) Print_backends() {
	if sp.Backends == nil {
		fmt.Println("nil backends")
	} else {
		fmt.Println(sp.Backends[0])
	}
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