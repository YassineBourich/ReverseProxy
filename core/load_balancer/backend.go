package load_balancer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL *url.URL `json:"url"`
	Alive bool `json:"alive"`
	CurrentConns int64 `json:"current_connections"`
	LastResponseTime time.Duration
	mux sync.RWMutex
}

// Auxilary backend to define custom UnmarshalJSON
type aux_backend struct {
	URL string `json:"url"`
	Alive bool `json:"alive"`
	CurrentConns int64 `json:"current_connections"`
}

// Method to define custom Unmarshaling approche
// We have to unmarshal a url into a string then parse it
func (b *Backend) UnmarshalJSON(data []byte) error {
	aux := aux_backend{}

	// Unmarshal the data to aux and check for error
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("Backend Unmarshal json error: %w\n", err)
	}

	// Parsing the url
	parsed_url, err := url.Parse(aux.URL)
	if err != nil {
		return fmt.Errorf("Backend parsing url error: %w\n", err)
	}

	// Assigning value to the original struct
	b.URL = parsed_url
	b.Alive = aux.Alive
	b.CurrentConns = aux.CurrentConns
	return nil
}

// Thread-safe counter incrementing
func (b *Backend) AddRequest() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.CurrentConns++
}

// Thread-safe counter decrementing
func (b *Backend) RequestDone() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.CurrentConns--
}

// Thread-safe status update
func (b *Backend) UpdateStatus(status bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = status
}

// Thread-safe response time update
func (b *Backend) UpdateResponseTime(responseTime time.Duration) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.LastResponseTime = responseTime
}