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

func (b *Backend) UnmarshalJSON(data []byte) error {
	aux_backend := struct {
		URL string `json:"url"`
		Alive bool `json:"alive"`
		CurrentConns int64 `json:"current_connections"`
	}{}

	if err := json.Unmarshal(data, &aux_backend); err != nil {
		return fmt.Errorf("Backend Unmarshal json error: %w\n", err)
	}

	parsed_url, err := url.Parse(aux_backend.URL)
	if err != nil {
		return fmt.Errorf("Backend parsing url error: %w\n", err)
	}

	b.URL = parsed_url
	b.Alive = aux_backend.Alive
	b.CurrentConns = aux_backend.CurrentConns
	return nil
}

func (b *Backend) AddRequest() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.CurrentConns++
}

func (b *Backend) RequestDone() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.CurrentConns--
}

func (b *Backend) UpdateStatus(status bool) {
	b.Alive = status
}

func (b *Backend) UpdateResponseTime(responseTime time.Duration) {
	b.LastResponseTime = responseTime
}