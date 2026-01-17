package load_balancer

import (
	"encoding/json"
	"net/url"
	"sync"
)

type Backend struct {
	URL *url.URL `json:"url"`
	Alive bool `json:"alive"`
	CurrentConns int64 `json:"current_connections"`
	mux sync.RWMutex
}

func (b *Backend) UnmarshalJSON(data []byte) error {
	aux_backend := struct {
		URL string `json:"url"`
		Alive bool `json:"alive"`
		CurrentConns int64 `json:"current_connections"`
	}{}

	if err := json.Unmarshal(data, &aux_backend); err != nil {
		return err
	}

	parsed_url, err := url.Parse(aux_backend.URL)
	if err != nil {
		return nil
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