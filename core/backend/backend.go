package backend

import (
	"net/url"
	"sync"
)

type Backend struct {
	URL *url.URL `json:"url"`
	Alive bool `json:"alive"`
	CurrentConns int64 `json:"current_connections"`
	mux sync.RWMutex
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