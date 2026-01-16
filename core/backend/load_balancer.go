package backend

import (
	"net/url"
)

type LoadBalancer interface {
	GetNextValidPeer() *Backend
	LeastConnValidPeer() *Backend
	AddBackend(backend *Backend)
	SetBackendStatus(uri *url.URL, alive bool)
}